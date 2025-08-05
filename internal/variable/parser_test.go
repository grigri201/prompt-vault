package variable

import (
	"reflect"
	"testing"
)

func TestExtractVariables(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "empty content",
			content:  "",
			expected: []string{},
		},
		{
			name:     "no variables",
			content:  "Hello world!",
			expected: []string{},
		},
		{
			name:     "single variable",
			content:  "Hello {name}!",
			expected: []string{"name"},
		},
		{
			name:     "multiple unique variables",
			content:  "Hello {name}, your role is {role}",
			expected: []string{"name", "role"},
		},
		{
			name:     "duplicate variables",
			content:  "Hello {name}, your {role} is {role}",
			expected: []string{"name", "role"},
		},
		{
			name:     "variables with spaces and special characters",
			content:  "Hello {user_name}, your {team-lead} and {company.name}",
			expected: []string{"company.name", "team-lead", "user_name"},
		},
		{
			name:     "empty braces",
			content:  "Hello {}, your role is {role}",
			expected: []string{"role"},
		},
		{
			name:     "nested braces",
			content:  "Hello {{name}}, your role is {role}",
			expected: []string{"role", "{name"},
		},
		{
			name:     "variables at start and end",
			content:  "{greeting} world {punctuation}",
			expected: []string{"greeting", "punctuation"},
		},
		{
			name:     "multiple variables in complex text",
			content:  "Dear {customer_name},\n\nThank you for your {order_type} order #{order_id}.\nYour {product} will be delivered to {address} on {date}.\n\nBest regards,\n{company}",
			expected: []string{"address", "company", "customer_name", "date", "order_id", "order_type", "product"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ExtractVariables(tt.content)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ExtractVariables() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestReplaceVariables(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		content  string
		values   map[string]string
		expected string
	}{
		{
			name:     "empty content",
			content:  "",
			values:   map[string]string{"name": "Alice"},
			expected: "",
		},
		{
			name:     "nil values map",
			content:  "Hello {name}!",
			values:   nil,
			expected: "Hello {name}!",
		},
		{
			name:     "empty values map",
			content:  "Hello {name}!",
			values:   map[string]string{},
			expected: "Hello {name}!",
		},
		{
			name:     "no variables in content",
			content:  "Hello world!",
			values:   map[string]string{"name": "Alice"},
			expected: "Hello world!",
		},
		{
			name:     "single variable replacement",
			content:  "Hello {name}!",
			values:   map[string]string{"name": "Alice"},
			expected: "Hello Alice!",
		},
		{
			name:     "multiple variables replacement",
			content:  "Hello {name}, your role is {role}",
			values:   map[string]string{"name": "Alice", "role": "developer"},
			expected: "Hello Alice, your role is developer",
		},
		{
			name:     "duplicate variables replacement",
			content:  "Hello {name}, your {role} is {role}",
			values:   map[string]string{"name": "Alice", "role": "developer"},
			expected: "Hello Alice, your developer is developer",
		},
		{
			name:     "partial replacement - missing variable",
			content:  "Hello {name}, your role is {role}",
			values:   map[string]string{"name": "Alice"},
			expected: "Hello Alice, your role is {role}",
		},
		{
			name:     "extra values in map",
			content:  "Hello {name}!",
			values:   map[string]string{"name": "Alice", "role": "developer", "age": "30"},
			expected: "Hello Alice!",
		},
		{
			name:     "variables with special characters",
			content:  "Hello {user_name}, your {team-lead} and {company.name}",
			values:   map[string]string{"user_name": "Alice", "team-lead": "Bob", "company.name": "TechCorp"},
			expected: "Hello Alice, your Bob and TechCorp",
		},
		{
			name:     "empty braces not replaced",
			content:  "Hello {}, your role is {role}",
			values:   map[string]string{"": "empty", "role": "developer"},
			expected: "Hello {}, your role is developer",
		},
		{
			name:     "nested braces handling",
			content:  "Hello {{name}}, your role is {role}",
			values:   map[string]string{"{name": "Alice", "role": "developer"},
			expected: "Hello Alice}, your role is developer",
		},
		{
			name:     "variables at start and end",
			content:  "{greeting} world {punctuation}",
			values:   map[string]string{"greeting": "Hello", "punctuation": "!"},
			expected: "Hello world !",
		},
		{
			name:     "empty string replacement",
			content:  "Hello {name}, your role is {role}",
			values:   map[string]string{"name": "", "role": "developer"},
			expected: "Hello , your role is developer",
		},
		{
			name:     "whitespace in replacement values",
			content:  "Hello {name}, your role is {role}",
			values:   map[string]string{"name": "Alice Smith", "role": "senior developer"},
			expected: "Hello Alice Smith, your role is senior developer",
		},
		{
			name:     "newlines and complex content",
			content:  "Dear {customer_name},\n\nThank you for your {order_type} order #{order_id}.\nYour {product} will be delivered to {address} on {date}.\n\nBest regards,\n{company}",
			values: map[string]string{
				"customer_name": "John Doe",
				"order_type":    "premium",
				"order_id":      "12345",
				"product":       "laptop",
				"address":       "123 Main St",
				"date":          "2024-01-15",
				"company":       "TechStore Inc",
			},
			expected: "Dear John Doe,\n\nThank you for your premium order #12345.\nYour laptop will be delivered to 123 Main St on 2024-01-15.\n\nBest regards,\nTechStore Inc",
		},
		{
			name:     "unicode characters in variables and values",
			content:  "你好 {姓名}, 欢迎来到 {公司}",
			values:   map[string]string{"姓名": "张三", "公司": "科技公司"},
			expected: "你好 张三, 欢迎来到 科技公司",
		},
		{
			name:     "special characters in replacement values",
			content:  "Config: {setting1} and {setting2}",
			values:   map[string]string{"setting1": "key=value", "setting2": "{\"json\": true}"},
			expected: "Config: key=value and {\"json\": true}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.ReplaceVariables(tt.content, tt.values)
			if result != tt.expected {
				t.Errorf("ReplaceVariables() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestHasVariables(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "empty content",
			content:  "",
			expected: false,
		},
		{
			name:     "no variables",
			content:  "Hello world!",
			expected: false,
		},
		{
			name:     "single variable",
			content:  "Hello {name}!",
			expected: true,
		},
		{
			name:     "multiple variables",
			content:  "Hello {name}, your role is {role}",
			expected: true,
		},
		{
			name:     "duplicate variables",
			content:  "Hello {name}, your {role} is {role}",
			expected: true,
		},
		{
			name:     "variables with special characters",
			content:  "Hello {user_name}, your {team-lead} and {company.name}",
			expected: true,
		},
		{
			name:     "empty braces",
			content:  "Hello {}, your role is nice",
			expected: false,
		},
		{
			name:     "empty braces with valid variable",
			content:  "Hello {}, your role is {role}",
			expected: true,
		},
		{
			name:     "nested braces",
			content:  "Hello {{name}}, your role is nice",
			expected: true,
		},
		{
			name:     "variable at start",
			content:  "{greeting} world!",
			expected: true,
		},
		{
			name:     "variable at end",
			content:  "Hello world {punctuation}",
			expected: true,
		},
		{
			name:     "only braces without content",
			content:  "{}",
			expected: false,
		},
		{
			name:     "incomplete braces - only opening",
			content:  "Hello {name world",
			expected: false,
		},
		{
			name:     "incomplete braces - only closing",
			content:  "Hello name} world",
			expected: false,
		},
		{
			name:     "spaces in variable name",
			content:  "Hello {user name}!",
			expected: true,
		},
		{
			name:     "numbers in variable name",
			content:  "Item {item123} is available",
			expected: true,
		},
		{
			name:     "variable with underscores and dashes",
			content:  "User {user_id} from {team-name}",
			expected: true,
		},
		{
			name:     "complex text with variables",
			content:  "Dear {customer_name},\n\nThank you for your {order_type} order #{order_id}.\nYour {product} will be delivered to {address} on {date}.\n\nBest regards,\n{company}",
			expected: true,
		},
		{
			name:     "unicode variables",
			content:  "你好 {姓名}, 欢迎来到 {公司}",
			expected: true,
		},
		{
			name:     "text with braces but no variables",
			content:  "This is code: if (x > 0) { return true; }",
			expected: true, // The regex matches " return true; " as a variable
		},
		{
			name:     "escaped braces",
			content:  "This {variable} and this \\{not_variable\\}",
			expected: true,
		},
		{
			name:     "single character variable",
			content:  "Value is {x}",
			expected: true,
		},
		{
			name:     "multiple single braces",
			content:  "{ } { }",
			expected: true, // The regex matches " " (space) as a variable name
		},
		{
			name:     "whitespace only content",
			content:  "   \n\t  ",
			expected: false,
		},
		{
			name:     "variable with only whitespace",
			content:  "Hello {   }!",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parser.HasVariables(tt.content)
			if result != tt.expected {
				t.Errorf("HasVariables(%q) = %v, expected %v", tt.content, result, tt.expected)
			}
		})
	}
}