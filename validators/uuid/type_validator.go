package uuid

import (
	verrs "github.com/aatuh/validate/v3/errors"
	"github.com/aatuh/validate/v3/translator"
	"github.com/aatuh/validate/v3/types"
)

// UUIDTypeValidator implements types.TypeValidator for UUID validation.
type UUIDTypeValidator struct {
	translator translator.Translator
}

// Validate validates a value as a UUID.
func (v *UUIDTypeValidator) Validate(value any) error {
	s, ok := value.(string)
	if !ok {
		msg := v.translateMessage("uuid.type", "expected string", nil)
		return verrs.Errors{verrs.FieldError{Path: "", Code: verrs.CodeStringType, Msg: msg}}
	}

	// Create a proper compiler instance for the existing validateUUIDString function
	compiler := types.NewCompiler(v.translator)
	if fe := validateUUIDString(compiler, s); fe.Code != "" {
		return verrs.Errors{fe}
	}
	return nil
}

// translateMessage returns a translated message if translator is available.
func (v *UUIDTypeValidator) translateMessage(code string, defaultMsg string, params []any) string {
	if v.translator != nil {
		if translated := v.translator.T(code, params...); translated != "" {
			return translated
		}
	}
	return defaultMsg
}

// UUIDTypeValidatorFactory creates UUID type validators.
type UUIDTypeValidatorFactory struct{}

// CreateValidator creates a new UUID type validator.
func (f *UUIDTypeValidatorFactory) CreateValidator(translator translator.Translator) types.TypeValidator {
	return &UUIDTypeValidator{translator: translator}
}

// RegisterUUIDType registers the UUID type in the global registry.
func RegisterUUIDType() {
	types.RegisterGlobalType("uuid", &UUIDTypeValidatorFactory{})
}
