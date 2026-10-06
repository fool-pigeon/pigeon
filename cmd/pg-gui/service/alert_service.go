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

type AlertService struct{ app *application.App }

func NewAlertService(app *application.App) *AlertService {
	return &AlertService{app}
}
func (as *AlertService) Msg(alert AlertEvent, msg string) {
	as.app.Event.Emit(alert, msg)
}
func (as *AlertService) Info(msg string) {
	as.Msg(AlertInfo, msg)
}
func (as *AlertService) Success(msg string) {
	as.Msg(AlertSuccess, msg)
}
func (as *AlertService) Warn(msg string) {
	as.Msg(AlertWarn, msg)
}
func (as *AlertService) Error(msg string) {
	as.Msg(AlertError, msg)
}
