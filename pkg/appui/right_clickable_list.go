package appui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*RightClickableList)(nil)

// RightClickableList 是可右键点击的列表控件
type RightClickableList struct {
	widget.BaseWidget
	list             *widget.List
	items            []string
	OnItemTapped     func(index int)
	OnItemRightClick func(index int, pos fyne.Position)
}

// NewRightClickableList 创建新RightClickableList
func NewRightClickableList() *RightClickableList {
	rcl := &RightClickableList{}
	rcl.ExtendBaseWidget(rcl)
	return rcl
}

// SetItems 设置列表数据
func (rcl *RightClickableList) SetItems(items []string) {
	rcl.items = items
	if rcl.list != nil {
		rcl.list.Refresh()
	}
}

// Build 构建内部widget.List
func (rcl *RightClickableList) Build() {
	rcl.list = widget.NewList(
		func() int { return len(rcl.items) },
		func() fyne.CanvasObject {
			return NewItemContainer(
				func(i int) {
					if rcl.OnItemTapped != nil {
						rcl.OnItemTapped(i)
					}
				},
				func(i int, pos fyne.Position) {
					if rcl.OnItemRightClick != nil {
						rcl.OnItemRightClick(i, pos)
					}
				},
			)
		},
		func(i int, o fyne.CanvasObject) {
			itemContainer := o.(*ItemContainer)
			itemContainer.SetText(rcl.items[i])
			itemContainer.SetIndex(i)
		},
	)
}

// CreateRenderer 实现 fyne.Widget 接口
func (rcl *RightClickableList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(rcl.list)
}

// Refresh 刷新显示
func (rcl *RightClickableList) Refresh() {
	if rcl.list != nil {
		rcl.list.Refresh()
	}
}

// UnselectAll 取消选中
func (rcl *RightClickableList) UnselectAll() {
	if rcl.list != nil {
		rcl.list.UnselectAll()
	}
}

// GetList 返回内部widget.List
func (rcl *RightClickableList) GetList() *widget.List {
	return rcl.list
}
