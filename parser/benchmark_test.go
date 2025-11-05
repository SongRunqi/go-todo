package parser

import (
	"testing"
)

var markdownSample = `# Update Task
- **Task ID:** 42
- **Task Name:** Sample Task
- **Status:** pending
- **User:** testuser
- **Due Date:** 2025-11-06
- **Urgency:** high
- **Created:** 2025-11-05 10:00:00
- **End Time:** 2025-11-06 18:00:00

## Description
This is a sample task description for benchmarking the markdown parser.
It has multiple lines to simulate real-world usage.
`

var jsonSample = `{
  "taskId": 42,
  "taskName": "Sample Task",
  "taskDesc": "This is a sample task description",
  "status": "pending",
  "user": "testuser",
  "createTime": "2025-11-05T10:00:00Z",
  "endTime": "2025-11-06T18:00:00Z",
  "dueDate": "2025-11-06",
  "urgent": "high"
}`

// BenchmarkParseMarkdown benchmarks markdown parsing
func BenchmarkParseMarkdown(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMarkdown(markdownSample)
	}
}

// BenchmarkParseJSON benchmarks JSON parsing
func BenchmarkParseJSON(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseJSON(jsonSample)
	}
}

// BenchmarkParse benchmarks auto-detection parsing
func BenchmarkParse(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"Markdown", markdownSample},
		{"JSON", jsonSample},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				Parse(tt.input)
			}
		})
	}
}

// BenchmarkLargeMarkdown benchmarks parsing large markdown
func BenchmarkLargeMarkdown(b *testing.B) {
	// Create a large markdown with many fields
	largeMarkdown := markdownSample
	for i := 0; i < 100; i++ {
		largeMarkdown += "\nAdditional line " + string(rune(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMarkdown(largeMarkdown)
	}
}

// BenchmarkSmallMarkdown benchmarks parsing minimal markdown
func BenchmarkSmallMarkdown(b *testing.B) {
	smallMarkdown := `# Task
- **Task ID:** 1
- **Task Name:** Simple Task
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseMarkdown(smallMarkdown)
	}
}
