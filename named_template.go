package jsont

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// JsonNamedTemplate is a JSON template with named args
type JsonNamedTemplate interface {
	// String produces a JSON string from the template using the specified args
	//
	// If a named arg are missing from the supplied args:
	//
	// * if a default value for the named arg has been set - that value is used
	//
	// * if the template has not been made Strict, null is used
	//
	// * otherwise, an error is returned
	//
	// Each arg must be able to JSON Marshall
	String(args map[string]interface{}) (string, error)
	// Data produces a JSON []byte data from the template using the specified args
	//
	// If a named arg are missing from the supplied args:
	//
	// * if a default value for the named arg has been set - that value is used
	//
	// * if the template has not been made Strict, null is used
	//
	// * otherwise, an error is returned
	//
	// Each arg must be able to JSON Marshall
	Data(args map[string]interface{}) ([]byte, error)
	// ExpectedArgs returns a map of expected arg names - the boolean
	// value for each map entry indicates whether the template has a
	// default value for that named arg
	ExpectedArgs() map[string]bool
	// DefaultArgValue provides a default value for a specific named arg
	DefaultArgValue(argName string, value interface{}) JsonNamedTemplate
	// DefaultArgValues provides default values for the specified named args
	DefaultArgValues(defaults map[string]interface{}) JsonNamedTemplate
	// NewWith creates a new template with the args supplied being resolved in the new template
	//
	// Note: when resolving args into the new template, defaults are NOT used (but are copied over to the new)
	NewWith(args map[string]interface{}) (JsonNamedTemplate, error)
	// Strict makes the template strict - i.e. when generating String or Data, if any of the named args are not supplied an
	// error is returned
	Strict() JsonNamedTemplate
}

type jsonNamedTemplate struct {
	argNames         map[string]bool
	tokens           tokens
	fixedLens        int
	strict           bool
	defaultArgValues map[string]interface{}
	// used only during parsing...
	lastTokenStart int
}

// NewJsonNamedTemplate creates a new JSON template from a template string
//
// The template string can be any JSON with arg positions specified by '?'
//
// To escape a '?' in the template, use '??'
//
// Example:
//   jt, _ := NewJsonNamedTemplate(`{"foo":?foo,"bar":?bar,"baz":"??","qux":?qux}`)
//   println(jt.String(map[string]interface{}{"foo":"aaa", "bar":true, "qux":1.2}))
// would produce:
//   {"foo":"aaa","bar":true,"baz":"?","qux":1.2}
func NewJsonNamedTemplate(template string) (JsonNamedTemplate, error) {
	result := &jsonNamedTemplate{
		argNames:         map[string]bool{},
		tokens:           make([]jsonTemplateToken, 0),
		defaultArgValues: map[string]interface{}{},
	}
	if err := result.parse(template); err != nil {
		return nil, err
	}
	// test it...
	tArgs := map[string]interface{}{}
	for k := range result.argNames {
		tArgs[k] = nil
	}
	tData, _ := result.Data(tArgs)
	var v interface{}
	if err := json.Unmarshal(tData, &v); err != nil {
		return nil, err
	}
	return result, nil
}

func MustCompileJsonNamedTemplate(template string) JsonNamedTemplate {
	if jt, err := NewJsonNamedTemplate(template); err == nil {
		return jt
	} else {
		panic(err.Error())
	}
}

// Strict makes the template strict - i.e. when generating String or Data, if any of the named args are not supplied an
// error is returned
func (t *jsonNamedTemplate) Strict() JsonNamedTemplate {
	t.strict = true
	return t
}

// String produces a JSON string from the template using the specified args
//
// If a named arg are missing from the supplied args:
//
// * if a default value for the named arg has been set - that value is used
//
// * if the template has not been made Strict, null is used
//
// * otherwise, an error is returned
//
// Each arg must be able to JSON Marshall
func (t *jsonNamedTemplate) String(args map[string]interface{}) (string, error) {
	var builder strings.Builder
	builder.Grow(t.fixedLens)
	for _, tkn := range t.tokens {
		if tkn.fixed {
			builder.Write(tkn.fixedValue)
		} else if ad, err := t.getNamedArgValue(tkn.argName, args); err == nil {
			builder.Write(ad)
		} else {
			return "", err
		}
	}
	return builder.String(), nil
}

// Data produces a JSON []byte data from the template using the specified args
//
// If a named arg are missing from the supplied args:
//
// * if a default value for the named arg has been set - that value is used
//
// * if the template has not been made Strict, null is used
//
// * otherwise, an error is returned
//
// Each arg must be able to JSON Marshall
func (t *jsonNamedTemplate) Data(args map[string]interface{}) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.Grow(t.fixedLens)
	for _, tkn := range t.tokens {
		if tkn.fixed {
			buffer.Write(tkn.fixedValue)
		} else if ad, err := t.getNamedArgValue(tkn.argName, args); err == nil {
			buffer.Write(ad)
		} else {
			return nil, err
		}
	}
	return buffer.Bytes(), nil
}

// ExpectedArgs returns a map of expected arg names - the boolean
// value for each map entry indicates whether the template has a
// default value for that named arg
func (t jsonNamedTemplate) ExpectedArgs() map[string]bool {
	result := map[string]bool{}
	for k := range t.argNames {
		_, dv := t.defaultArgValues[k]
		result[k] = dv
	}
	return result
}

// NewWith creates a new template with the args supplied being resolved in the new template
//
// Note: when resolving args into the new template, defaults are NOT used (but are copied over to the new)
func (t *jsonNamedTemplate) NewWith(args map[string]interface{}) (JsonNamedTemplate, error) {
	result := &jsonNamedTemplate{
		argNames:         map[string]bool{},
		tokens:           tokens{},
		fixedLens:        t.fixedLens,
		strict:           t.strict,
		defaultArgValues: map[string]interface{}{},
	}
	for _, tkn := range t.tokens {
		if tkn.fixed {
			result.tokens = append(result.tokens, tkn)
		} else if v, ok := args[tkn.argName]; ok {
			if aData, err := argValueToData(v); err != nil {
				return nil, err
			} else {
				result.tokens = append(result.tokens, jsonTemplateToken{
					fixed:      true,
					fixedValue: aData,
				})
				result.fixedLens += len(aData)
			}
		} else {
			result.tokens = append(result.tokens, tkn)
			result.argNames[tkn.argName] = true
			if dv, ok := t.defaultArgValues[tkn.argName]; ok {
				result.defaultArgValues[tkn.argName] = dv
			}
		}
	}
	result.tokens = result.tokens.joinContiguousFixed()
	return result, nil
}

// DefaultArgValue provides a default value for a specific named arg
func (t *jsonNamedTemplate) DefaultArgValue(argName string, value interface{}) JsonNamedTemplate {
	t.defaultArgValues[argName] = value
	return t
}

// DefaultArgValues provides default values for the specified named args
func (t *jsonNamedTemplate) DefaultArgValues(defaults map[string]interface{}) JsonNamedTemplate {
	for k, v := range defaults {
		t.defaultArgValues[k] = v
	}
	return t
}

func (t *jsonNamedTemplate) getNamedArgValue(argName string, args map[string]interface{}) ([]byte, error) {
	if v, ok := args[argName]; ok {
		return argValueToData(v)
	} else if dv, dvok := t.defaultArgValues[argName]; dvok {
		return argValueToData(dv)
	} else if !ok && !t.strict {
		return argValueToData(v)
	}
	return nil, fmt.Errorf("expected named arg '%s'", argName)
}

func (t *jsonNamedTemplate) parse(template string) error {
	data := []byte(template)
	l := len(data)
	maxI := l - 1
	t.lastTokenStart = 0
	t.fixedLens = 0
	for i := 0; i < l; i++ {
		if data[i] == '?' {
			if i < maxI && data[i+1] == '?' {
				t.parseAddFixedToken(i+1, data)
				i++
				t.lastTokenStart = i + 1
			} else {
				if nameLen, err := t.parseAddArgToken(i, data); err == nil {
					i += nameLen
				} else {
					return err
				}
			}
		}
	}
	t.parseAddFixedToken(l, data)
	return nil
}

func (t *jsonNamedTemplate) parseAddFixedToken(i int, data []byte) {
	if i > t.lastTokenStart {
		t.tokens = append(t.tokens, jsonTemplateToken{
			fixed:      true,
			fixedValue: data[t.lastTokenStart:i],
		})
		t.fixedLens += i - t.lastTokenStart
	}
}

func (t *jsonNamedTemplate) parseAddArgToken(i int, data []byte) (int, error) {
	t.parseAddFixedToken(i, data)
	nameLen := scanForNameChars(i, data)
	if nameLen == 0 {
		return 0, fmt.Errorf("named token with no name at position %d", i)
	}
	argName := string(data[i+1 : i+1+nameLen])
	t.tokens = append(t.tokens, jsonTemplateToken{
		argName: argName,
	})
	t.argNames[argName] = true
	t.lastTokenStart = i + 1 + nameLen
	return nameLen, nil
}

func scanForNameChars(i int, data []byte) int {
	n := 0
	for j := i + 1; j < len(data); j++ {
		if isArgNameChar(data[j]) {
			n++
		} else {
			break
		}
	}
	return n
}

func isArgNameChar(b byte) bool {
	return b == '_' || b == '-' || (b >= '0' && b <= '9') ||
		(b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}
