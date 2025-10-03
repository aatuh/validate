package structvalidator

import (
	"testing"

	"github.com/aatuh/validate/v3/core"
)

type benchItem struct {
	Name  string `validate:"string;min=3;max=20"`
	Price int    `validate:"int;min=0"`
}

type benchOrder struct {
	ID    string `validate:"string;min=8"`
	Lines []benchItem
}

func BenchmarkStruct_Medium_Aggregate(b *testing.B) {
	v := core.New()
	sv := NewStructValidator(v)
	in := benchOrder{
		ID: "ORDER001",
		Lines: []benchItem{
			{Name: "Alpha", Price: 10},
			{Name: "Bravo", Price: 20},
			{Name: "Charlie", Price: 30},
			{Name: "Delta", Price: 40},
			{Name: "Echo", Price: 50},
			{Name: "Foxtrot", Price: 60},
			{Name: "Golf", Price: 70},
			{Name: "Hotel", Price: 80},
			{Name: "India", Price: 90},
			{Name: "Juliet", Price: 100},
		},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sv.ValidateStruct(in)
	}
}

func BenchmarkStruct_Medium_StopOnFirst(b *testing.B) {
	v := core.New()
	sv := NewStructValidator(v)
	in := benchOrder{
		ID: "",
		Lines: []benchItem{
			{Name: "", Price: -1},
		},
	}
	opts := core.ValidateOpts{StopOnFirst: true}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sv.ValidateStructWithOpts(in, opts)
	}
}
