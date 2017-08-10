package interpolate_test

import (
	"testing"

	interpolate "."
)

type interpolateCase struct {
	Vars           map[string]interface{}
	String         string
	ExpectedString string
	ExpectedOk     error
}

var interpolateCases = []interpolateCase{
	interpolateCase{
		Vars:           map[string]interface{}{},
		String:         "Hi!",
		ExpectedString: "Hi!",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars:           nil,
		String:         "${woof}",
		ExpectedString: "",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars:           nil,
		String:         "woof ${woof}",
		ExpectedString: "woof ",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars:           map[string]interface{}{},
		String:         "${woof}",
		ExpectedString: "",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars:           map[string]interface{}{},
		String:         "$",
		ExpectedString: "$",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars:           map[string]interface{}{},
		String:         "$${woof}",
		ExpectedString: "$",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars:           map[string]interface{}{},
		String:         "$\\${woof}",
		ExpectedString: "$${woof}",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars: map[string]interface{}{
			"name": func() string {
				return "George"
			},
		},
		String:         "Hi, my name is ${name}.",
		ExpectedString: "Hi, my name is George.",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars: map[string]interface{}{
			"name": func() string {
				return "George"
			},
			"place": func() string {
				return "the restaurant"
			},
			"object": func() string {
				return "steak"
			},
		},
		String:         "Hi, my name is ${name}. Meet me at ${place} for ${object}.",
		ExpectedString: "Hi, my name is George. Meet me at the restaurant for steak.",
		ExpectedOk:     nil,
	},
	interpolateCase{
		Vars: map[string]interface{}{
			"name": func() string {
				return "George"
			},
		},
		String:         "Hi, my name is ${name",
		ExpectedString: "Hi, my name is ",
		ExpectedOk:     interpolate.Unterminated,
	},
	interpolateCase{
		Vars: map[string]interface{}{
			"": func() string {
				return "Nihil"
			},
		},
		String:         "The science of ${}.",
		ExpectedString: "The science of Nihil.",
		ExpectedOk:     nil,
	},
}

func TestInterpolate(t *testing.T) {
	interp := &interpolate.Interpolator{}
	for i, vals := range interpolateCases {
		interp.SetMap(vals.Vars)
		v, ok := interp.Interpolate(vals.String)
		if v != vals.ExpectedString || ok != vals.ExpectedOk {
			t.Errorf("Test %2d: Expected %q (%q), got %q (%q)", i, vals.ExpectedString, vals.ExpectedOk, v, ok)
		}
	}
}
