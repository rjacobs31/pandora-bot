package main

import (
	"regexp"
	"testing"
)

type handlerRegexExistTest struct {
	regex *regexp.Regexp
	input string
	exist bool
}

var handlerRegexExistCases = []handlerRegexExistTest{
	{regex: &addressRegex, input: "pandora: ", exist: true},
	{regex: &addressRegex, input: "pan: ", exist: true},
	{regex: &addressRegex, input: "pineapples", exist: false},
	{regex: &addressRegex, input: "pineapples: ", exist: false},
}

func TestHandlerRegexExist(t *testing.T) {
	for i, test := range handlerRegexExistCases {
		match := test.regex.MatchString(test.input)
		if match != test.exist {
			if test.exist {
				t.Errorf("Test %2d: Expected /%s/ to match %q", i, test.regex.String(), test.input)
			} else {
				t.Errorf("Test %2d: Expected /%s/ not to match %q", i, test.regex.String(), test.input)
			}
		}
	}
}
