package interpolate

import (
	"bytes"

	pandora ".."
)

// Error errors encountered during string interpolation.
type Error string

func (e Error) Error() string {
	return string(e)
}

// IllegalEscape Indicates that an illegal escape sequence was
// encountered in an interpolate command.
const IllegalEscape Error = Error("interpolate.IllegalEscape")

// Unterminated Indicates that an interpolate section wasn't
// terminated.
const Unterminated Error = Error("interpolate.Unterminated")

type interpolateState int

const (
	stateNone interpolateState = iota
	stateDelim
	stateEscape
	stateKey
)

var _ pandora.Interpolator = &Interpolator{}

// Interpolator takes a lookup map to find replacements for interpolations
// of the form `${key}`.
type Interpolator struct {
	lookup map[string]interface{}
}

// SetMap sets the lookup map for the interpolator.
func (interp *Interpolator) SetMap(lookup map[string]interface{}) error {
	interp.lookup = lookup
	return nil
}

func (interp *Interpolator) fetchValue(key string) (value string) {
	if interp.lookup == nil {
		return
	}
	val, ok := interp.lookup[key]
	if !ok {
		return
	}

	switch v := val.(type) {
	case int:
		value = string(v)
	case string:
		value = v
	case func() string:
		value = v()
	}
	return
}

// Interpolate replaces occurrences of `${key}` with the value of `key` in the
// lookup map.
func (interp *Interpolator) Interpolate(str string) (string, error) {
	var (
		b     bytes.Buffer
		k     bytes.Buffer
		state interpolateState
	)

	for _, c := range str {
		switch state {
		case stateKey:
			if c == '}' {
				state = stateNone
				key := k.String()
				k.Reset()
				b.WriteString(interp.fetchValue(key))
			} else {
				k.WriteRune(c)
			}
		case stateDelim:
			if c == '{' {
				state = stateKey
			} else if c == '$' {
				b.WriteRune('$')
			} else if c == '\\' {
				state = stateEscape
				b.WriteRune('$')
			} else {
				state = stateNone
				b.WriteRune('$')
				b.WriteRune(c)
			}
		case stateEscape:
			state = stateNone
			switch c {
			case '$':
				b.WriteRune('$')
			case 'n':
				b.WriteRune('\n')
			case 'r':
				b.WriteRune('\r')
			case 't':
				b.WriteRune('\t')
			case '\\':
				b.WriteRune('\\')
			default:
				b.WriteRune('\\')
				b.WriteRune(c)
			}
		default:
			switch c {
			case '$':
				state = stateDelim
			case '\\':
				state = stateEscape
			default:
				b.WriteRune(c)
			}
		}
	}

	if state == stateDelim {
		b.WriteRune('$')
	}

	if state == stateKey {
		return b.String(), Unterminated
	}

	return b.String(), nil
}
