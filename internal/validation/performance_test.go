package validation

import (
	"fmt"
	"testing"

	"github.com/grigri201/prompt-vault/internal/models"
	"github.com/grigri201/prompt-vault/internal/parser"
)

func BenchmarkPromptValidation(b *testing.B) {
	meta := models.PromptMeta{
		Name:   "Benchmark Prompt",
		Author: "testuser",
		Tags:   []string{"test", "benchmark"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := meta.Validate()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkYAMLParsing(b *testing.B) {
	yamlContent := `---
name: "Benchmark Prompt"
author: "testuser"
tags: ["test", "benchmark"]
version: "1.0"
---
This is benchmark content with {variable}.`

	parser := createTestParser()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.ParsePromptFile(yamlContent)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIndexOperations(b *testing.B) {
	index := &models.Index{}
	entry := models.IndexEntry{
		GistID: "benchmark-test",
		Name:   "Benchmark Entry",
		Author: "testuser",
		Tags:   []string{"benchmark"},
	}

	b.Run("AddEntry", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			testIndex := &models.Index{}
			testIndex.AddImportedEntry(entry)
		}
	})

	// Prepare index with entries for search benchmark
	for i := 0; i < 1000; i++ {
		index.AddImportedEntry(models.IndexEntry{
			GistID: fmt.Sprintf("test-%d", i),
			Name:   fmt.Sprintf("Entry %d", i),
			Author: "testuser",
			Tags:   []string{"test"},
		})
	}

	b.Run("FindEntry", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = index.FindImportedEntry("test-500")
		}
	})

	b.Run("UpdateEntry", func(b *testing.B) {
		updateEntry := models.IndexEntry{
			GistID: "test-500",
			Name:   "Updated Entry",
			Author: "testuser",
			Tags:   []string{"updated"},
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			index.UpdateImportedEntry(updateEntry)
		}
	})
}

func BenchmarkPromptMetaOperations(b *testing.B) {
	b.Run("SetDefaultVersion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			meta := models.PromptMeta{
				Name:   "Test",
				Author: "testuser",
				Tags:   []string{"test"},
			}
			meta.SetDefaultVersion()
		}
	})

	b.Run("ValidateID", func(b *testing.B) {
		meta := models.PromptMeta{
			ID: "valid-test-id-123",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := meta.ValidateID()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkPromptToIndexEntry(b *testing.B) {
	prompt := models.Prompt{
		PromptMeta: models.PromptMeta{
			Name:        "Benchmark Prompt",
			Author:      "testuser",
			Tags:        []string{"benchmark", "test"},
			Version:     "1.0",
			Description: "Benchmark description",
		},
		GistID:  "benchmark123",
		GistURL: "https://gist.github.com/user/benchmark123",
		Content: "This is benchmark content with multiple variables {var1} and {var2}",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = prompt.ToIndexEntry()
	}
}

// Helper function to create a test parser
func createTestParser() *parser.YAMLParser {
	return parser.NewYAMLParser(parser.YAMLParserConfig{
		Strict: false, // non-strict mode for benchmark
	})
}
