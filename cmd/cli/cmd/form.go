package cmd

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
	"github.com/homoarea/pigeon/internal/request/httpx"
	"github.com/pelletier/go-toml/v2"
)

var formKeyMap = huh.NewDefaultKeyMap()

func initFormKeyMap() {
	formKeyMap.MultiSelect.SetFilter = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "set filter"), key.WithDisabled())
}

func formWrap(groups ...*huh.Group) *huh.Form {
	return huh.NewForm(groups...).
		WithKeyMap(formKeyMap).
		WithShowHelp(true)
}

// bind the v with tomlParser
func tomlTextFormField(v interface{}) *huh.Text {
	b, _ := toml.Marshal(v)
	initValue := string(b)
	return huh.NewText().
		ShowLineNumbers(true).
		Value(&initValue).
		Validate(func(s string) error {
			return toml.Unmarshal([]byte(s), v)
		})
}

func httpBodyFormGroup(b *httpx.Body) *huh.Group {
	return huh.NewGroup(
		huh.NewSelect[string]().
			Title("Select body type").
			Options(huh.NewOptions(httpx.BodyTypes...)...).
			Value(&b.Type),
		huh.NewText().
			ShowLineNumbers(true).
			TitleFunc(func() string {
				return b.Type
			}, &b.Type).
			DescriptionFunc(func() string {
				switch b.Type {
				case httpx.BodyRaw, httpx.BodyJSON, httpx.BodyXML, httpx.BodyText:
					return "type string"
				case httpx.BodyBinary:
					return "type string\ninput file_path:"
				case httpx.BodyFormUrlEncoded:
					return `type map[string][]string
example:
	k1=['v1','v2']
	k2=['v3','v3']`
				case httpx.BodyMultiForm:
					return `type []MultipartField
example:
	[[field]]
	name=''
	values=['v1','v2']
	content_type=''
	file_name=''
	file_path=''
`
				}
				return ""
			}, &b.Type).
			Validate(func(s string) error {
				if s == "" {
					return nil
				}
				var v interface{}
				switch b.Type {
				case httpx.BodyJSON:
					if err := json.Unmarshal([]byte(s), &v); err != nil {
						return err
					}
					b.Data.Raw = s
				case httpx.BodyXML:
					if err := xml.Unmarshal([]byte(s), &v); err != nil && !errors.Is(err, io.EOF) {
						return err
					}
					b.Data.Raw = s
				case httpx.BodyText, httpx.BodyRaw:
					b.Data.Raw = s
				case httpx.BodyMultiForm:
					fds := struct {
						Field []httpx.MultipartField `toml:"field"`
					}{}

					if err := toml.NewDecoder(strings.NewReader(s)).DisallowUnknownFields().Decode(&fds); err != nil {
						return err
					}
					b.Data.MultiForm = fds.Field
				case httpx.BodyFormUrlEncoded:
					if err := toml.Unmarshal([]byte(s), &b.Data.FormUrlEncoded); err != nil {
						return err
					}
				case httpx.BodyBinary:
					f, err := os.Lstat(s)
					if err != nil {
						return err
					}
					if f.IsDir() {
						return errors.New(s + " is not a file.")
					}
					b.Data.BinaryPath = s
				}
				return nil
			}),
	)
}
