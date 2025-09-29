package appui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

var _ fyne.Widget = (*RightClickableList)(nil)

// RightClickableList 是可右键点击的列表控件
type RightClickableList struct {
	widget.BaseWidget
	list               *widget.List
	items              []string
	selectedIndex      int
	OnItemTapped       func(index int)
	OnItemRightClick   func(index int, pos fyne.Position)
	OnItemDoubleTapped func(index int)
}

// NewRightClickableList 创建新RightClickableList
func NewRightClickableList() *RightClickableList {
	rcl := &RightClickableList{
		selectedIndex: -1,
	}
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
			itemContainer := NewItemContainer(
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
			// 设置双击回调
			itemContainer.SetOnDoubleTapped(func(i int) {
				if rcl.OnItemDoubleTapped != nil {
					rcl.OnItemDoubleTapped(i)
				}
			})
			return itemContainer
		},
		func(i int, o fyne.CanvasObject) {
			itemContainer := o.(*ItemContainer)
			itemContainer.SetText(rcl.items[i])
			itemContainer.SetIndex(i)
			itemContainer.SetSelected(i == rcl.selectedIndex)
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

// SetSelectedIndex 设置选中的索引
func (rcl *RightClickableList) SetSelectedIndex(index int) {
	rcl.selectedIndex = index
	if rcl.list != nil {
		rcl.list.Refresh()
	}
}

// GetSelectedIndex 获取选中的索引
func (rcl *RightClickableList) GetSelectedIndex() int {
	return rcl.selectedIndex
}

// UnselectAll 取消选中
func (rcl *RightClickableList) UnselectAll() {
	rcl.selectedIndex = -1
	if rcl.list != nil {
		rcl.list.UnselectAll()
		rcl.list.Refresh()
	}
}

// GetList 返回内部widget.List
func (rcl *RightClickableList) GetList() *widget.List {
	return rcl.list
}
