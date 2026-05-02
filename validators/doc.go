// Package validators provides legacy validation helpers and plugin architecture.
//
// New code should usually prefer the root package builders, tags, and
// structured errors. This package remains a supported compatibility layer for
// older direct validators and as the plugin host for domain validators such as
// universal formats, email, UUID, and ULID. Some direct helpers intentionally
// preserve historical behavior that differs from the main compiler path; for example,
// StringValidators.OneOf compares case-insensitively while the main compiler's
// oneof rule compares exact strings.
package validators
