package menu

import (
	"fmt"

	"fyne.io/fyne/v2"
)

func CreateMainMenu() *fyne.MainMenu {
	open := fyne.NewMenuItem("open", func() {})
	save := fyne.NewMenuItem("save", func() {})

	filemenu := fyne.NewMenu("file", open, save)

	mainMenu := fyne.NewMainMenu(filemenu)
	return mainMenu
}

type Menu interface {
	ParentLabel() string
	NewMenuItem() *fyne.MenuItem
}

type OpenMenu struct{}

func (om *OpenMenu) ParentLabel() string {
	return "file"
}

func (om *OpenMenu) label() string {
	return "open"
}

func (om *OpenMenu) action() func() {
	return func() {
		fmt.Printf("open...")
	}
}

func (om *OpenMenu) NewMenuItem() *fyne.MenuItem {
	m1 := fyne.NewMenuItem(om.label(), om.action())
	return m1
}

type initMenu func() *fyne.Menu

func newFileMenu() *fyne.Menu {
	open := fyne.NewMenuItem("open", func() {})
	save := fyne.NewMenuItem("save", func() {})
	filemenu := fyne.NewMenu("file", open, save)
	return filemenu
}

func NewMenu() *fyne.MainMenu {
	conttrollers := []initMenu{
		newFileMenu,
	}

	ms := make([]*fyne.Menu, 0, len(conttrollers))

	for i, c := range conttrollers {
		ms[i] = c()
	}

	mainMenu := fyne.NewMainMenu(ms...)
	return mainMenu

}
