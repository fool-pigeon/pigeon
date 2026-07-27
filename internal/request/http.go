package request

import (
	"os"
	"text/template"
)

type HttpSettings struct {
	EncodeUrl     bool `toml:"encode_url,omitempty"`
	Timeout       uint `toml:"timeout,omitempty"`
	AutoRedirects bool `toml:"auto_redirects"`
	MaxRedirects  uint `toml:"max_redirects,omitempty"`
}
type HttpRequest struct {
	RequestMeta
	Method   string       `toml:"mothod"`
	Url      string       `toml:"url"`
	Params   [][2]string  `toml:"params,omitempty,multiline"`
	Headers  [][2]string  `toml:"headers,omitempty,multiline"`
	BodyType string       `toml:"body_type,omitempty"`
	Body     string       `toml:"body,omitempty,multiline"`
	Settings HttpSettings `toml:"settings"`
	Docs     string       `toml:"docs,omitempty,multiline"`
}

func (r *HttpRequest) Send() error {
	tplStr := "name:{{.Meta.Name}} sended!\n"
	tpl, err := template.New("req").Parse(tplStr)
	if err != nil {
		return err
	}
	tpl.Execute(os.Stdout, r)
	return nil
}
