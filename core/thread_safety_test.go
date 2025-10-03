package core

import (
	"sync"
	"testing"

	"github.com/aatuh/validate/v3/translator"
)

func TestValidate_ThreadSafety(t *testing.T) {
	// Create a base validator
	base := New()

	// Test concurrent access to methods that return new instances
	var wg sync.WaitGroup
	results := make(chan *Validate, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Each goroutine creates its own validator instance
			tr := translator.NewSimpleTranslator(
				translator.DefaultEnglishTranslations(),
			)
			validator := base.WithTranslator(tr).PathSeparator("_")

			// Add a custom rule
			validator = validator.WithCustomRule("test", func(any) error { return nil })

			results <- validator
		}(i)
	}

	wg.Wait()
	close(results)

	// Collect results
	var validators []*Validate
	for v := range results {
		validators = append(validators, v)
	}

	// Verify we got 10 different validators
	if len(validators) != 10 {
		t.Fatalf("expected 10 validators, got %d", len(validators))
	}

	// Verify each validator has the expected configuration
	for i, v := range validators {
		if v.pathSep != "_" {
			t.Errorf("validator %d: expected pathSep '_', got '%s'", i, v.pathSep)
		}
		if v.translator == nil {
			t.Errorf("validator %d: expected translator, got nil", i)
		}
		if v.customRules["test"] == nil {
			t.Errorf("validator %d: expected custom rule 'test', got nil", i)
		}
	}
}

func TestValidate_Copy(t *testing.T) {
	tr := translator.NewSimpleTranslator(translator.DefaultEnglishTranslations())
	original := New().WithTranslator(tr).PathSeparator("_")

	// Add a custom rule
	original = original.WithCustomRule("test", func(any) error { return nil })

	// Create a copy
	copy := original.Copy()

	// Verify the copy has the same configuration
	if copy.pathSep != original.pathSep {
		t.Errorf("expected pathSep '%s', got '%s'", original.pathSep, copy.pathSep)
	}
	if copy.translator != original.translator {
		t.Errorf("expected same translator instance")
	}
	if copy.customRules["test"] == nil {
		t.Errorf("expected custom rule 'test' in copy")
	}

	// Verify they are different instances
	if copy == original {
		t.Errorf("copy should be a different instance")
	}
}

func TestValidate_Immutability(t *testing.T) {
	original := New()
	original.pathSep = "."
	original.translator = nil

	// Create a modified version
	tr := translator.NewSimpleTranslator(translator.DefaultEnglishTranslations())
	modified := original.WithTranslator(tr).PathSeparator("_")

	// Verify original is unchanged
	if original.pathSep != "." {
		t.Errorf("original pathSep should be unchanged, got '%s'", original.pathSep)
	}
	if original.translator != nil {
		t.Errorf("original translator should be unchanged")
	}

	// Verify modified has new values
	if modified.pathSep != "_" {
		t.Errorf("modified pathSep should be '_', got '%s'", modified.pathSep)
	}
	if modified.translator == nil {
		t.Errorf("modified should have translator")
	}
}
