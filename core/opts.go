package core

// ValidateOpts tunes validation behavior per call.
type ValidateOpts struct {
	StopOnFirst bool
	PathSep     string
}

// WithDefaults keeps the door open for future defaults.
func (o ValidateOpts) WithDefaults() ValidateOpts { return o }

// ApplyOpts fills missing values using the given *Validate instance.
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
