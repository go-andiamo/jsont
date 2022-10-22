package jsont

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Template is a JSON template with positional args
type Template interface {
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
	NewWith(args ...interface{}) (Template, error)
	Options(options ...Option) Template
}

type jsonTemplate struct {
	argsCount int
	tokens    tokens
	fixedLens int
	strict    bool
	checkReqd bool
	// used only during parsing...
	lastTokenStart int
}

// NewTemplate creates a new JSON template from a template string
//
// The template string can be any JSON with arg positions specified by '?'
//
// To escape a '?' in the template, use '??'
//
// Example:
//   jt, _ := NewTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
//   println(jt.String("aaa", "bbb", 1.2))
// would produce:
//   {"foo":"aaa","bar":"bbb","baz":"?","qux":1.2}
func NewTemplate(template string, options ...Option) (Template, error) {
	result := &jsonTemplate{
		tokens: make(tokens, 0),
		strict: true,
	}
	result.parse(template)
	if err := result.applyOptions(options, false); err != nil {
		return nil, err
	}

	if err := result.check(); err != nil {
		return nil, err
	}
	return result, nil
}

// MustCompileTemplate is the same as NewTemplate, except it panics if there is an error
func MustCompileTemplate(template string, options ...Option) Template {
	if jt, err := NewTemplate(template, options...); err == nil {
		return jt
	} else {
		panic(any(err))
	}
}

// Options applies the specified options to the template
//
// Note: unlike using options with NewTemplate and MustCompileTemplate, this method
// does not panic or error if any of the options are not applicable to this type
func (t *jsonTemplate) Options(options ...Option) Template {
	_ = t.applyOptions(options, true)
	return t
}

func (t *jsonTemplate) applyOptions(options []Option, ignoreErrs bool) error {
	for _, o := range options {
		if o != nil {
			if err := o.Apply(t); err != nil && !ignoreErrs {
				return err
			}
		}
	}
	return nil
}

// String produces a JSON string from the template using the specified args
//
// The number of args must match args specified in the original template (otherwise an error is returned)
//
// Each arg must be able to JSON Marshall
func (t *jsonTemplate) String(args ...interface{}) (string, error) {
	if err := t.checkArgs(args); err != nil {
		return "null", err
	}
	argsData, argsLen, err := t.getArgsData(args)
	if err != nil {
		return "null", err
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
	if err := t.checkArgs(args); err != nil {
		return nil, err
	}
	argsData, argsLen, err := t.getArgsData(args)
	if err != nil {
		return nil, err
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

func (t *jsonTemplate) getArgsData(args []interface{}) (argsData [][]byte, argsLen int, err error) {
	argsData = make([][]byte, t.argsCount)
	argsLen = 0
	l := len(args)
	for i := 0; i < l; i++ {
		if ad, e := argValueToData(args[i]); e == nil {
			argsData[i] = ad
			argsLen += len(ad)
		} else {
			err = e
			break
		}
	}
	for i := l; i < t.argsCount; i++ {
		argsData[i] = nullData
		argsLen += nullDataLen
	}
	return
}

// ExpectedArgs returns the expected number of args (that String() and Data() expects)
func (t *jsonTemplate) ExpectedArgs() int {
	return t.argsCount
}

func (t *jsonTemplate) checkArgs(args []interface{}) error {
	if t.strict && len(args) != t.argsCount {
		return fmt.Errorf("expected %d args but supplied %d args", t.argsCount, len(args))
	}
	return nil
}

// NewWith creates a new template with the args supplied being resolved in the new template
func (t *jsonTemplate) NewWith(args ...interface{}) (Template, error) {
	lArgs := len(args)
	if lArgs > t.argsCount {
		return nil, fmt.Errorf("too many args supplied (%d) - expected maximum of %d", lArgs, t.argsCount)
	}
	result := &jsonTemplate{
		argsCount: t.argsCount - len(args),
		tokens:    tokens{},
		fixedLens: t.fixedLens,
		strict:    t.strict,
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

func (t *jsonTemplate) check() (err error) {
	if t.checkReqd {
		tArgs := make([]interface{}, t.argsCount)
		tData, _ := t.Data(tArgs...)
		var v interface{}
		err = json.Unmarshal(tData, &v)
	}
	return
}
