package jsont

import "fmt"

type Option interface {
	Apply(on any) error
}

var (
	_OptionChecked         = &optionChecked{true}
	_OptionUnChecked       = &optionChecked{false}
	_OptionStrict          = &optionStrict{true}
	_OptionNonStrict       = &optionStrict{false}
	_OptionDefaultArgValue = func(name string, value interface{}) Option {
		return &optionDefaultArgValue{
			name:  name,
			value: value,
		}
	}
	_OptionDefaultArgValues = func(defaults map[string]interface{}) Option {
		return &optionDefaultArgValues{
			defaults: defaults,
		}
	}
)

var (
	OptionChecked          Option = _OptionChecked
	OptionUnChecked        Option = _OptionUnChecked
	OptionStrict           Option = _OptionStrict
	OptionNonStrict        Option = _OptionNonStrict
	OptionDefaultArgValue         = _OptionDefaultArgValue
	OptionDefaultArgValues        = _OptionDefaultArgValues
)

type optionChecked struct {
	checkReqd bool
}

func (o *optionChecked) Apply(on any) error {
	switch ont := on.(type) {
	case *jsonTemplate:
		ont.checkReqd = o.checkReqd
	case *jsonNamedTemplate:
		ont.checkReqd = o.checkReqd
	}
	return nil
}

type optionStrict struct {
	strict bool
}

func (o *optionStrict) Apply(on any) error {
	switch ont := on.(type) {
	case *jsonNamedTemplate:
		ont.strict = o.strict
		return nil
	case *jsonTemplate:
		ont.strict = o.strict
		return nil
	}
	return fmt.Errorf("option Strict cannot be applied to type '%T'", on)
}

type optionDefaultArgValue struct {
	name  string
	value interface{}
}

func (o *optionDefaultArgValue) Apply(on any) error {
	if ont, ok := on.(NamedTemplate); ok {
		ont.DefaultArgValue(o.name, o.value)
		return nil
	}
	return fmt.Errorf("option OptionDefaultArgValue cannot be applied to type '%T'", on)
}

type optionDefaultArgValues struct {
	defaults map[string]interface{}
}

func (o *optionDefaultArgValues) Apply(on any) error {
	if ont, ok := on.(NamedTemplate); ok {
		ont.DefaultArgValues(o.defaults)
		return nil
	}
	return fmt.Errorf("option OptionDefaultArgValues cannot be applied to type '%T'", on)
}
