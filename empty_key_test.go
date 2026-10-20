package sjson

import (
	"encoding/json"
	"testing"
)

func TestEmptyKey(t *testing.T) {
	cases := []struct {
		name, input, path, set, delete string
		setError                       bool
	}{
		{"only", `{"":"old"}`, "", `{"":"new"}`, `{}`, false},
		{"first", `{"":"old","a":1}`, "", `{"":"new","a":1}`, `{"a":1}`, false},
		{"middle", `{"a":1,"":"old","b":2}`, "", `{"a":1,"":"new","b":2}`, `{"a":1,"b":2}`, false},
		{"last", `{"a":1,"":"old"}`, "", `{"a":1,"":"new"}`, `{"a":1}`, false},
		{"null", `{"":null}`, "", `{"":"new"}`, `{}`, false},
		{"grow", `{"":0}`, "", `{"":"new"}`, `{}`, false},
		{"shrink", `{"":"longer value","a":1}`, "", `{"":"new","a":1}`, `{"a":1}`, false},
		{"missing", `{"a":1}`, "", `{"a":1,"":"new"}`, `{"a":1}`, false},
		{"empty_object", `{}`, "", `{"":"new"}`, `{}`, false},
		{"empty_document", "", "", `{"":"new"}`, "", false},
		{"scalar", `123`, "", `{"":"new"}`, `123`, false},
		{"array", `[1,2]`, "", `[1,2]`, `[1,2]`, true},
		{"empty_array", `[]`, "", `[]`, `[]`, true},
		{"nested", `{"a":{"":"old","b":1}}`, "a.", `{"a":{"":"new","b":1}}`, `{"a":{"b":1}}`, false},
		{"nested_missing", `{}`, "a.", `{"a":{"":"new"}}`, `{}`, false},
		{"leading", `{"":{"a":"old"},"b":1}`, ".a", `{"":{"a":"new"},"b":1}`, `{"":{},"b":1}`, false},
		{"consecutive", `{}`, ".", `{"":{"":"new"}}`, `{}`, false},
		{"array_member", `[{"":"old"}]`, "0.", `[{"":"new"}]`, `[{}]`, false},
		{"nested_array", `{"a":[1,2]}`, "a.", `{"a":[1,2]}`, `{"a":[1,2]}`, true},
		{"forced", `{}`, ":", `{"":"new"}`, `{}`, false},
	}
	options := []struct {
		name string
		opts *Options
	}{
		{"default", nil},
		{"optimistic", &Options{Optimistic: true}},
		{"in_place", &Options{Optimistic: true, ReplaceInPlace: true}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, option := range options {
				t.Run(option.name, func(t *testing.T) {
					apis := []struct {
						name string
						call func() (string, error)
						del  bool
					}{
						{"Set", func() (string, error) {
							if option.opts == nil {
								return Set(tc.input, tc.path, "new")
							}
							return SetOptions(tc.input, tc.path, "new", option.opts)
						}, false},
						{"SetBytes", func() (string, error) {
							var result []byte
							var err error
							if option.opts == nil {
								result, err = SetBytes([]byte(tc.input), tc.path, "new")
							} else {
								result, err = SetBytesOptions([]byte(tc.input), tc.path, "new", option.opts)
							}
							return string(result), err
						}, false},
						{"SetRaw", func() (string, error) {
							if option.opts == nil {
								return SetRaw(tc.input, tc.path, `"new"`)
							}
							return SetRawOptions(tc.input, tc.path, `"new"`, option.opts)
						}, false},
						{"SetRawBytes", func() (string, error) {
							var result []byte
							var err error
							if option.opts == nil {
								result, err = SetRawBytes([]byte(tc.input), tc.path, []byte(`"new"`))
							} else {
								result, err = SetRawBytesOptions([]byte(tc.input), tc.path, []byte(`"new"`), option.opts)
							}
							return string(result), err
						}, false},
						{"Delete", func() (string, error) {
							if option.opts == nil {
								return Delete(tc.input, tc.path)
							}
							return DeleteOptions(tc.input, tc.path, option.opts)
						}, true},
						{"DeleteBytes", func() (string, error) {
							var result []byte
							var err error
							if option.opts == nil {
								result, err = DeleteBytes([]byte(tc.input), tc.path)
							} else {
								result, err = DeleteBytesOptions([]byte(tc.input), tc.path, option.opts)
							}
							return string(result), err
						}, true},
					}
					for _, api := range apis {
						t.Run(api.name, func(t *testing.T) {
							got, err := api.call()
							want := tc.set
							wantError := tc.setError
							if api.del {
								want, wantError = tc.delete, false
							}
							if (err != nil) != wantError {
								t.Fatalf("unexpected error: %v", err)
							}
							if got != want {
								t.Fatalf("expected %q, got %q", want, got)
							}
							if got != "" && !json.Valid([]byte(got)) {
								t.Fatalf("invalid JSON: %q", got)
							}
							if wantError && err.Error() != "cannot set array element for non-numeric key ''" {
								t.Fatalf("unexpected array key error: %v", err)
							}
						})
					}
				})
			}
		})
	}
}
