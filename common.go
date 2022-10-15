package jsont

import "encoding/json"

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
	case []byte:
		return vt, nil
	case json.RawMessage:
		return vt, nil
	default:
		if jArg, err := json.Marshal(v); err == nil {
			return jArg, nil
		} else {
			return nil, err
		}
	}
}
