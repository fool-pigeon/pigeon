package cmd

import (
	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
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
