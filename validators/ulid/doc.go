// Package ulid provides ULID validation as a plugin.
//
// The ulid package implements ULID (Universally Unique Lexicographically Sortable
// Identifier) format validation with support for Crockford's Base32 encoding
// validation, 26-character length requirement, character set validation (0-9, A-Z
// excluding I, L, O, U), and custom error messages and translations. The package
// registers itself as a plugin with the main validation system and provides strict
// format checking for ULID representations using Crockford's Base32 alphabet
// (0-9, A-Z excluding I, L, O, U for readability).
package ulid
