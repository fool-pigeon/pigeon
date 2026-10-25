// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package service

import (
	"github.com/fool-pigeon/pigeon/cmd/pg-gui/gui"
	"github.com/fool-pigeon/pigeon/protocol"
)

// A service to auto expose Event const nad Events data struct type
type EventConstExposeService struct{}

func (e *EventConstExposeService) ExposedMenuEvt(evt gui.MenuEvent) {}

func (e *EventConstExposeService) ExposedFsEvt(evt FsEvent) {}

func (e *EventConstExposeService) ExposedFseEvt(evt FesEvent, fesUpdateSubs FesEvtUpdateSubs) {}

func (e *EventConstExposeService) ExposedAlertEvt(evt AlertEvent) {}

func (e *EventConstExposeService) ExposedTrsEvt(
	evt TaskSEvent,
	ss TaskSStream,
	taskResViewMeta TaskResViewMeta,
	traceView protocol.TraceView,
) {
}

func (e *EventConstExposeService) ExposedSettingsServiceEvent(evt SettingsServiceEvent)   {}
func (e *EventConstExposeService) ExposedExtendedWindowEvent(evt gui.ExtendedWindowEvent) {}
