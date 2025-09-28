package appui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*ItemContainer)(nil)
var _ fyne.Tappable = (*ItemContainer)(nil)
var _ fyne.SecondaryTappable = (*ItemContainer)(nil)

// ItemContainer 是单个列表项，只负责显示文字和点击回调
type ItemContainer struct {
	widget.BaseWidget
	label          *widget.Label
	index          int
	onTapped       func(index int)
	onRightClicked func(index int, pos fyne.Position)
}

// NewItemContainer 创建新ItemContainer
func NewItemContainer(onTapped func(int), onRightClicked func(int, fyne.Position)) *ItemContainer {
	ic := &ItemContainer{
		label:          widget.NewLabel(""),
		onTapped:       onTapped,
		onRightClicked: onRightClicked,
	}
	ic.ExtendBaseWidget(ic)
	return ic
}

// CreateRenderer 实现 fyne.Widget 接口
func (ic *ItemContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(ic.label)
}

// SetText 更新显示文本
func (ic *ItemContainer) SetText(text string) {
	ic.label.SetText(text)
}

// SetIndex 设置当前索引
func (ic *ItemContainer) SetIndex(i int) {
	ic.index = i
}

// Tapped 左键点击
func (ic *ItemContainer) Tapped(pe *fyne.PointEvent) {
	if ic.onTapped != nil {
		ic.onTapped(ic.index)
	}
}

// TappedSecondary 右键点击
func (ic *ItemContainer) TappedSecondary(pe *fyne.PointEvent) {
	if ic.onRightClicked != nil {
		ic.onRightClicked(ic.index, pe.AbsolutePosition)
	}
}
