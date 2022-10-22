package jsont

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNamedJsonTemplate(t *testing.T) {
	jt, err := NewNamedTemplate(`{"foo":?foo,"bar":?bar,"baz":"??","qux":?qux}`)
	require.NoError(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonNamedTemplate)).argNames))
	str, err := jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb", "qux": 1.2})
	require.NoError(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":1.2}`, str)

	data, err := jt.Data(map[string]interface{}{"foo": "aaa", "bar": "bbb", "qux": 1.2})
	require.NoError(t, err)
	require.Equal(t, len(str), len(data))

	jt.Options(OptionNonStrict)
	str, err = jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.NoError(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":null}`, str)

	jt.Options(OptionStrict)
	_, err = jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.Error(t, err)
	require.Equal(t, "expected named arg 'qux'", err.Error())
	_, err = jt.Data(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.Error(t, err)
	require.Equal(t, "expected named arg 'qux'", err.Error())

	jt.DefaultArgValue("qux", 2.2)
	expArgs := jt.ExpectedArgs()
	require.Equal(t, 3, len(expArgs))
	require.False(t, expArgs["foo"])
	require.False(t, expArgs["bar"])
	require.True(t, expArgs["qux"])
	str, err = jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.NoError(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":2.2}`, str)
	data, err = jt.Data(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.NoError(t, err)
	require.Equal(t, str, string(data[:]))

	jt.DefaultArgValues(map[string]interface{}{"foo": "xxx", "bar": "yyy", "qux": 3.3})
	str, err = jt.String(map[string]interface{}{})
	require.NoError(t, err)
	require.Equal(t, `{"foo":"xxx","bar":"yyy","baz":"?","qux":3.3}`, str)
	data, err = jt.Data(map[string]interface{}{})
	require.NoError(t, err)
	require.Equal(t, str, string(data[:]))
}

func TestNewNamedJsonTemplateCreateErrors(t *testing.T) {
	_, err := NewNamedTemplate(`{`)
	require.NoError(t, err)
	_, err = NewNamedTemplate(`{`, OptionChecked)
	require.Error(t, err)

	_, err = NewNamedTemplate(`{?}`)
	require.Error(t, err)
}

func TestMustCompileNamedTemplate(t *testing.T) {
	jt := MustCompileNamedTemplate(`{"foo":?foo}`)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonNamedTemplate)).tokens))
}

func TestMustCompileNamedTemplatePanics(t *testing.T) {
	const str = `{?foo}`
	_, err := NewNamedTemplate(str, OptionChecked)
	require.Error(t, err)
	require.Panics(t, func() {
		MustCompileNamedTemplate(str, OptionChecked)
	})
}

func TestNamedTemplateParseFailsWithBardArgName(t *testing.T) {
	_, err := NewNamedTemplate(`{?}`)
	require.Error(t, err)
	require.Equal(t, "named token with no nameData at position 1", err.Error())
}

func TestNamedTemplate_NewWith(t *testing.T) {
	orig, err := NewNamedTemplate(`{"foo":?foo,"bar":?bar,"baz":"??","qux":?qux}`)
	orig.DefaultArgValue("foo", "aaa")
	require.NoError(t, err)
	require.NotNil(t, orig)
	require.Equal(t, 8, len((orig.(*jsonNamedTemplate)).tokens))
	require.Equal(t, 3, len((orig.(*jsonNamedTemplate)).argNames))

	jt, err := orig.NewWith(map[string]interface{}{"qux": "ddd"})
	require.NoError(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 1, len((jt.(*jsonNamedTemplate)).defaultArgValues))
	require.Equal(t, 2, len((jt.(*jsonNamedTemplate)).argNames))
	require.Equal(t, 5, len((jt.(*jsonNamedTemplate)).tokens))
	jt.Options(OptionNonStrict)
	str, err := jt.String(map[string]interface{}{})
	require.NoError(t, err)
	require.Equal(t, `{"foo":"aaa","bar":null,"baz":"?","qux":"ddd"}`, str)

	jt, err = orig.NewWith(map[string]interface{}{"foo": nil, "bar": nil, "qux": "ddd"})
	require.NoError(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 1, len((jt.(*jsonNamedTemplate)).tokens))
	require.Equal(t, 0, len((jt.(*jsonNamedTemplate)).argNames))
	require.Equal(t, len((jt.(*jsonNamedTemplate)).tokens[0].fixedValue), (jt.(*jsonNamedTemplate)).fixedLens)

	_, err = orig.NewWith(map[string]interface{}{"foo": func() {}})
	require.Error(t, err)
}

type erroringOption struct{}

func (o *erroringOption) Apply(on any) error {
	return errors.New("Fooey")
}

func TestNewNamedTemplateWithErrorOption(t *testing.T) {
	_, err := NewNamedTemplate(`{"foo":?foo}`)
	require.NoError(t, err)
	_, err = NewNamedTemplate(`{"foo":?foo}`, &erroringOption{})
	require.Error(t, err)
	require.Equal(t, "Fooey", err.Error())
}
