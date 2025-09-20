// Package errors provides error types and error handling utilities for validation.
//
// The errors package defines structured error types that provide detailed
// information about validation failures, including field paths, error codes,
// and human-readable messages.
//
// Key types:
// - FieldError: Represents a single validation failure at a specific field path
// - Errors: A collection of FieldError that implements the error interface
//
// The package also provides utility functions for error handling, filtering,
// and conversion to different formats (JSON, maps, etc.).
package errors
