package formatter

import (
	"strings"

	"github.com/pterm/pterm"
)

func FieldColor(a ...any) string {
	return pterm.LightBlue(a...)
}
func FieldLine(name string, val any) string {
	return pterm.Sprintfln("%v %v", FieldColor(name), val)
}
func PrintValues(title string, isAddLine bool, content ...string) {
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

func FormatValues(data map[string][]string, sep ...string) []string {
	s := ", "
	if len(sep) > 0 {
		s = sep[0]
	}
	values := make([]string, 0, len(data))
	for k, v := range data {
		values = append(values, pterm.Sprintf("%v %v", FieldColor(k, ":"), strings.Join(v, FieldColor(s))))
	}
	return values
}

func FlattenValues(data map[string][]string) []string {
	values := make([]string, 0)
	for k, v := range data {
		for _, vv := range v {
			values = append(values, pterm.Sprintf("%v %v", FieldColor(k, ":"), vv))
		}
	}
	return values
}
func FormatMap(data map[string]string) []string {
	values := make([]string, 0, len(data))
	for k, v := range data {
		values = append(values, FieldLine(k, v))
	}
	return values
}
