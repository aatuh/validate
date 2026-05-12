package examples

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/aatuh/validate/v3"
)

// Test_apiProblemResponse demonstrates mapping validation errors to an
// RFC 7807-style response without adding HTTP framework behavior to validate.
func Test_apiProblemResponse(t *testing.T) {
	type SignupRequest struct {
		Email    string `json:"email" validate:"string;required;email"`
		Password string `json:"password" validate:"string;required;min=12"`
	}

	input := SignupRequest{
		Email:    "not-an-email",
		Password: "short",
	}

	v := validate.New()
	err := v.ValidateStructWithOpts(input, validate.ValidateOpts{
		FieldNameFunc: validate.JSONFieldName,
	})

	body, err := json.MarshalIndent(validationProblem(err), "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got := string(body)
	fmt.Println(got)

	const want = `{
  "type": "https://example.com/problems/validation",
  "title": "Invalid request body",
  "status": 422,
  "detail": "The request body failed validation.",
  "invalid-params": [
    {
      "name": "email",
      "code": "string.email.invalid"
    },
    {
      "name": "password",
      "code": "string.min"
    }
  ]
}`
	if got != want {
		t.Fatalf("problem response:\n%s", got)
	}

	// Output:
	// {
	//   "type": "https://example.com/problems/validation",
	//   "title": "Invalid request body",
	//   "status": 422,
	//   "detail": "The request body failed validation.",
	//   "invalid-params": [
	//     {
	//       "name": "email",
	//       "code": "string.email.invalid"
	//     },
	//     {
	//       "name": "password",
	//       "code": "string.min"
	//     }
	//   ]
	// }
}

type problemResponse struct {
	Type          string         `json:"type"`
	Title         string         `json:"title"`
	Status        int            `json:"status"`
	Detail        string         `json:"detail"`
	InvalidParams []invalidParam `json:"invalid-params,omitempty"`
}

type invalidParam struct {
	Name  string `json:"name"`
	Code  string `json:"code"`
	Param any    `json:"param,omitempty"`
}

func validationProblem(err error) problemResponse {
	problem := problemResponse{
		Type:   "https://example.com/problems/validation",
		Title:  "Invalid request body",
		Status: http.StatusUnprocessableEntity,
		Detail: "The request body failed validation.",
	}

	var es validate.Errors
	if !stderrors.As(err, &es) {
		return problem
	}

	problem.InvalidParams = make([]invalidParam, 0, len(es))
	for _, fe := range es {
		problem.InvalidParams = append(problem.InvalidParams, invalidParam{
			Name:  fe.Path,
			Code:  fe.Code,
			Param: fe.Param,
		})
	}

	sort.SliceStable(problem.InvalidParams, func(i, j int) bool {
		left := problem.InvalidParams[i]
		right := problem.InvalidParams[j]
		if left.Name == right.Name {
			return left.Code < right.Code
		}
		return left.Name < right.Name
	})

	return problem
}
