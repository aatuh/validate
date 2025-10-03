// Package email provides email address validation as a plugin.
//
// The email package implements RFC-compliant email validation with support for
// basic email format validation, length limits, display name rejection (bare
// addresses only), local and domain part validation, and custom error messages
// and translations. The package registers itself as a plugin with the main
// validation system and provides comprehensive error handling with detailed
// error codes for different validation failure scenarios. It includes integration
// tests that verify end-to-end functionality through the main validation library.
package email
