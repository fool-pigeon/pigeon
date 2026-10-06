// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
)

// A service to auto expose Event const nad Events data struct type
type EventConstExposeService struct{}

func (e *EventConstExposeService) ExposedMenuEvt(evt gui.MenuEvent) {}

func (e *EventConstExposeService) ExposedFsEvt(evt FsEvent) {}

func (e *EventConstExposeService) ExposedFseEvt(evt FesEvent, fesUpdateSubs FesEvtUpdateSubs) {}

func (e *EventConstExposeService) ExposedAlertEvt(evt AlertEvent) {}

func (e *EventConstExposeService) ExposedTrsEvt(
	evt TrsEvent,
	trErr TaskResErr,
	trHttp TaskResHttp,
	trAny TaskResAny,
) {
}
