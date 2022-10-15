package jsont

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// JsonTemplate is a JSON template with positional args
type JsonTemplate interface {
	// String produces a JSON string from the template using the specified args
	//
	// The number of args must match args specified in the original template (otherwise an error is returned)
	//
	// Each arg must be able to JSON Marshall
	String(args ...interface{}) (string, error)
	// Data produces a JSON []byte data from the template using the specified args
	//
	// The number of args must match args specified in the original template (otherwise an error is returned)
	//
	// Each arg must be able to JSON Marshall
	Data(args ...interface{}) ([]byte, error)
	// ExpectedArgs returns the expected number of args (that String() and Data() expects)
	ExpectedArgs() int
	// NewWith creates a new template with the args supplied being resolved in the new template
	NewWith(args ...interface{}) (JsonTemplate, error)
}

type jsonTemplate struct {
	argsCount int
	tokens    tokens
	fixedLens int
	// used only during parsing...
	lastTokenStart int
}

// NewJsonTemplate creates a new JSON template from a template string
//
// The template string can be any JSON with arg positions specified by '?'
//
// To escape a '?' in the template, use '??'
//
// Example:
//   jt, _ := NewJsonTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
//   println(jt.String("aaa", "bbb", 1.2))
// would produce:
//   {"foo":"aaa","bar":"bbb","baz":"?","qux":1.2}
func NewJsonTemplate(template string) (JsonTemplate, error) {
	result := &jsonTemplate{
		tokens: make(tokens, 0),
	}
	result.parse(template)
	// test it...
	tArgs := make([]interface{}, result.argsCount)
	tData, _ := result.Data(tArgs...)
	var v interface{}
	if err := json.Unmarshal(tData, &v); err != nil {
		return nil, err
	}
	return result, nil
}

// MustCompileJsonTemplate is the same as NewJsonTemplate, except it panics if there is an error
func MustCompileJsonTemplate(template string) JsonTemplate {
	if jt, err := NewJsonTemplate(template); err == nil {
		return jt
	} else {
		panic(err.Error())
	}
}

// String produces a JSON string from the template using the specified args
//
// The number of args must match args specified in the original template (otherwise an error is returned)
//
// Each arg must be able to JSON Marshall
func (t *jsonTemplate) String(args ...interface{}) (string, error) {
	if len(args) != t.argsCount {
		return "", fmt.Errorf("expected %d args but supplied %d args", t.argsCount, len(args))
	}
	argsData := make([][]byte, t.argsCount)
	argsLen := 0
	for i, v := range args {
		if ad, err := argValueToData(v); err == nil {
			argsData[i] = ad
			argsLen += len(ad)
		} else {
			return "", err
		}
	}
	var builder strings.Builder
	builder.Grow(t.fixedLens + argsLen)
	arg := 0
	for _, tkn := range t.tokens {
		if tkn.fixed {
			builder.Write(tkn.fixedValue)
		} else {
			builder.Write(argsData[arg])
			arg++
		}
	}
	return builder.String(), nil
}

// Data produces a JSON []byte data from the template using the specified args
//
// The number of args must match args specified in the original template (otherwise an error is returned)
//
// Each arg must be able to JSON Marshall
func (t *jsonTemplate) Data(args ...interface{}) ([]byte, error) {
	if len(args) != t.argsCount {
		return nil, fmt.Errorf("expected %d args but supplied %d args", t.argsCount, len(args))
	}
	argsData := make([][]byte, t.argsCount)
	argsLen := 0
	for i, v := range args {
		if ad, err := argValueToData(v); err == nil {
			argsData[i] = ad
			argsLen += len(ad)
		} else {
			return nil, err
		}
	}
	var buffer bytes.Buffer
	buffer.Grow(t.fixedLens + argsLen)
	arg := 0
	for _, tkn := range t.tokens {
		if tkn.fixed {
			buffer.Write(tkn.fixedValue)
		} else {
			buffer.Write(argsData[arg])
			arg++
		}
	}
	return buffer.Bytes(), nil
}

// ExpectedArgs returns the expected number of args (that String() and Data() expects)
func (t *jsonTemplate) ExpectedArgs() int {
	return t.argsCount
}

// NewWith creates a new template with the args supplied being resolved in the new template
func (t *jsonTemplate) NewWith(args ...interface{}) (JsonTemplate, error) {
	lArgs := len(args)
	if lArgs > t.argsCount {
		return nil, fmt.Errorf("too many args supplied (%d) - expected maximum of %d", lArgs, t.argsCount)
	}
	result := &jsonTemplate{
		argsCount: t.argsCount - len(args),
		tokens:    tokens{},
		fixedLens: t.fixedLens,
	}
	onArg := 0
	for _, tkn := range t.tokens {
		if tkn.fixed || onArg >= lArgs {
			result.tokens = append(result.tokens, tkn)
		} else if aData, err := argValueToData(args[onArg]); err != nil {
			return nil, err
		} else {
			result.tokens = append(result.tokens, jsonTemplateToken{
				fixed:      true,
				fixedValue: aData,
			})
			result.fixedLens += len(aData)
			onArg++
		}
	}
	result.tokens = result.tokens.joinContiguousFixed()
	return result, nil
}

func (t *jsonTemplate) parse(template string) {
	data := []byte(template)
	l := len(data)
	maxI := l - 1
	t.lastTokenStart = 0
	t.fixedLens = 0
	t.argsCount = 0
	for i := 0; i < l; i++ {
		if data[i] == '?' {
			if i < maxI && data[i+1] == '?' {
				t.parseAddFixedToken(i+1, data)
				i++
				t.lastTokenStart = i + 1
			} else {
				t.parseAddArgToken(i, data)
			}
		}
	}
	t.parseAddFixedToken(l, data)
	t.tokens = t.tokens.joinContiguousFixed()
}

func (t *jsonTemplate) parseAddFixedToken(i int, data []byte) {
	if i > t.lastTokenStart {
		t.tokens = append(t.tokens, jsonTemplateToken{
			fixed:      true,
			fixedValue: data[t.lastTokenStart:i],
		})
		t.fixedLens += i - t.lastTokenStart
	}
}

func (t *jsonTemplate) parseAddArgToken(i int, data []byte) {
	t.parseAddFixedToken(i, data)
	t.tokens = append(t.tokens, jsonTemplateToken{})
	t.lastTokenStart = i + 1
	t.argsCount++
}
