package mongo_test

import (
	"testing"

	"bradenkennedy.com/compass-cli/internal/mongo"
	"go.mongodb.org/mongo-driver/bson"
)

func TestAnalyzeSchemaDocs(t *testing.T) {
	// Sample docs
	docs := []bson.M{
		{
			"name":  "Alice",
			"age":   30,
			"alive": true,
			"address": bson.M{
				"city": "Seattle",
			},
			"tags": []interface{}{"admin", "developer"},
		},
		{
			"name": "Bob",
			"age":  25,
			"address": bson.M{
				"city": "Boston",
			},
		},
	}

	fields := mongo.AnalyzeSchemaDocs(docs)

	// We expect fields:
	// - name
	// - age
	// - alive
	// - address
	// - address.city
	// - tags
	// - tags[]

	expectedPaths := map[string]struct{}{
		"name":         {},
		"age":          {},
		"alive":        {},
		"address":      {},
		"address.city": {},
		"tags":         {},
		"tags[]":       {},
	}

	if len(fields) != len(expectedPaths) {
		t.Errorf("Expected %d fields, got %d", len(expectedPaths), len(fields))
	}

	fieldMap := make(map[string]mongo.SchemaField)
	for _, f := range fields {
		fieldMap[f.Path] = f
	}

	for path := range expectedPaths {
		sf, ok := fieldMap[path]
		if !ok {
			t.Errorf("Expected field path %q not found", path)
			continue
		}

		if sf.TotalDocCount != 2 {
			t.Errorf("Expected TotalDocCount of 2 for path %q, got %d", path, sf.TotalDocCount)
		}
	}

	// Verify specific counts
	nameField := fieldMap["name"]
	if count := nameField.Types["String"]; count != 2 {
		t.Errorf("Expected name to have 2 string types, got %d", count)
	}

	aliveField := fieldMap["alive"]
	if count := aliveField.Types["Boolean"]; count != 1 {
		t.Errorf("Expected alive to have 1 boolean type, got %d", count)
	}

	cityField := fieldMap["address.city"]
	if count := cityField.Types["String"]; count != 2 {
		t.Errorf("Expected address.city to have 2 string types, got %d", count)
	}
}
