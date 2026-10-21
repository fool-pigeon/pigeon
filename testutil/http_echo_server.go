// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package testutil

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type EchoReq struct {
	Method      string              `json:"method"`
	URL         string              `json:"url"`
	Proto       string              `json:"proto"`
	Headers     map[string][]string `json:"headers"`
	Query       map[string][]string `json:"query"`
	Form        map[string][]string `json:"form"`
	Multipart   map[string][]string `json:"multipart"`
	Body        string              `json:"body"`
	BodyBase64  string              `json:"body_base64"`
	ContentType string              `json:"content_type"`
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		_ = r.ParseMultipartForm(32 << 20)
	} else {
		_ = r.ParseForm()
	}

	b, _ := io.ReadAll(r.Body)
	bstr := string(b)
	bBase64 := base64.StdEncoding.EncodeToString(b)

	var multipart map[string][]string
	if form := r.MultipartForm; form != nil {
		multipart = form.Value
	}
	echoReq := EchoReq{
		Method:      r.Method,
		URL:         r.URL.String(),
		Proto:       r.Proto,
		Headers:     r.Header,
		Query:       r.URL.Query(),
		Form:        r.Form,
		Multipart:   multipart,
		Body:        bstr,
		BodyBase64:  bBase64,
		ContentType: r.Header.Get("Content-Type"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(echoReq)

}
func ParseEchoResp(res []byte) EchoReq {
	var e EchoReq
	json.Unmarshal(res, &e)
	return e
}
func NewEchoServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(EchoHandler))
}
