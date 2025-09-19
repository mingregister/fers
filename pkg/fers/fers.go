package fers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/mingregister/fers/pkg/menu"
)

type Fers struct {
	App    fyne.App
	Window fyne.Window
}

func NewFers(appid string, winid string) *Fers {
	app := app.NewWithID(appid)
	// 配置Window
	w := app.NewWindow(winid)
	w.Resize(fyne.NewSize(1200, 900))
	w.CenterOnScreen()

	f := Fers{
		App:    app,
		Window: w,
	}
	return &f
}

func (f *Fers) SetMainMenu() {
	m := menu.CreateMainMenu()
	f.Window.SetMainMenu(m)
}

func (f *Fers) SetContent(container *fyne.Container) {
	f.Window.SetContent(container)
}

func (f *Fers) ShowWindow() {
	f.Window.Show()
}

func (f *Fers) Run() {
	f.App.Run()
}
