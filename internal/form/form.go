package form

import (
	"errors"

	"github.com/homoarea/pigeon/internal/request"
	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
)

type TextInput = pterm.InteractiveTextInputPrinter
type Select = pterm.InteractiveSelectPrinter
type MultiSelect = pterm.InteractiveMultiselectPrinter

func NewInput() *TextInput {
	return pterm.DefaultInteractiveTextInput.WithMask("")
}
func NewMultiInput() *TextInput {
	return pterm.DefaultInteractiveTextInput.WithMultiLine()
}
func NewSelect() *Select {
	return &pterm.DefaultInteractiveSelect
}
func NewMultiSelect() *MultiSelect {
	return &pterm.DefaultInteractiveMultiselect
}

// Reqeust form

// For all request,name is not allowed be empty.
func RequestName(name *string) error {
	var err error
	*name, err = NewInput().WithOnValidateFunc(func(s string) error {
		if s == "" {
			return errors.New("name inputed is empty.")
		}
		return nil
	}).Show("Input request's name")
	if err != nil {
		return err
	}

	return nil
}

// pick the kind of protocols for request
func ReqeustProtocol(protocol *string) error {
	var err error
	*protocol, err = NewSelect().WithOptions(request.Protocols).Show("Pick a protocol")
	if err != nil {
		return err
	}
	return nil
}

func printlns(s ...string) {
	for _, v := range s {
		pterm.Println(pterm.LightGreen(v))
	}
}

// bind the v with a multiLineText
func multiLineStringText(v *string, desc ...string) error {
	printlns(desc...)

	s, err := NewMultiInput().
		WithDefaultValue(*v).
		Show("*string")
	if err != nil {
		return err
	}
	*v = s
	return nil

}

// v is a pointer
func validTomlValueText[T any](v *T, desc ...string) error {
	printlns(desc...)

	b, err := toml.Marshal(v)
	if err != nil {
		return err
	}
	text := NewMultiInput().
		WithDefaultValue(string(b)).
		WithOnValidateFunc(func(s string) error {
			return toml.Unmarshal([]byte(s), v)
		})
	if _, err := text.Show(pterm.Sprintf("%T", v)); err != nil {
		return err
	}

	return nil
}
