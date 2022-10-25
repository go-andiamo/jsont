package jsont

import (
	"bytes"
	"encoding/json"
)

var nullData = []byte{'n', 'u', 'l', 'l'}

const nullDataLen = 4

type jsonTemplateToken struct {
	fixed      bool
	fixedValue []byte
	argName    string
}

type tokens []jsonTemplateToken

func (t tokens) joinContiguousFixed() tokens {
	l := len(t)
	result := make(tokens, 0)
	if l > 0 {
		var curr []byte
		for _, tkn := range t {
			if tkn.fixed {
				if curr != nil {
					curr = append(curr, tkn.fixedValue...)
				} else {
					curr = make([]byte, 0, len(tkn.fixedValue))
					curr = append(curr, tkn.fixedValue...)
				}
			} else {
				if curr != nil {
					result = append(result, jsonTemplateToken{fixed: true, fixedValue: curr})
				}
				result = append(result, tkn)
				curr = nil
			}
		}
		if curr != nil {
			result = append(result, jsonTemplateToken{fixed: true, fixedValue: curr})
		}
	}
	return result
}

func argValueToData(v interface{}) ([]byte, error) {
	switch vt := v.(type) {
	case nil:
		return nullData, nil
	case []byte:
		return vt, nil
	case json.RawMessage:
		return vt, nil
	case *NameValuePair:
		return vt.ToData()
	case *NameValuePairs:
		return vt.ToData()
	default:
		if jArg, err := json.Marshal(v); err == nil {
			return jArg, nil
		} else {
			return nil, err
		}
	}
}

type NameValuePair struct {
	name      string
	nameData  []byte
	value     interface{}
	omitEmpty bool
}

func NameValue(name string, value interface{}) *NameValuePair {
	return &NameValuePair{
		name:     name,
		nameData: []byte(`"` + name + `":`),
		value:    value,
	}
}

func (nvp *NameValuePair) OmitEmpty() *NameValuePair {
	nvp.omitEmpty = true
	return nvp
}

func (nvp *NameValuePair) ToData() (result []byte, err error) {
	useValue := nvp.value
	if gdfn, ok := useValue.(func(string) interface{}); ok {
		useValue = gdfn(nvp.name)
	}
	if nvp.omitEmpty && useValue == nil {
		return
	}
	vData, err := argValueToData(useValue)
	if err != nil {
		return nil, err
	}
	capacity := checkedCapacity(len(nvp.nameData), len(vData))
	result = make([]byte, 0, capacity)
	result = append(result, nvp.nameData...)
	result = append(result, vData...)
	return
}

func checkedCapacity(sz1, sz2 int) int {
	if tot := sz1 + sz2; tot < sz1 || tot < sz2 {
		return 0
	} else {
		return tot
	}
}

type NameValuePairs struct {
	pairs []*NameValuePair
}

func NameValues(pairs ...*NameValuePair) *NameValuePairs {
	return &NameValuePairs{
		pairs: pairs,
	}
}

func (nvps *NameValuePairs) ToData() ([]byte, error) {
	var buffer bytes.Buffer
	added := false
	for _, nvp := range nvps.pairs {
		if nvp != nil {
			if nvData, err := nvp.ToData(); err == nil && len(nvData) > 0 {
				if added {
					buffer.WriteByte(',')
				}
				buffer.Write(nvData)
				added = true
			} else if err != nil {
				return nil, err
			}
		}
	}
	return buffer.Bytes(), nil
}
