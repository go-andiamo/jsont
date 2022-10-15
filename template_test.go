package jsont

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJsonTemplate(t *testing.T) {
	jt, err := NewJsonTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
	require.Nil(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 3, (jt.(*jsonTemplate)).argsCount)
	require.Equal(t, 3, jt.ExpectedArgs())
	str, err := jt.String("aaa", "bbb", 1.2)
	require.Nil(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":1.2}`, str)

	data, err := jt.Data("aaa", "bbb", 1.2)
	require.Nil(t, err)
	require.Equal(t, len(str), len(data))
}

func TestNewJsonTemplateCreateErrors(t *testing.T) {
	_, err := NewJsonTemplate(`{`)
	require.NotNil(t, err)

	_, err = NewJsonTemplate(`{?}`)
	require.NotNil(t, err)
}

func TestMustCompileJsonTemplate(t *testing.T) {
	jt := MustCompileJsonTemplate(`{"foo":?}`)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonTemplate)).tokens))
}

func TestMustCompileJsonTemplatePanics(t *testing.T) {
	const str = `{?}`
	_, err := NewJsonTemplate(str)
	require.NotNil(t, err)
	require.Panics(t, func() {
		MustCompileJsonTemplate(str)
	})
}

func TestJsonTemplateErrors(t *testing.T) {
	jt, err := NewJsonTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
	require.Nil(t, err)

	_, err = jt.Data()
	require.NotNil(t, err)
	require.Equal(t, "expected 3 args but supplied 0 args", err.Error())

	_, err = jt.Data("a", true, "b", "too many")
	require.NotNil(t, err)
	require.Equal(t, "expected 3 args but supplied 4 args", err.Error())

	_, err = jt.String()
	require.NotNil(t, err)
	require.Equal(t, "expected 3 args but supplied 0 args", err.Error())

	_, err = jt.String("a", true, "b", "too many")
	require.NotNil(t, err)
	require.Equal(t, "expected 3 args but supplied 4 args", err.Error())
}

func TestJsonTemplateErrorsWithUnmarshallable(t *testing.T) {
	jt, err := NewJsonTemplate(`{"foo":?}`)
	require.Nil(t, err)

	str, err := jt.String(nil)
	require.Nil(t, err)
	require.Equal(t, `{"foo":null}`, str)

	_, err = jt.String(func() {})
	require.NotNil(t, err)
	_, err = jt.Data(func() {})
	require.NotNil(t, err)
}

func TestJsonTemplate_NewWith(t *testing.T) {
	org, err := NewJsonTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
	require.Nil(t, err)
	require.NotNil(t, org)
	require.Equal(t, 7, len((org.(*jsonTemplate)).tokens))

	jt, err := org.NewWith("aaa", "bbb")
	require.Nil(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonTemplate)).tokens))
	require.Equal(t, 1, jt.ExpectedArgs())
	str, err := jt.String("ccc")
	require.Nil(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":"ccc"}`, str)

	jt, err = org.NewWith("aaa", "bbb", "ddd")
	require.Nil(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 1, len((jt.(*jsonTemplate)).tokens))
	require.Equal(t, 0, jt.ExpectedArgs())
	str, err = jt.String()
	require.Nil(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":"ddd"}`, str)

	_, err = org.NewWith("aaa", "bbb", "ccc", "too many")
	require.NotNil(t, err)
	require.Equal(t, "too many args supplied (4) - expected maximum of 3", err.Error())

	_, err = org.NewWith(func() {})
	require.NotNil(t, err)
}
