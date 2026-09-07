package main

import (
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestExtractStringConstraints(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromFile(filepath.Join("testdata", "spec.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	got := extract(doc)
	want := []fieldConstraint{
		{Schema: "WidgetInput", Field: "note", MaxLength: intPtr(500)},
		{Schema: "WidgetInput", Field: "title", MinLength: intPtr(2), MaxLength: intPtr(40), Pattern: `^[^\x00]*$`},
	}
	if len(got) != len(want) {
		t.Fatalf("extract() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i].Schema != want[i].Schema || got[i].Field != want[i].Field || got[i].Pattern != want[i].Pattern {
			t.Fatalf("extract()[%d] = %#v, want %#v", i, got[i], want[i])
		}
		if !intPtrEqual(got[i].MinLength, want[i].MinLength) || !intPtrEqual(got[i].MaxLength, want[i].MaxLength) {
			t.Fatalf("extract()[%d] lengths = min=%v max=%v, want min=%v max=%v", i, got[i].MinLength, got[i].MaxLength, want[i].MinLength, want[i].MaxLength)
		}
	}
}

func TestExtractFromRepositoryContract(t *testing.T) {
	doc, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "..", "api", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	got := extract(doc)
	var name, description *fieldConstraint
	for i := range got {
		if got[i].Schema == "ItemInput" && got[i].Field == "name" {
			name = &got[i]
		}
		if got[i].Schema == "ItemInput" && got[i].Field == "description" {
			description = &got[i]
		}
	}
	if name == nil || name.MinLength == nil || *name.MinLength != 1 || name.MaxLength == nil || *name.MaxLength != 100 || name.Pattern != `^[^\x00]*$` {
		t.Fatalf("ItemInput.name = %#v", name)
	}
	if description == nil || description.MinLength != nil || description.MaxLength == nil || *description.MaxLength != 2000 || description.Pattern != `^[^\x00]*$` {
		t.Fatalf("ItemInput.description = %#v", description)
	}
}

func intPtr(v int) *int { return &v }

func intPtrEqual(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}
