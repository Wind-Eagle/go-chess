package uci

import (
	"fmt"
	"strconv"
	"strings"
)

type OptValue interface {
	TypeName() string

	optValueMarker()
	serialize() string
}

var (
	_ OptValue = OptValueBool(false)
	_ OptValue = OptValueInt(0)
	_ OptValue = OptValueString("")
	_ OptValue = OptValueButton{}
)

type OptValueBool bool

func (v OptValueBool) TypeName() string { return "bool" }
func (v OptValueBool) optValueMarker()  {}
func (v OptValueBool) serialize() string {
	if bool(v) {
		return "true"
	} else {
		return "false"
	}
}

type OptValueInt int64

func (v OptValueInt) TypeName() string  { return "int" }
func (v OptValueInt) optValueMarker()   {}
func (v OptValueInt) serialize() string { return strconv.FormatInt(int64(v), 10) }

type OptValueString string

func (v OptValueString) TypeName() string { return "string" }
func (v OptValueString) optValueMarker()  {}
func (v OptValueString) serialize() string {
	if len(v) == 0 {
		return "<empty>"
	}
	return string(v)
}

type OptValueButton struct{}

func (v OptValueButton) TypeName() string  { return "button" }
func (v OptValueButton) optValueMarker()   {}
func (v OptValueButton) serialize() string { panic("must not happen") }

type Option interface {
	Value() OptValue
	Clone() Option

	optionMarker()
	setValue(v OptValue, o coderOptions) error
}

var (
	_ Option = (*OptionCheck)(nil)
	_ Option = (*OptionSpin)(nil)
	_ Option = (*OptionCombo)(nil)
	_ Option = (*OptionButton)(nil)
	_ Option = (*OptionString)(nil)
)

type OptionCheck struct {
	val bool
}

func (o *OptionCheck) BoolValue() bool { return o.val }

func (o *OptionCheck) Value() OptValue { return OptValueBool(o.val) }
func (o *OptionCheck) Clone() Option   { return rawClone(o) }
func (o *OptionCheck) optionMarker()   {}
func (o *OptionCheck) setValue(v OptValue, _ coderOptions) error {
	if b, ok := v.(OptValueBool); ok {
		o.val = bool(b)
		return nil
	}
	return fmt.Errorf("bad option value type %v", v.TypeName())
}

type OptionSpin struct {
	val    int64
	minVal int64
	maxVal int64
}

func (o *OptionSpin) IntValue() int64 { return o.val }
func (o *OptionSpin) MinValue() int64 { return o.minVal }
func (o *OptionSpin) MaxValue() int64 { return o.maxVal }

func (o *OptionSpin) Value() OptValue { return OptValueInt(o.val) }
func (o *OptionSpin) Clone() Option   { return rawClone(o) }
func (o *OptionSpin) optionMarker()   {}
func (o *OptionSpin) setValue(v OptValue, _ coderOptions) error {
	if i, ok := v.(OptValueInt); ok {
		val := int64(i)
		if !(o.minVal <= val && val <= o.maxVal) {
			return fmt.Errorf("out of range: %v not in [%v; %v]", val, o.minVal, o.maxVal)
		}
		o.val = val
		return nil
	}
	return fmt.Errorf("bad option value type %v", v.TypeName())
}

type OptionCombo struct {
	val       string
	choices   []string
	choiceMap map[string]string
}

func (o *OptionCombo) StrValue() string    { return o.val }
func (o *OptionCombo) NumChoices() int     { return len(o.choices) }
func (o *OptionCombo) Choice(i int) string { return o.choices[i] }
func (o *OptionCombo) HasChoice(s string) bool {
	_, ok := o.choiceMap[caseFold(s)]
	return ok
}

func (o *OptionCombo) Value() OptValue { return OptValueString(o.val) }
func (o *OptionCombo) optionMarker()   {}

func (o *OptionCombo) Clone() Option {
	// No need to clone choices and choiceMap, as they don't change at all after creation.
	return rawClone(o)
}

func (o *OptionCombo) setValue(v OptValue, _ coderOptions) error {
	if s, ok := v.(OptValueString); ok {
		val, ok := o.choiceMap[caseFold(string(s))]
		if !ok {
			return fmt.Errorf("bad choice %q", string(s))
		}
		o.val = val
		return nil
	}
	return fmt.Errorf("bad option value type %v", v.TypeName())
}

type OptionButton struct{}

func (o *OptionButton) Value() OptValue { return OptValueButton{} }
func (o *OptionButton) Clone() Option   { return rawClone(o) }
func (o *OptionButton) optionMarker()   {}
func (o *OptionButton) setValue(v OptValue, _ coderOptions) error {
	if _, ok := v.(OptValueButton); ok {
		return nil
	}
	return fmt.Errorf("bad option value type %v", v.TypeName())
}

type OptionString struct {
	val string
}

func (o *OptionString) StrValue() string { return o.val }

func (o *OptionString) Value() OptValue { return OptValueString(o.val) }
func (o *OptionString) Clone() Option   { return rawClone(o) }
func (o *OptionString) optionMarker()   {}
func (o *OptionString) setValue(v OptValue, co coderOptions) error {
	if s, ok := v.(OptValueString); ok {
		val := string(s)
		if !isGoodString(val, co) || val == "<empty>" {
			return fmt.Errorf("bad option string %q", s)
		}
		if !co.AllowBadSubstringsInOptions &&
			(strings.Contains(val, "name") || strings.Contains(val, "value")) {
			return fmt.Errorf("option string %q contains forbidden substrings", s)
		}
		o.val = val
		return nil
	}
	return fmt.Errorf("bad option value type %v", v.TypeName())
}

type optPair struct {
	name  string
	value Option
}

var ponderOptName = caseFold("Ponder")
