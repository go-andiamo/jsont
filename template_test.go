package jsont

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTemplate(t *testing.T) {
	jt, err := NewTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 3, (jt.(*jsonTemplate)).argsCount)
	require.Equal(t, 3, jt.ExpectedArgs())
	str, err := jt.String("aaa", "bbb", 1.2)
	require.NoError(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":1.2}`, str)

	data, err := jt.Data("aaa", "bbb", 1.2)
	require.NoError(t, err)
	require.Equal(t, len(str), len(data))
}

func TestNewTemplateCreateErrors(t *testing.T) {
	_, err := NewTemplate(`{`, OptionChecked)
	require.Error(t, err)

	_, err = NewTemplate(`{?}`, OptionChecked)
	require.Error(t, err)
}

func TestMustCompileTemplate(t *testing.T) {
	jt := MustCompileTemplate(`{"foo":?}`)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonTemplate)).tokens))
}

func TestMustCompileTemplatePanics(t *testing.T) {
	const str = `{?}`
	_, err := NewTemplate(str, OptionChecked)
	require.Error(t, err)
	require.Panics(t, func() {
		MustCompileTemplate(str, OptionChecked)
	})
}

func TestTemplateErrors(t *testing.T) {
	jt, err := NewTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
	require.NoError(t, err)

	_, err = jt.Data()
	require.Error(t, err)
	require.Equal(t, "expected 3 args but supplied 0 args", err.Error())

	_, err = jt.Data("a", true, "b", "too many")
	require.Error(t, err)
	require.Equal(t, "expected 3 args but supplied 4 args", err.Error())

	_, err = jt.String()
	require.Error(t, err)
	require.Equal(t, "expected 3 args but supplied 0 args", err.Error())

	_, err = jt.String("a", true, "b", "too many")
	require.Error(t, err)
	require.Equal(t, "expected 3 args but supplied 4 args", err.Error())
}

func TestTemplateErrorsWithUnmarshallable(t *testing.T) {
	jt, err := NewTemplate(`{"foo":?}`)
	require.NoError(t, err)

	str, err := jt.String(nil)
	require.NoError(t, err)
	require.Equal(t, `{"foo":null}`, str)

	_, err = jt.String(func() {})
	require.Error(t, err)
	_, err = jt.Data(func() {})
	require.Error(t, err)
}

func TestTemplateNonStrict(t *testing.T) {
	jt, err := NewTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`, OptionNonStrict)
	require.NoError(t, err)

	str, err := jt.String(1.1)
	require.NoError(t, err)
	require.Equal(t, `{"foo":1.1,"bar":null,"baz":"?","qux":null}`, str)

	data, err := jt.Data(1.1)
	require.NoError(t, err)
	require.Equal(t, str, string(data[:]))
}

func TestTemplate_NewWith(t *testing.T) {
	org, err := NewTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
	require.NoError(t, err)
	require.NotNil(t, org)
	require.Equal(t, 7, len((org.(*jsonTemplate)).tokens))

	jt, err := org.NewWith("aaa", "bbb")
	require.NoError(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 3, len((jt.(*jsonTemplate)).tokens))
	require.Equal(t, 1, jt.ExpectedArgs())
	str, err := jt.String("ccc")
	require.NoError(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":"ccc"}`, str)

	jt, err = org.NewWith("aaa", "bbb", "ddd")
	require.NoError(t, err)
	require.NotNil(t, jt)
	require.Equal(t, 1, len((jt.(*jsonTemplate)).tokens))
	require.Equal(t, 0, jt.ExpectedArgs())
	str, err = jt.String()
	require.NoError(t, err)
	require.Equal(t, `{"foo":"aaa","bar":"bbb","baz":"?","qux":"ddd"}`, str)

	_, err = org.NewWith("aaa", "bbb", "ccc", "too many")
	require.Error(t, err)
	require.Equal(t, "too many args supplied (4) - expected maximum of 3", err.Error())

	_, err = org.NewWith(func() {})
	require.Error(t, err)
}

func TestTemplate_Options(t *testing.T) {
	jt, err := NewTemplate(`{"foo":?,"bar":?,"baz":"??","qux":?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)

	require.True(t, (jt.(*jsonTemplate)).strict)
	jt.Options(OptionNonStrict)
	require.False(t, (jt.(*jsonTemplate)).strict)
}
