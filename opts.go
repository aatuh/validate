// Package validate - opts.go
// Runtime options for struct validation and field path formatting.
package validate

// ValidateOpts tunes validation behavior per call.
type ValidateOpts struct {
	// StopOnFirst makes validation fail fast on the first FieldError.
	// Default is false (aggregate all errors).
	StopOnFirst bool

	// PathSep sets the separator between nested field parts.
	// Example: "User.Addresses[2].Zip".
	// If empty, it will be taken from Validate (or "." if nil).
	PathSep string
}

// WithDefaults currently does not change anything, but is kept to allow
// future option defaults without breaking callers.
func (o ValidateOpts) WithDefaults() ValidateOpts {
	return o
}

// ApplyOpts fills missing values using the given *Validate instance.
// If PathSep is empty, it uses v's separator, or "." when v is nil.
func ApplyOpts(v *Validate, o ValidateOpts) ValidateOpts {
	o = o.WithDefaults()
	if o.PathSep == "" {
		if v != nil {
			o.PathSep = v.pathSep
		} else {
			o.PathSep = "."
		}
	}
	return o
}
