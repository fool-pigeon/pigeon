// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import "github.com/wailsapp/wails/v3/pkg/application"

type AlertEvent = string

const (
	AlertInfo    AlertEvent = "alert:info"
	AlertSuccess AlertEvent = "alert:success"
	AlertWarn    AlertEvent = "alert:warn"
	AlertError   AlertEvent = "alert:error"
)

var alert alertNotify

type alertNotify struct{ evt *application.EventManager }

func InitAlertNotify(app *application.App) {
	alert.evt = app.Event
}
func (as *alertNotify) msg(alert AlertEvent, msg string) {
	as.evt.Emit(alert, msg)
}
func (as *alertNotify) info(msg string) {
	as.msg(AlertInfo, msg)
}
func (as *alertNotify) success(msg string) {
	as.msg(AlertSuccess, msg)
}
func (as *alertNotify) warn(msg string) {
	as.msg(AlertWarn, msg)
}
func (as *alertNotify) error(msg string) {
	as.msg(AlertError, msg)
}
