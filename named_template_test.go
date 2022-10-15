package jsont

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNamedJsonTemplate(t *testing.T) {
	jt, err := NewJsonNamedTemplate(`{"foo":?foo,"bar":?bar,"baz":"??","qux":?qux}`)
	require.Nil(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonNamedTemplate)).argNames))
	str, err := jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb", "qux": 1.2})
	require.Nil(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":1.2}`, str)

	data, err := jt.Data(map[string]interface{}{"foo": "aaa", "bar": "bbb", "qux": 1.2})
	require.Nil(t, err)
	require.Equal(t, len(str), len(data))

	str, err = jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.Nil(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":null}`, str)

	jt.Strict()
	_, err = jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.NotNil(t, err)
	require.Equal(t, "expected named arg 'qux'", err.Error())
	_, err = jt.Data(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.NotNil(t, err)
	require.Equal(t, "expected named arg 'qux'", err.Error())

	jt.DefaultArgValue("qux", 2.2)
	expArgs := jt.ExpectedArgs()
	require.Equal(t, 3, len(expArgs))
	require.False(t, expArgs["foo"])
	require.False(t, expArgs["bar"])
	require.True(t, expArgs["qux"])
	str, err = jt.String(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.Nil(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":2.2}`, str)
	data, err = jt.Data(map[string]interface{}{"foo": "aaa", "bar": "bbb"})
	require.Nil(t, err)
	require.Equal(t, str, string(data[:]))

	jt.DefaultArgValues(map[string]interface{}{"foo": "xxx", "bar": "yyy", "qux": 3.3})
	str, err = jt.String(map[string]interface{}{})
	require.Nil(t, err)
	require.Equal(t, `{"foo":"xxx","bar":"yyy","baz":"?","qux":3.3}`, str)
	data, err = jt.Data(map[string]interface{}{})
	require.Nil(t, err)
	require.Equal(t, str, string(data[:]))
}

func TestNewNamedJsonTemplateCreateErrors(t *testing.T) {
	_, err := NewJsonNamedTemplate(`{`)
	require.NotNil(t, err)

	_, err = NewJsonNamedTemplate(`{?}`)
	require.NotNil(t, err)
}

func TestMustCompileJsonNamedTemplate(t *testing.T) {
	jt := MustCompileJsonNamedTemplate(`{"foo":?foo}`)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonNamedTemplate)).tokens))
}

func TestMustCompileJsonNamedTemplatePanics(t *testing.T) {
	const str = `{?foo}`
	_, err := NewJsonNamedTemplate(str)
	require.NotNil(t, err)
	require.Panics(t, func() {
		MustCompileJsonNamedTemplate(str)
	})
}

func TestJsonNamedTemplateParseFailsWithBardArgName(t *testing.T) {
	_, err := NewJsonNamedTemplate(`{?}`)
	require.NotNil(t, err)
	require.Equal(t, "named token with no name at position 1", err.Error())
}

func TestJsonNamedTemplate_NewWith(t *testing.T) {
	orig, err := NewJsonNamedTemplate(`{"foo":?foo,"bar":?bar,"baz":"??","qux":?qux}`)
	orig.DefaultArgValue("foo", "aaa")
	require.Nil(t, err)
	require.NotNil(t, orig)
	require.Equal(t, 8, len((orig.(*jsonNamedTemplate)).tokens))
	require.Equal(t, 3, len((orig.(*jsonNamedTemplate)).argNames))

	jt, err := orig.NewWith(map[string]interface{}{"qux": "ddd"})
	require.Nil(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 1, len((jt.(*jsonNamedTemplate)).defaultArgValues))
	require.Equal(t, 2, len((jt.(*jsonNamedTemplate)).argNames))
	require.Equal(t, 5, len((jt.(*jsonNamedTemplate)).tokens))
	str, err := jt.String(map[string]interface{}{})
	require.Nil(t, err)
	require.Equal(t, `{"foo":"aaa","bar":null,"baz":"?","qux":"ddd"}`, str)

	jt, err = orig.NewWith(map[string]interface{}{"foo": nil, "bar": nil, "qux": "ddd"})
	require.Nil(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 1, len((jt.(*jsonNamedTemplate)).tokens))
	require.Equal(t, 0, len((jt.(*jsonNamedTemplate)).argNames))
	require.Equal(t, len((jt.(*jsonNamedTemplate)).tokens[0].fixedValue), (jt.(*jsonNamedTemplate)).fixedLens)

	_, err = orig.NewWith(map[string]interface{}{"foo": func() {}})
	require.NotNil(t, err)
}
