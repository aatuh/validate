package examples

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

func Test_structOptions(t *testing.T) {
	type Account struct {
		Email    *string `json:"email" validate:"string;omitempty;email"`
		Password string  `json:"password" validate:"string;required;min=8"`
		Confirm  string  `json:"confirm" validate:"string;eqField=Password"`
		Token    string  `json:"token" validate:"string;requiredWith=Password"`
	}

	password := "long-password"
	account := Account{
		Email:    nil,
		Password: password,
		Confirm:  "mismatch",
	}

	v := validate.New()
	err := v.ValidateStructWithOpts(account, validate.ValidateOpts{
		FieldNameFunc: validate.JSONFieldName,
	})
	var es validate.Errors
	if errors.As(err, &es) {
		fmt.Println("has confirm:", es.Has("confirm"))
		fmt.Println("has token:", es.Has("token"))
	}

	// Output:
	// has confirm: true
	// has token: true
}
