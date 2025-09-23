package appui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestRightClickableList_SetItems(t *testing.T) {
	rcl := NewRightClickableList()

	testItems := []string{"item1", "item2", "item3"}
	rcl.SetItems(testItems)

	if len(rcl.items) != len(testItems) {
		t.Errorf("Expected %d items, got %d", len(testItems), len(rcl.items))
	}

	for i, item := range testItems {
		if rcl.items[i] != item {
			t.Errorf("Expected item[%d] = %s, got %s", i, item, rcl.items[i])
		}
	}
}

func TestRightClickableList_SetItemsEmpty(t *testing.T) {
	rcl := NewRightClickableList()

	// Set some items first
	rcl.SetItems([]string{"item1", "item2"})

	// Then set empty slice
	rcl.SetItems([]string{})

	if len(rcl.items) != 0 {
		t.Errorf("Expected empty items slice, got length %d", len(rcl.items))
	}
}

func TestRightClickableList_SetItemsNil(t *testing.T) {
	rcl := NewRightClickableList()

	// Set nil slice
	rcl.SetItems(nil)

	if rcl.items != nil {
		t.Error("Expected items to be nil")
	}
}

func TestRightClickableList_Build(t *testing.T) {
	rcl := NewRightClickableList()
	rcl.SetItems([]string{"item1", "item2", "item3"})

	// Build should create the internal list
	rcl.Build()

	if rcl.list == nil {
		t.Error("Build() did not create internal list")
	}
}

func TestRightClickableList_GetList(t *testing.T) {
	rcl := NewRightClickableList()
	rcl.Build()

	list := rcl.GetList()
	if list == nil {
		t.Error("GetList() returned nil")
	}

	if list != rcl.list {
		t.Error("GetList() returned different list than internal list")
	}
}

func TestRightClickableList_GetListBeforeBuild(t *testing.T) {
	rcl := NewRightClickableList()

	list := rcl.GetList()
	if list != nil {
		t.Error("GetList() should return nil before Build() is called")
	}
}

func TestRightClickableList_CreateRenderer(t *testing.T) {
	rcl := NewRightClickableList()
	rcl.Build()

	renderer := rcl.CreateRenderer()
	if renderer == nil {
		t.Error("CreateRenderer returned nil")
	}
}

func TestRightClickableList_Refresh(t *testing.T) {
	rcl := NewRightClickableList()
	rcl.SetItems([]string{"item1", "item2"})
	rcl.Build()

	// Should not panic
	rcl.Refresh()
}

func TestRightClickableList_RefreshBeforeBuild(t *testing.T) {
	rcl := NewRightClickableList()

	// Should not panic even if list is nil
	rcl.Refresh()
}

func TestRightClickableList_UnselectAll(t *testing.T) {
	rcl := NewRightClickableList()
	rcl.SetItems([]string{"item1", "item2"})
	rcl.Build()

	// Should not panic
	rcl.UnselectAll()
}

func TestRightClickableList_UnselectAllBeforeBuild(t *testing.T) {
	rcl := NewRightClickableList()

	// Should not panic even if list is nil
	rcl.UnselectAll()
}

func TestRightClickableList_CallbacksSetup(t *testing.T) {
	var tappedIndex int
	var rightClickedIndex int
	var rightClickedPos fyne.Position
	var tappedCalled bool
	var rightClickedCalled bool

	rcl := NewRightClickableList()

	rcl.OnItemTapped = func(index int) {
		tappedIndex = index
		tappedCalled = true
	}

	rcl.OnItemRightClick = func(index int, pos fyne.Position) {
		rightClickedIndex = index
		rightClickedPos = pos
		rightClickedCalled = true
	}

	if rcl.OnItemTapped == nil {
		t.Error("OnItemTapped callback was not set")
	}

	if rcl.OnItemRightClick == nil {
		t.Error("OnItemRightClick callback was not set")
	}

	// Test callbacks work
	rcl.OnItemTapped(5)
	if !tappedCalled || tappedIndex != 5 {
		t.Error("OnItemTapped callback did not work correctly")
	}

	testPos := fyne.NewPos(100, 200)
	rcl.OnItemRightClick(3, testPos)
	if !rightClickedCalled || rightClickedIndex != 3 || rightClickedPos != testPos {
		t.Error("OnItemRightClick callback did not work correctly")
	}
}

func TestRightClickableList_InterfaceCompliance(t *testing.T) {
	rcl := NewRightClickableList()

	// Test that RightClickableList implements fyne.Widget
	var _ fyne.Widget = rcl
}

func TestRightClickableList_WithTestApp(t *testing.T) {
	testApp := test.NewApp()
	defer testApp.Quit()

	rcl := NewRightClickableList()
	rcl.SetItems([]string{"Test Item 1", "Test Item 2", "Test Item 3"})
	rcl.Build()

	testWindow := testApp.NewWindow("Test")
	testWindow.SetContent(rcl)

	// Test that the widget can be displayed
	if rcl.list == nil {
		t.Error("Internal list should be created after Build()")
	}
}

func TestRightClickableList_SetItemsAfterBuild(t *testing.T) {
	rcl := NewRightClickableList()
	rcl.SetItems([]string{"item1", "item2"})
	rcl.Build()

	// Change items after build
	newItems := []string{"new1", "new2", "new3"}
	rcl.SetItems(newItems)

	if len(rcl.items) != len(newItems) {
		t.Errorf("Expected %d items after update, got %d", len(newItems), len(rcl.items))
	}

	for i, item := range newItems {
		if rcl.items[i] != item {
			t.Errorf("Expected item[%d] = %s, got %s", i, item, rcl.items[i])
		}
	}
}

func TestRightClickableList_LargeItemList(t *testing.T) {
	rcl := NewRightClickableList()

	// Test with large number of items
	largeItems := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeItems[i] = "Item " + string(rune('0'+i%10))
	}

	rcl.SetItems(largeItems)

	if len(rcl.items) != 1000 {
		t.Errorf("Expected 1000 items, got %d", len(rcl.items))
	}

	rcl.Build()
	if rcl.list == nil {
		t.Error("Build() should work with large item lists")
	}
}

func TestRightClickableList_UnicodeItems(t *testing.T) {
	rcl := NewRightClickableList()

	unicodeItems := []string{
		"测试项目1",
		"テストアイテム2",
		"🚀 Rocket Item",
		"Ñoño español",
		"Русский текст",
	}

	rcl.SetItems(unicodeItems)

	if len(rcl.items) != len(unicodeItems) {
		t.Errorf("Expected %d unicode items, got %d", len(unicodeItems), len(rcl.items))
	}

	for i, item := range unicodeItems {
		if rcl.items[i] != item {
			t.Errorf("Expected unicode item[%d] = %s, got %s", i, item, rcl.items[i])
		}
	}
}
