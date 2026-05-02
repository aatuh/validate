package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
	verrs "github.com/aatuh/validate/v3/errors"
)

func Test_customTypeValidator(t *testing.T) {
	v := validate.New().WithTypeValidator("slug", slugFactory{})

	fmt.Println("slice tag ok:", v.CheckTag("slice;foreach=(slug)", []string{"alpha"}) == nil)
	fmt.Println("map keys tag ok:", v.CheckTag("map;keys=(slug)", map[string]int{"beta": 1}) == nil)
	fmt.Println("map values tag ok:", v.CheckTag("map;values=(slug)", map[string]string{"id": "gamma"}) == nil)
	fmt.Println("invalid ok:", v.CheckTag("slice;foreach=(slug)", []string{"not slug"}) == nil)

	// Output:
	// slice tag ok: true
	// map keys tag ok: true
	// map values tag ok: true
	// invalid ok: false
}

type slugFactory struct{}

func (slugFactory) CreateValidator(validate.Translator) validate.TypeValidator {
	return slugValidator{}
}

type slugValidator struct{}

func (slugValidator) Validate(value any) error {
	s, ok := value.(string)
	if !ok || s == "" {
		return slugError()
	}
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' {
			return slugError()
		}
	}
	return nil
}

func slugError() error {
	return verrs.Errors{verrs.FieldError{Code: "slug.invalid", Msg: "invalid slug"}}
}
