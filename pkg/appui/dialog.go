package appui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

type DiaLogKind string

var (
	InfoDialog  DiaLogKind = "ShowInformation"
	ErrorDialog DiaLogKind = "ShowError"
)

func ShowDialog(kind DiaLogKind, parent fyne.Window, title, message string) {
	switch kind {
	case InfoDialog:
		fyne.Do(func() {
			dialog.ShowInformation(title, message, parent)
		})
	default:
		fyne.Do(func() {
			dialog.ShowInformation(title, message, parent)
		})
	}
}

func ShowDialogError(err error, parent fyne.Window) {
	fyne.Do(func() {
		dialog.ShowError(err, parent)
	})
}
