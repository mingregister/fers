package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	App = app.NewWithID("fers")
)

func main() {
	w := App.NewWindow("fers")
	w.Resize(fyne.NewSize(1200, 900))
	w.CenterOnScreen()
	menu := createMenu()
	content := createContent()
	sidebar := createSideBar()

	layout := container.NewBorder(nil, nil, sidebar, nil, content)

	w.SetMainMenu(menu)
	w.SetContent(layout)

	w.Show()
	App.Run()
}

func createMenu() *fyne.MainMenu {
	open := fyne.NewMenuItem("open", func() {})
	save := fyne.NewMenuItem("save", func() {})

	filemenu := fyne.NewMenu("file", open, save)

	mainMenu := fyne.NewMainMenu(filemenu)
	return mainMenu
}

func createSideBar() *fyne.Container {
	sidebar := container.NewVBox(
		widget.NewButton("Home", func() { println("Home clicked") }),
		widget.NewButton("Settings", func() { println("Settings clicked") }),
		widget.NewButton("About", func() { println("About clicked") }),
	)
	return sidebar
}

func createContent() *fyne.Container {
	// 主内容区
	hello := widget.NewLabel("你好 Fyne!")
	hi := widget.NewLabel("hi Fyne!")
	content := container.NewVBox(
		hello,
		widget.NewButton("Hi!", func() {
			hello.SetText("Welcome :)")
		}),
		hi,
		widget.NewButton("Hi!", func() {
			hi.SetText("HI Welcome :)")
		}),
	)
	return content
}
