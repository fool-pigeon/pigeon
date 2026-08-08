package form

import (
	"errors"
	"slices"

	"atomicgo.dev/keyboard/keys"
	"github.com/homoarea/pigeon/internal/request"
	"github.com/homoarea/pigeon/internal/request/httpx"
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

// help description of field forms
const (
	exampleInputMap = `
for example:
	name1="v1"
	name2='2'
`
	exampleInputMapValues = `
for example:
	k1=['v1','v2']
	k2=["v1"]
`
	exampleInputMultiFormFields = `
example:
	[[field]]
	name='name1'
	values=['v1','v2']
	content_type=''
	file_name=''
	file_path=''
`
	helpURL         = `Input URL`
	helpMethod      = `Select Method`
	helpPathParams  = `Input PathParams` + exampleInputMap
	helpQueryParams = `Input QueryParams` + exampleInputMapValues
	helpHeaders     = `Input Headers` + exampleInputMapValues
	helpDesc        = `Input Description of Request`
	helpSettings    = `Input settings`

	helpBodyType           = `Select body type`
	helpBodyJSON           = `Input JSON`
	helpBodyXML            = `Input XML`
	helpBodyRaw            = `Input Raw`
	helpBodyBinary         = `Input the file path`
	helpBodyMultiForm      = `Input MultiFormFields` + exampleInputMultiFormFields
	helpBodyFormUrlEncoded = `Input Form URL Encoded` + exampleInputMapValues
)

// For httpx.Request
func HttpForm(r *httpx.Request, defaultOptions ...string) error {
	actions := map[string]func(r *httpx.Request) error{
		httpx.FieldURL:         HttpURL,
		httpx.FieldMethod:      HttpMethod,
		httpx.FieldPathParams:  HttpPathParams,
		httpx.FieldQueryParams: HttpQueryParams,
		httpx.FieldHeaders:     HttpHeaders,
		httpx.FieldBody:        HttpBody,
		httpx.FieldDesc:        HttpDesc,
		httpx.FieldSettings:    HttpSetting,
	}
	nextActions, err := NewMultiSelect().
		WithOptions(httpx.Fields).
		WithDefaultOptions(defaultOptions).
		WithMaxHeight(8).
		WithFilter(false).
		WithKeySelect(keys.Space).
		Show("What's next?")
	if err != nil {
		return err
	}
	for _, k := range httpx.Fields {
		if !slices.Contains(nextActions, k) {
			continue
		}
		if err := actions[k](r); err != nil {
			return err
		}
	}
	return nil
}
func HttpURL(r *httpx.Request) error {
	text := NewInput().WithDefaultValue(r.RawURL()).WithOnValidateFunc(func(s string) error {
		return r.ParseRawURL(s)
	})
	if _, err := text.Show(helpURL); err != nil {
		return err
	}
	return nil
}
func HttpMethod(r *httpx.Request) error {
	method, err := NewSelect().WithOptions(httpx.Methods).WithDefaultOption(r.Method).Show(helpMethod)
	if err != nil {
		return err
	}

	r.Method = method
	return nil
}
func HttpPathParams(r *httpx.Request) error {
	return validTomlValueText(&r.PathParams, helpPathParams)
}
func HttpQueryParams(r *httpx.Request) error {
	return validTomlValueText(&r.QueryParams, helpQueryParams)
}
func HttpHeaders(r *httpx.Request) error {
	return validTomlValueText(&r.Headers, helpHeaders)
}
func HttpSetting(r *httpx.Request) error {
	return validTomlValueText(&r.Settings, helpSettings)
}
func HttpDesc(r *httpx.Request) error {
	return multiLineStringText(&r.Desc, helpDesc)
}
func HttpBody(r *httpx.Request) error {
	b := &r.Body
	var err error

	b.Type, err = NewSelect().WithOptions(httpx.BodyTypes).WithDefaultOption(b.Type).Show(helpBodyType)
	if err != nil {
		return err
	}
	switch b.Type {
	default:
		return multiLineStringText(&b.Data.Raw, helpBodyRaw)
	case httpx.BodyBinary:
		return multiLineStringText(&b.Data.Raw, helpBodyBinary)
	case httpx.BodyJSON:
		return multiLineStringText(&b.Data.Raw, helpBodyJSON)
	case httpx.BodyXML:
		return multiLineStringText(&b.Data.Raw, helpBodyXML)
	case httpx.BodyFormUrlEncoded:
		return validTomlValueText(&b.Data.FormUrlEncoded, helpBodyFormUrlEncoded)
	case httpx.BodyMultiForm:
		fds := struct {
			Field []httpx.MultipartField `toml:"field,omitempty"`
		}{}
		if err := validTomlValueText(&fds, helpBodyMultiForm); err != nil {
			return err
		}
		b.Data.MultiForm = fds.Field
		return nil
	}
}
