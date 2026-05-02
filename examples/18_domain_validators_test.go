package examples

import (
	"fmt"
	"testing"

	"github.com/aatuh/validate/v3"
)

func Test_domainValidators(t *testing.T) {
	v := validate.New()

	checkSlug := v.String().Slug().Build()

	type Release struct {
		Version string `validate:"string;semver"`
		Phone   string `validate:"string;e164"`
	}

	fmt.Println("slug ok:", checkSlug("release-2026") == nil)
	fmt.Println("semver and e164 ok:", v.ValidateStruct(Release{
		Version: "1.2.3",
		Phone:   "+358401234567",
	}) == nil)
	fmt.Println("nested jwt ok:", v.CheckTag(
		"slice;foreach=(string;jwt)",
		[]string{"eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiIxMjMifQ.c2lnbmF0dXJl"},
	) == nil)
	fmt.Println("array ok:", v.CheckTag("array;len=2;foreach=(string;slug)", [2]string{"api", "docs"}) == nil)

	// Output:
	// slug ok: true
	// semver and e164 ok: true
	// nested jwt ok: true
	// array ok: true
}
