package handler_test

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestContractRejectsNULInItemStrings(t *testing.T) {
	contract, err := openapi3.NewLoader().LoadFromFile("../../../../api/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, model := range []string{"ItemInput", "Item"} {
		for _, field := range []string{"name", "description"} {
			schema := contract.Components.Schemas[model].Value.Properties[field].Value
			for _, tc := range []struct {
				name, value string
				valid       bool
			}{
				{"ordinary unicode", "名前", true},
				{"empty", "", field == "description"},
				{"NUL at start", "\x00text", false},
				{"NUL in middle", "a\x00b", false},
				{"NUL at end", "text\x00", false},
				{"NUL before newline", "text\x00\n", false},
				{"literal escape", `\u0000`, true},
				{"newline", "first\nsecond", true},
			} {
				t.Run(model+"/"+field+"/"+tc.name, func(t *testing.T) {
					err := schema.VisitJSON(tc.value)
					if (err == nil) != tc.valid {
						t.Fatalf("value %q: validation = %v, want valid=%v", tc.value, err, tc.valid)
					}
				})
			}
		}
	}
}
