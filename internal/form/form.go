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
	*name, err = NewInput().Show("Input request's name")
	if err != nil {
		return err
	}
	if *name == "" {
		return errors.New("name inputed is empty.")
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

// interactive a value util the value can be validated or the interactive casue a err
// exec sequence: InitFunc -> Validate -> EndFuc
type validTextInput[T any] struct {
	ptr         *T
	InitFunc    func(v *T) (string, error)
	Validate    func(s string, v *T) error
	EndFunc     func(s string, v *T)
	Desc        []string
	isMultiLine bool
}

func NewValidText[T any]() *validTextInput[T] {
	return &validTextInput[T]{}
}
func (t *validTextInput[T]) SetDesc(desc ...string) *validTextInput[T] {
	t.Desc = desc
	return t
}
func (t *validTextInput[T]) Value(v *T) *validTextInput[T] {
	t.ptr = v
	return t
}
func (t *validTextInput[T]) WithInitFunc(f func(v *T) (string, error)) *validTextInput[T] {
	t.InitFunc = f
	return t
}
func (t *validTextInput[T]) WithValidate(f func(s string, v *T) error) *validTextInput[T] {
	t.Validate = f
	return t
}

func (t *validTextInput[T]) WithEndFunc(f func(s string, v *T)) *validTextInput[T] {
	t.EndFunc = f
	return t
}
func (t *validTextInput[T]) WithMultiLine() *validTextInput[T] {
	t.isMultiLine = true
	return t
}
func (t *validTextInput[T]) Run() error {
	for _, desc := range t.Desc {
		pterm.Println(pterm.LightGreen(desc))
	}

	text := NewInput().WithDefaultText(pterm.Sprintf("%T", t.ptr))
	if t.isMultiLine {
		text = text.WithMultiLine()
	}
	// init
	if t.InitFunc != nil {
		initValue, err := t.InitFunc(t.ptr)
		if err != nil {
			return err
		}
		text = text.WithDefaultValue(initValue)
	}
	for {
		value, err := text.Show()
		if err != nil {
			return err
		}

		// validate
		if valid := t.Validate; valid != nil {
			if err := valid(value, t.ptr); err != nil {
				pterm.Error.Println(err)
				continue
			}
		}

		// end
		if t.EndFunc != nil {
			t.EndFunc(value, t.ptr)
		}
		return nil
	}
}

// bind the v with a multiLineText
func multiLineStringText(v *string, desc ...string) error {
	for _, d := range desc {
		pterm.Println(pterm.LightGreen(d))
	}

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
	return NewValidText[T]().
		WithMultiLine().
		SetDesc(desc...).
		Value(v).
		WithInitFunc(func(v *T) (string, error) {
			b, err := toml.Marshal(v)
			return string(b), err
		}).
		WithValidate(func(s string, v *T) error {
			return toml.Unmarshal([]byte(s), v)
		}).
		Run()
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
func HttpForm(r *httpx.Request) error {
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
		WithDefaultOptions([]string{httpx.FieldURL, httpx.FieldMethod}).
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
	return NewValidText[string]().
		SetDesc(helpURL).
		WithValidate(func(s string, v *string) error {
			return r.ParseRawURL(s)
		}).
		Run()
}
func HttpMethod(r *httpx.Request) error {
	method, err := NewSelect().WithOptions(httpx.Methods).Show(helpMethod)
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

	b.Type, err = NewSelect().WithOptions(httpx.BodyTypes).Show(helpBodyType)
	if err != nil {
		return err
	}
	var desc string
	switch b.Type {
	default:
		desc = helpBodyRaw
	case httpx.BodyBinary:
		desc = helpBodyBinary
	case httpx.BodyJSON:
		desc = helpBodyJSON
	case httpx.BodyXML:
		desc = helpBodyXML
	case httpx.BodyFormUrlEncoded:
		desc = helpBodyFormUrlEncoded
	case httpx.BodyMultiForm:
		desc = helpBodyMultiForm

	}

	switch b.Type {
	default:
		return multiLineStringText(&b.Data.Raw, desc)
	case httpx.BodyFormUrlEncoded:
		return validTomlValueText(&b.Data.FormUrlEncoded, desc)
	case httpx.BodyMultiForm:
		fds := struct {
			Field []httpx.MultipartField `toml:"field"`
		}{}
		if err := validTomlValueText(&fds, desc); err != nil {
			return err
		}
		b.Data.MultiForm = fds.Field
	}
	return nil
}
