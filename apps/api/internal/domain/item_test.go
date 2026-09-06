package domain

import (
	"strings"
	"testing"
)

func TestItemInputValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		input  ItemInput
		fields []string
	}{
		{"valid unicode", ItemInput{Name: strings.Repeat("名", 100), Description: strings.Repeat("文", 2000)}, nil},
		{"empty name", ItemInput{}, []string{"name"}},
		{"both too long", ItemInput{Name: strings.Repeat("名", 101), Description: strings.Repeat("文", 2001)}, []string{"name", "description"}},
		{"NUL in name", ItemInput{Name: "a\x00b"}, []string{"name"}},
		{"NUL in description", ItemInput{Name: "valid", Description: "\x00"}, []string{"description"}},
		{"NUL in both fields", ItemInput{Name: "\x00name", Description: "details\x00"}, []string{"name", "description"}},
		{"length and NUL errors", ItemInput{Name: "", Description: "\x00"}, []string{"name", "description"}},
		{"literal escape and newline", ItemInput{Name: `\u0000`, Description: "first\nsecond"}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			errs := tc.input.Validate()
			if len(errs) != len(tc.fields) {
				t.Fatalf("errors = %v, want fields %v", errs, tc.fields)
			}
			for i, field := range tc.fields {
				if errs[i].Field != field || errs[i].Message == "" {
					t.Fatalf("error = %+v", errs[i])
				}
			}
		})
	}
}
