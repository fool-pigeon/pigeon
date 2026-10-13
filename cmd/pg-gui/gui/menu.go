// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package gui

import (
	// wails app
	"runtime"

	wa "github.com/wailsapp/wails/v3/pkg/application"
)

type MenuEvent = string

const (
	// File Subs
	NewFile        MenuEvent = "menu:file:new-file"
	NewCollection  MenuEvent = "menu:file:new-collection"
	OpenCollection MenuEvent = "menu:file:open-collection"
	Save           MenuEvent = "menu:file:save"
	Preferences    MenuEvent = "menu:file:preferences"
	CloseEditor    MenuEvent = "menu:file:close-editor"
	// View Subs
	ToggleRunBoard MenuEvent = "menu:view:toggle-run-board"
	// Run Subs
	RunTask       MenuEvent = "menu:run:run-task"
	RunRequest    MenuEvent = "menu:run:run-request"
	CancelRequest MenuEvent = "menu:run:cancel-request"
	// Help Subs
	Documentation MenuEvent = "menu:help:docmentation"
)

type MenuWrap struct {
	menu *wa.Menu
}

func NewMenuWrap(menu *wa.Menu) *MenuWrap {
	return &MenuWrap{menu}
}
func (m *MenuWrap) addSeparator() *MenuWrap {
	m.menu.AddSeparator()
	return m
}
func (m *MenuWrap) bindSubMenu(label string, f func(mw *MenuWrap), condition ...bool) *MenuWrap {
	if len(condition) <= 0 || condition[0] {
		sm := m.menu.AddSubmenu(label)
		f(NewMenuWrap(sm))
	}
	return m
}
func (m *MenuWrap) bindRole(role wa.Role, label string, condition ...bool) *MenuWrap {
	if len(condition) <= 0 || condition[0] {
		m.menu.AddRole(role).FindByRole(role).SetLabel(label)
	}
	return m
}
func (m *MenuWrap) bindRoleWithShortcut(role wa.Role, label string, shortcut string, condition ...bool) *MenuWrap {
	if len(condition) <= 0 || condition[0] {
		m.menu.AddRole(role).FindByRole(role).SetAccelerator(shortcut).SetLabel(label)
	}
	return m
}
func (m *MenuWrap) bindItem(label string, on func(ctx *wa.Context), condition ...bool) *MenuWrap {
	if len(condition) <= 0 || condition[0] {
		bindItem(m.menu, label, on)
	}
	return m
}
func (m *MenuWrap) bindItemWithShortcut(label string, shortcut string, on func(ctx *wa.Context), condition ...bool) *MenuWrap {
	if len(condition) <= 0 || condition[0] {
		bindShortcutedItem(m.menu, label, shortcut, on)
	}
	return m
}
func bindItem(menu *wa.Menu, label string, on func(ctx *wa.Context)) {
	menu.Add(label).OnClick(on)
}
func bindShortcutedItem(menu *wa.Menu, label string, shortcut string, on func(ctx *wa.Context)) {
	menu.Add(label).SetAccelerator(shortcut).OnClick(on)
}

func MainMenu(app *wa.App) *wa.Menu {
	isDarwin := runtime.GOOS == "darwin"
	var (
		FileMenu = "File"
		EditMenu = "Edit"
		ViewMenu = "View"
		RunMenu  = "Run"
		HelpMenu = "Help"
	)

	menu := app.NewMenu()

	// darwin: AppMenu:
	if isDarwin {
		NewMenuWrap(menu.AddSubmenu(app.Config().Name)).
			bindRole(wa.About, "About").
			addSeparator().
			bindRole(wa.Hide, "Hide").
			bindRole(wa.HideOthers, "Hide Others").
			bindRole(wa.UnHide, "Show All").
			addSeparator().
			bindRole(wa.Quit, "Quit")
	}
	// FileMenu:
	// 	New Request
	// 	New Project
	// 	----------
	// 	Open Collection
	// 	----------
	// 	Save
	// 	----------
	// 	Preferences
	// 	----------
	// 	Close Editor
	// 	----------
	// 	Quit
	NewMenuWrap(menu.AddSubmenu(FileMenu)).
		bindItemWithShortcut("New File", "CmdOrCtrl+n", func(ctx *wa.Context) {
			app.Logger.Info(NewFile)
			app.Event.Emit(NewFile)
		}).
		bindItemWithShortcut("New Collection", "CmdOrCtrl+Shift+n", func(ctx *wa.Context) {
			app.Logger.Info(NewCollection)
			app.Event.Emit(NewCollection)
		}).
		addSeparator().
		bindItemWithShortcut("Open Collection", "CmdOrCtrl+Shift+o", func(ctx *wa.Context) {
			app.Logger.Info(OpenCollection)
			app.Event.Emit(OpenCollection)
		}).
		addSeparator().
		bindItemWithShortcut("Save", "CmdOrCtrl+s", func(ctx *wa.Context) {
			app.Logger.Info(Save)
			app.Event.Emit(Save)
		}).
		addSeparator().
		bindItemWithShortcut("Preferences", "CmdOrCtrl+,", func(ctx *wa.Context) {
			app.Logger.Info(Preferences)
			app.Event.Emit(Preferences)
		}).
		addSeparator().
		bindItemWithShortcut("Close Editor", "CmdOrCtrl+w", func(ctx *wa.Context) {
			app.Logger.Info(CloseEditor)
			app.Event.Emit(CloseEditor)
		}).
		addSeparator().
		bindRole(wa.Quit, "Quit", !isDarwin)

	// EditMenu:
	// 	Undo
	// 	Redo
	// 	----------
	// 	Cut
	// 	Copy
	// 	Paste
	// 	(Paste & Match Style)
	// 	----------
	// 	(Speech)
	// 	SelectAll
	NewMenuWrap(menu.AddSubmenu(EditMenu)).
		bindRole(wa.Undo, "Undo").
		bindRole(wa.Redo, "Redo").
		addSeparator().
		bindRole(wa.Cut, "Cut").
		bindRole(wa.Copy, "Copy").
		bindRole(wa.Paste, "Paste").
		bindRole(wa.PasteAndMatchStyle, "Paste & Match Style", isDarwin).
		bindRole(wa.Delete, "Delete").
		addSeparator().
		bindSubMenu("Speech", func(mw *MenuWrap) {
			mw.bindRole(wa.StartSpeaking, "Start Speaking")
			mw.bindRole(wa.StopSpeaking, "Stop Speaking")
		}, isDarwin).
		bindRole(wa.SelectAll, "Selete All")

	// ViewMenu:
	// 	----------
	//  Toggle Run Board
	// 	Reload
	// 	Force Reload
	// 	----------
	// 	Reset Zoom
	// 	Zoom In
	// 	Zoom Out
	// 	----------
	// 	Toggle Fullscreen
	NewMenuWrap(menu.AddSubmenu(ViewMenu)).
		bindItem("Toggle Run Board", func(ctx *wa.Context) {
			app.Logger.Info(ToggleRunBoard)
			app.Event.Emit(ToggleRunBoard)
		}).
		addSeparator().
		bindRole(wa.Reload, "Reload").
		bindRole(wa.ForceReload, "Force Reload").
		addSeparator().
		bindRole(wa.ResetZoom, "Reset Zoom").
		bindRoleWithShortcut(wa.ZoomIn, "Zoom In", "CmdOrCtrl+=").
		bindRoleWithShortcut(wa.ZoomOut, "Zoom Out", "CmdOrCtrl+-").
		addSeparator().
		bindRole(wa.ToggleFullscreen, "Toggle Fullscreen")

	// RunMenu:
	//  Run Task
	// 	----------
	// 	Run Request
	// 	Cancel Request
	NewMenuWrap(menu.AddSubmenu(RunMenu)).
		bindItem("Run Task", func(ctx *wa.Context) {
			app.Logger.Info(RunTask)
			app.Event.Emit(RunTask)
		}).
		addSeparator().
		bindItemWithShortcut("Run Request", "CmdOrCtrl+Enter", func(ctx *wa.Context) {
			app.Logger.Info(RunRequest)
			app.Event.Emit(RunRequest)
		}).
		bindItem("Cancel Request", func(ctx *wa.Context) {
			app.Logger.Info(CancelRequest)
			app.Event.Emit(CancelRequest)
		})

	// HelpMenu:
	// 	About
	// 	Documentation
	NewMenuWrap(menu.AddSubmenu(HelpMenu)).
		bindRole(wa.About, "About", !isDarwin).
		bindItem("Documentation", func(ctx *wa.Context) {
			app.Logger.Info(Documentation)
			app.Event.Emit(Documentation)
		})

	// go:build !productioin || devtools
	menu.FindByLabel(HelpMenu).GetSubmenu().AddRole(wa.OpenDevTools)

	return menu
}
