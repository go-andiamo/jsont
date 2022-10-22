package jsont

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestJoinContiguousFixed(t *testing.T) {
	orig := tokens{
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'a'},
		},
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'b'},
		},
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'c'},
		},
	}
	joined := orig.joinContiguousFixed()
	assert.Equal(t, 1, len(joined))
	assert.Equal(t, 3, len(joined[0].fixedValue))

	orig = tokens{
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'a'},
		},
		jsonTemplateToken{
			fixed: false,
		},
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'c'},
		},
	}
	joined = orig.joinContiguousFixed()
	assert.Equal(t, 3, len(joined))

	orig = tokens{
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'a'},
		},
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'b'},
		},
		jsonTemplateToken{
			fixed: false,
		},
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'c'},
		},
		jsonTemplateToken{
			fixed:      true,
			fixedValue: []byte{'d'},
		},
	}
	joined = orig.joinContiguousFixed()
	assert.Equal(t, 3, len(joined))
	assert.True(t, joined[0].fixed)
	assert.Equal(t, 2, len(joined[0].fixedValue))
	assert.False(t, joined[1].fixed)
	assert.True(t, joined[2].fixed)
	assert.Equal(t, 2, len(joined[2].fixedValue))
}

func TestArgValueToData(t *testing.T) {
	vdata, err := argValueToData("foo")
	assert.Nil(t, err)
	assert.Equal(t, 5, len(vdata))
	assert.Equal(t, `"foo"`, string(vdata[:]))

	vdata, err = argValueToData(true)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(vdata))
	assert.Equal(t, `true`, string(vdata[:]))

	vdata, err = argValueToData(1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(vdata))
	assert.Equal(t, `1`, string(vdata[:]))

	vdata, err = argValueToData(1.2)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(vdata))
	assert.Equal(t, `1.2`, string(vdata[:]))

	vdata, err = argValueToData(struct {
		Foo string
	}{"bar"})
	assert.Nil(t, err)
	assert.Equal(t, 13, len(vdata))
	assert.Equal(t, `{"Foo":"bar"}`, string(vdata[:]))

	origData := vdata
	vdata, err = argValueToData(origData)
	assert.Nil(t, err)
	assert.Equal(t, 13, len(vdata))
	assert.Equal(t, `{"Foo":"bar"}`, string(vdata[:]))
	assert.Equalf(t, vdata, origData, "must be the same")

	origRawData := json.RawMessage(vdata)
	vdata, err = argValueToData(origRawData)
	assert.Nil(t, err)
	assert.Equal(t, 13, len(vdata))
	assert.Equal(t, `{"Foo":"bar"}`, string(vdata[:]))
	assert.Equalf(t, vdata, origData, "must be the same")

	_, err = argValueToData(func() {})
	assert.NotNil(t, err)
}

func TestNameValue(t *testing.T) {
	jt, err := NewTemplate(`{?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)

	nv := NameValue("foo", true)

	str, err := jt.String(nv)
	require.NoError(t, err)
	require.Equal(t, `{"foo":true}`, str)

	nv = NameValue("foo", nil).OmitEmpty()
	str, err = jt.String(nv)
	require.NoError(t, err)
	require.Equal(t, `{}`, str)
}

func TestNameValueMarshallError(t *testing.T) {
	jt, err := NewTemplate(`{?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)

	nv := NameValue("foo", true)
	_, err = jt.String(nv)
	require.NoError(t, err)
	nv = NameValue("foo", func() {})
	_, err = jt.String(nv)
	require.Error(t, err)
}

func TestNameValueMarshallEmpty(t *testing.T) {
	jt, err := NewTemplate(`{?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)

	nv := NameValue("foo", nil).OmitEmpty()
	str, err := jt.String(nv)
	require.NoError(t, err)
	require.Equal(t, `{}`, str)
}

func TestNameValueFunc(t *testing.T) {
	jt, err := NewTemplate(`{?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)

	nv := NameValue("foo", func(name string) interface{} {
		return "foo"
	}).OmitEmpty()
	str, err := jt.String(nv)
	require.NoError(t, err)
	require.Equal(t, `{"foo":"foo"}`, str)

	cntr := &counter{}
	nv = NameValue("foo", cntr.getValue)
	str, err = jt.String(nv)
	require.NoError(t, err)
	require.Equal(t, `{"foo":1}`, str)
	str, err = jt.String(nv)
	require.NoError(t, err)
	require.Equal(t, `{"foo":2}`, str)
}

type counter struct {
	current int
}

func (c *counter) getValue(name string) interface{} {
	c.current++
	return c.current
}

func TestNameValues(t *testing.T) {
	jt, err := NewTemplate(`{?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)

	nvps := NameValues(NameValue("foo", nil), NameValue("bar", nil))
	str, err := jt.String(nvps)
	require.NoError(t, err)
	require.Equal(t, `{"foo":null,"bar":null}`, str)

	nvps = NameValues(NameValue("foo", nil).OmitEmpty(), NameValue("bar", nil).OmitEmpty())
	str, err = jt.String(nvps)
	require.NoError(t, err)
	require.Equal(t, `{}`, str)
}

func TestNameValuesMarshallError(t *testing.T) {
	jt, err := NewTemplate(`{?}`)
	require.NoError(t, err)
	require.NotNil(t, jt)

	nvps := NameValues(NameValue("foo", true))
	_, err = jt.String(nvps)
	require.NoError(t, err)
	nvps = NameValues(NameValue("foo", func() {}))
	_, err = jt.String(nvps)
	require.Error(t, err)
}
