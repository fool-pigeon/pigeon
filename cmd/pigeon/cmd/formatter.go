// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package cmd

import (
	"strings"

	"github.com/pterm/pterm"
)

var formatter formatPrinter

type formatPrinter struct{}

func (f formatPrinter) fieldColor(a ...any) string {
	return pterm.LightBlue(a...)
}
func (f formatPrinter) fieldLine(name string, val any) string {
	return pterm.Sprintfln("%v %v", f.fieldColor(name), val)
}
func (f formatPrinter) printValues(title string, isAddLine bool, content ...string) {
	var sb strings.Builder
	sb.WriteString(pterm.Cyan(title, ": "))
	if isAddLine {
		sb.WriteString("\n")
	}
	for _, v := range content {
		for _, line := range strings.Split(v, "\n") {
			if isAddLine {
				sb.WriteString("  ")
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}
	pterm.Print(sb.String())
}

func (f formatPrinter) formatValues(data map[string][]string, sep ...string) []string {
	s := ", "
	if len(sep) > 0 {
		s = sep[0]
	}
	values := make([]string, 0, len(data))
	for k, v := range data {
		values = append(values, pterm.Sprintf("%v %v", f.fieldColor(k, ":"), strings.Join(v, f.fieldColor(s))))
	}
	return values
}

func (f formatPrinter) flattenValues(data map[string][]string) []string {
	values := make([]string, 0)
	for k, v := range data {
		for _, vv := range v {
			values = append(values, pterm.Sprintf("%v %v", f.fieldColor(k, ":"), vv))
		}
	}
	return values
}
func (f formatPrinter) formatMap(data map[string]string) []string {
	values := make([]string, 0, len(data))
	for k, v := range data {
		values = append(values, f.fieldLine(k, v))
	}
	return values
}
