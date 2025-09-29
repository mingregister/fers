package appui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*ItemContainer)(nil)
var _ fyne.Tappable = (*ItemContainer)(nil)
var _ fyne.SecondaryTappable = (*ItemContainer)(nil)
var _ fyne.DoubleTappable = (*ItemContainer)(nil)

// ItemContainer 是单个列表项，只负责显示文字和点击回调
type ItemContainer struct {
	widget.BaseWidget
	label            *widget.Label
	background       *canvas.Rectangle
	containerObj     fyne.CanvasObject
	index            int
	selected         bool
	selectionColor   color.Color
	transparentColor color.Color
	onTapped         func(index int)
	onRightClicked   func(index int, pos fyne.Position)
	onDoubleTapped   func(index int)
}

// NewItemContainer 创建新ItemContainer
func NewItemContainer(onTapped func(int), onRightClicked func(int, fyne.Position)) *ItemContainer {
	label := widget.NewLabel("")
	background := canvas.NewRectangle(color.Transparent)

	ic := &ItemContainer{
		label:            label,
		background:       background,
		containerObj:     container.NewBorder(nil, nil, nil, nil, label),
		selectionColor:   theme.Color(theme.ColorNameSelection),
		transparentColor: color.Transparent,
		onTapped:         onTapped,
		onRightClicked:   onRightClicked,
	}
	ic.ExtendBaseWidget(ic)
	return ic
}

// SetOnDoubleTapped 设置双击回调
func (ic *ItemContainer) SetOnDoubleTapped(callback func(int)) {
	ic.onDoubleTapped = callback
}

// CreateRenderer 实现 fyne.Widget 接口
func (ic *ItemContainer) CreateRenderer() fyne.WidgetRenderer {
	return &itemContainerRenderer{
		container:  ic,
		background: ic.background,
		content:    ic.containerObj,
	}
}

// SetText 更新显示文本
func (ic *ItemContainer) SetText(text string) {
	ic.label.SetText(text)
}

// SetIndex 设置当前索引
func (ic *ItemContainer) SetIndex(i int) {
	ic.index = i
}

// SetSelected 设置选中状态
func (ic *ItemContainer) SetSelected(selected bool) {
	if ic.selected == selected {
		return // 状态没有变化，直接返回
	}

	ic.selected = selected

	// 使用缓存的颜色，避免重复查询主题系统
	if selected {
		ic.background.FillColor = ic.selectionColor
	} else {
		ic.background.FillColor = ic.transparentColor
	}

	// 只刷新背景，避免双重刷新
	ic.background.Refresh()
}

// IsSelected 获取选中状态
func (ic *ItemContainer) IsSelected() bool {
	return ic.selected
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

// DoubleTapped 双击
func (ic *ItemContainer) DoubleTapped(pe *fyne.PointEvent) {
	if ic.onDoubleTapped != nil {
		ic.onDoubleTapped(ic.index)
	}
}

// itemContainerRenderer 自定义渲染器
type itemContainerRenderer struct {
	container  *ItemContainer
	background *canvas.Rectangle
	content    fyne.CanvasObject
}

func (r *itemContainerRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.content.Resize(size)
}

func (r *itemContainerRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
}

func (r *itemContainerRenderer) Refresh() {
	r.background.Refresh()
	r.content.Refresh()
}

func (r *itemContainerRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.content}
}

func (r *itemContainerRenderer) Destroy() {}
