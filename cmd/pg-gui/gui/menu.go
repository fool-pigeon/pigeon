package gui

import (
	// wails app
	"runtime"

	wa "github.com/wailsapp/wails/v3/pkg/application"
)

type MenuEvent = string

const (
	OpenCollection MenuEvent = "menu:open-collection"
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

func InitMenu(app *wa.App) {
	isDarwin := runtime.GOOS == "darwin"
	var (
		FileMenu = "File"
		EditMenu = "Edit"
		ViewMenu = "View"
		HelpMenu = "Help"
	)

	menu := app.NewMenu()
	app.Menu.Set(menu)

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
	// 		NewRequest
	// 		NewProject
	// 		----------
	// 		OpenCollection
	// 		----------
	// 		Save
	// 		----------
	// 		Preferences
	// 		----------
	// 		Quit
	NewMenuWrap(menu.AddSubmenu(FileMenu)).
		bindItemWithShortcut("New Request", "CmdOrCtrl+n", func(ctx *wa.Context) {
			app.Logger.Info("file:NewRequest")
		}).
		bindItemWithShortcut("New Collection", "CmdOrCtrl+Shift+n", func(ctx *wa.Context) {
			app.Logger.Info("file:NewProject")
		}).
		addSeparator().
		bindItemWithShortcut("Open Collection", "CmdOrCtrl+Shift+o", func(ctx *wa.Context) {
			app.Logger.Info("file:OpenCollection")
			app.Event.Emit(OpenCollection)
		}).
		addSeparator().
		bindItemWithShortcut("Save", "CmdOrCtrl+s", func(ctx *wa.Context) {
			app.Logger.Info("file:Save")
		}).
		addSeparator().
		bindItemWithShortcut("Preferences", "CmdOrCtrl+,", func(ctx *wa.Context) {
			app.Logger.Info("file:Preferences")
		}).
		addSeparator().
		bindRole(wa.Quit, "Quit", !isDarwin)

	// EditMenu:
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
	NewMenuWrap(menu.AddSubmenu(ViewMenu)).
		bindRole(wa.Reload, "Reload").
		bindRole(wa.ForceReload, "Force Reload").
		addSeparator().
		bindRole(wa.ResetZoom, "Reset Zoom").
		bindRoleWithShortcut(wa.ZoomIn, "Zoom In", "CmdOrCtrl+=").
		bindRoleWithShortcut(wa.ZoomOut, "Zoom Out", "CmdOrCtrl+-").
		addSeparator().
		bindRole(wa.ToggleFullscreen, "Toggle Fullscreen")
	// go:build !productioin || devtools
	menu.FindByLabel(ViewMenu).GetSubmenu().AddRole(wa.OpenDevTools)

	// HelpMenu:
	NewMenuWrap(menu.AddSubmenu(HelpMenu)).
		bindRole(wa.About, "About", !isDarwin).
		bindItem("Documentation", func(ctx *wa.Context) {
			app.Logger.Info("Documentation clicked")
		})
}

// this is a empty Service to export the enum to js/ts
type MenuService struct {
	me MenuEvent
}

func (ms *MenuService) Example() MenuEvent {
	return ms.me
}
