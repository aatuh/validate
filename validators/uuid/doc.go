// Package uuid provides UUID validation as a plugin.
//
// The uuid package implements canonical UUID format validation with support for
// standard UUID format validation, hyphen placement verification, hexadecimal
// character validation, and custom error messages and translations. The package
// registers itself as a plugin with the main validation system and provides
// strict format checking for canonical UUID representations using the standard
// format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx where each 'x' is a hexadecimal
// digit (0-9, a-f, A-F).
package uuid
