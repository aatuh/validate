package errors

import (
	"encoding/json"
	stderr "errors"
	"testing"
)

func TestFieldError_String_WithAndWithoutMsg(t *testing.T) {
	e1 := FieldError{Path: "User.Name", Code: CodeStringMin, Param: 3}
	if got := e1.String(); got == "" || !contains(got, "User.Name") || !contains(got, CodeStringMin) {
		t.Fatalf("unexpected: %q", got)
	}
	e2 := FieldError{Path: "Age", Code: CodeNumberMin, Param: 18, Msg: "must be at least 18"}
	got := e2.String()
	if !contains(got, "Age") || !contains(got, "must be at least 18") {
		t.Fatalf("unexpected: %q", got)
	}
}

func TestErrors_Error_Has_Filter_AsMap_Sort(t *testing.T) {
	es := Errors{
		{Path: "User.Website", Code: CodeStringPattern, Msg: "bad"},
		{Path: "User.Name", Code: CodeStringMin, Param: 2},
		{Path: "Order.ID", Code: CodeStringNonEmpty},
	}
	if !es.Has("User.Website") || es.Has("User.Missing") {
		t.Fatalf("Has path failed")
	}
	sub := es.Filter("User.")
	if len(sub) != 2 {
		t.Fatalf("Filter size = %d", len(sub))
	}
	m := es.AsMap()
	if len(m["User.Website"]) != 1 || len(m["User.Name"]) != 1 {
		t.Fatalf("AsMap mismatch: %#v", m)
	}
	// Sorting should order by Path then Code.
	es.Sort()
	if es[0].Path != "Order.ID" {
		t.Fatalf("sort expected Order.ID first, got %s", es[0].Path)
	}
}

func TestErrors_ErrorJoin_Unwrap_JSON(t *testing.T) {
	e1 := Errors{{Path: "A", Code: CodeUnknown, Msg: "a"}}
	e2 := stderr.New("plain")
	joined := Join(e1, e2, nil)
	if got := joined.Error(); !contains(got, "A") || !contains(got, "plain") {
		t.Fatalf("join message: %q", got)
	}
	// Unwrap returns nil by design.
	if joined.Unwrap() != nil {
		t.Fatalf("unwrap must be nil")
	}
	// JSON round-trip using encoding/json.
	b, err := joined.MarshalJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var tmp []FieldError
	if err := json.Unmarshal(b, &tmp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	back := Errors(tmp)
	// Compare on essential fields (Path/Code).
	if !sameCore(joined, back) {
		t.Fatalf("round-trip mismatch: %#v vs %#v", joined, back)
	}
}

func sameCore(a, b Errors) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Path != b[i].Path || a[i].Code != b[i].Code {
			return false
		}
	}
	return true
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
