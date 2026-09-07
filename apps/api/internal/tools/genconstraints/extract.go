package main

import (
	"maps"
	"slices"

	"github.com/getkin/kin-openapi/openapi3"
)

type fieldConstraint struct {
	Schema    string
	Field     string
	MinLength *int
	MaxLength *int
	Pattern   string
}

func extract(doc *openapi3.T) []fieldConstraint {
	if doc == nil || doc.Components == nil {
		return nil
	}
	var out []fieldConstraint
	for _, schemaName := range slices.Sorted(maps.Keys(doc.Components.Schemas)) {
		ref := doc.Components.Schemas[schemaName]
		if ref == nil || ref.Value == nil {
			continue
		}
		for _, fieldName := range slices.Sorted(maps.Keys(ref.Value.Properties)) {
			prop := ref.Value.Properties[fieldName]
			if prop == nil || prop.Value == nil {
				continue
			}
			c, ok := constraintFromSchema(schemaName, fieldName, prop.Value)
			if ok {
				out = append(out, c)
			}
		}
	}
	return out
}

func constraintFromSchema(schemaName, fieldName string, schema *openapi3.Schema) (fieldConstraint, bool) {
	if schema.Type != nil && !schema.Type.Is(openapi3.TypeString) {
		return fieldConstraint{}, false
	}
	c := fieldConstraint{Schema: schemaName, Field: fieldName}
	if schema.MinLength > 0 {
		n := int(schema.MinLength)
		c.MinLength = &n
	}
	if schema.MaxLength != nil {
		n := int(*schema.MaxLength)
		c.MaxLength = &n
	}
	c.Pattern = schema.Pattern
	if c.MinLength == nil && c.MaxLength == nil && c.Pattern == "" {
		return fieldConstraint{}, false
	}
	return c, true
}
