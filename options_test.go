package jsont

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNewJsonTemplate_OptionChecked(t *testing.T) {
	_, err := NewTemplate(`{?}`, OptionChecked)
	require.Error(t, err)
	jt, err := NewTemplate(`{?}`, OptionUnChecked)
	require.NoError(t, err)
	require.False(t, (jt.(*jsonTemplate)).checkReqd)
}

func TestNewNamedTemplate_OptionChecked(t *testing.T) {
	_, err := NewNamedTemplate(`{?foo}`, OptionChecked)
	require.Error(t, err)
	jt, err := NewNamedTemplate(`{?foo}`, OptionUnChecked)
	require.NoError(t, err)
	require.False(t, (jt.(*jsonNamedTemplate)).checkReqd)
}

func TestNewNamedTemplate_OptionStrict(t *testing.T) {
	jt, err := NewNamedTemplate(`{"foo":?foo}`)
	require.NoError(t, err)
	require.True(t, (jt.(*jsonNamedTemplate)).strict)
	jt, err = NewNamedTemplate(`{"foo":?foo}`, OptionNonStrict)
	require.NoError(t, err)
	require.False(t, (jt.(*jsonNamedTemplate)).strict)
}

func TestNewJsonTemplate_OptionStrict(t *testing.T) {
	jt, err := NewTemplate(`{"foo":?}`)
	require.NoError(t, err)
	require.True(t, (jt.(*jsonTemplate)).strict)
	jt, err = NewTemplate(`{"foo":?}`, OptionNonStrict)
	require.NoError(t, err)
	require.False(t, (jt.(*jsonTemplate)).strict)
}

func TestOptionStrictErrors(t *testing.T) {
	err := OptionStrict.Apply(nil)
	require.Error(t, err)
}

func TestNewNamedTemplate_OptionDefaultArgValue(t *testing.T) {
	jt, err := NewNamedTemplate(`{"foo":?foo}`)
	require.NoError(t, err)
	require.Empty(t, (jt.(*jsonNamedTemplate)).defaultArgValues)
	jt, err = NewNamedTemplate(`{"foo":?foo}`, OptionDefaultArgValue("foo", nil))
	require.NoError(t, err)
	require.Equal(t, 1, len((jt.(*jsonNamedTemplate)).defaultArgValues))
	jt, err = NewNamedTemplate(`{"foo":?foo}`, OptionDefaultArgValue("foo", nil), OptionDefaultArgValue("bar", nil))
	require.NoError(t, err)
	require.Equal(t, 2, len((jt.(*jsonNamedTemplate)).defaultArgValues))
}

func TestNewJsonTemplate_OptionDefaultArgValue(t *testing.T) {
	_, err := NewTemplate(`{"foo":?}`)
	require.NoError(t, err)
	_, err = NewTemplate(`{"foo":?}`, OptionDefaultArgValue("foo", nil))
	require.Error(t, err)
}

func TestNewNamedTemplate_OptionDefaultArgValues(t *testing.T) {
	jt, err := NewNamedTemplate(`{"foo":?foo}`)
	require.NoError(t, err)
	require.Empty(t, (jt.(*jsonNamedTemplate)).defaultArgValues)
	jt, err = NewNamedTemplate(`{"foo":?foo}`, OptionDefaultArgValues(map[string]interface{}{"foo": nil}))
	require.NoError(t, err)
	require.Equal(t, 1, len((jt.(*jsonNamedTemplate)).defaultArgValues))
	jt, err = NewNamedTemplate(`{"foo":?foo}`, OptionDefaultArgValues(map[string]interface{}{"foo": nil, "bar": nil}))
	require.NoError(t, err)
	require.Equal(t, 2, len((jt.(*jsonNamedTemplate)).defaultArgValues))
}

func TestNewJsonTemplate_OptionDefaultArgValues(t *testing.T) {
	_, err := NewTemplate(`{"foo":?}`)
	require.NoError(t, err)
	_, err = NewTemplate(`{"foo":?}`, OptionDefaultArgValues(map[string]interface{}{"foo": nil}))
	require.Error(t, err)
}
