package handlers

import (
	"testing"
)

func TestProcessSQL(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		params   map[string]interface{}
		expected string
	}{
		{
			name:   "Param present logic",
			query:  "WHERE 1=1 AND {{status = :status else '1=1'}}",
			params: map[string]interface{}{"status": "active"},
			expected: "WHERE 1=1 AND status = \"active\"", // Note: simple string replacement adds quotes
		},
		{
			name:     "Param missing - use else",
			query:    "WHERE 1=1 AND {{status = :status else '1=0'}}",
			params:   map[string]interface{}{"other": "val"},
			expected: "WHERE 1=1 AND 1=0",
		},
		{
			name:     "Param missing - no else",
			query:    "WHERE {{status = :status}}",
			params:   map[string]interface{}{"other": "val"},
			expected: "WHERE ", // should return empty string if no else
		},
		{
			name:     "Direct replacement",
			query:    "SELECT * FROM users WHERE id = :id",
			params:   map[string]interface{}{"id": 123},
			expected: "SELECT * FROM users WHERE id = 123",
		},
		{
			name:     "Comments removal",
			query:    "{{/* x = :x */ else 'y'}}",
			params:   map[string]interface{}{"x": 1},
			expected: "x = 1",
		},
		{
			name: "Multiple blocks",
			query: "WHERE 1=1 AND {{a=:a else 'no_a'}} AND {{b=:b else 'no_b'}}",
			params: map[string]interface{}{"a": 10},
			expected: "WHERE 1=1 AND a=10 AND no_b",
		},
        {
            name: "String quoting",
            query: "name = :name",
            params: map[string]interface{}{"name": "John"},
            expected: "name = \"John\"",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProcessSQL(tt.query, tt.params)
			if got != tt.expected {
				t.Errorf("ProcessSQL() = %v, want %v", got, tt.expected)
			}
		})
	}
}
