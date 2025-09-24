package appui

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

func TestNewItemContainer(t *testing.T) {
	onTapped := func(index int) {
		// Test callback
	}

	onRightClicked := func(index int, pos fyne.Position) {
		// Test callback
	}

	ic := NewItemContainer(onTapped, onRightClicked)

	if ic == nil {
		t.Fatal("NewItemContainer returned nil")
	}

	if ic.label == nil {
		t.Error("ItemContainer label is nil")
	}

	if ic.onTapped == nil {
		t.Error("ItemContainer onTapped callback is nil")
	}

	if ic.onRightClicked == nil {
		t.Error("ItemContainer onRightClicked callback is nil")
	}
}

func TestItemContainer_SetText(t *testing.T) {
	ic := NewItemContainer(nil, nil)

	testText := "Test Item"
	ic.SetText(testText)

	if ic.label.Text != testText {
		t.Errorf("Expected label text %s, got %s", testText, ic.label.Text)
	}
}

func TestItemContainer_SetIndex(t *testing.T) {
	ic := NewItemContainer(nil, nil)

	testIndex := 42
	ic.SetIndex(testIndex)

	if ic.index != testIndex {
		t.Errorf("Expected index %d, got %d", testIndex, ic.index)
	}
}

func TestItemContainer_Tapped(t *testing.T) {
	var tappedIndex int
	var callbackCalled bool

	onTapped := func(index int) {
		tappedIndex = index
		callbackCalled = true
	}

	ic := NewItemContainer(onTapped, nil)
	ic.SetIndex(5)

	// Simulate tap event
	pe := &fyne.PointEvent{
		Position:         fyne.NewPos(10, 10),
		AbsolutePosition: fyne.NewPos(100, 100),
	}

	ic.Tapped(pe)

	if !callbackCalled {
		t.Error("Tapped callback was not called")
	}

	if tappedIndex != 5 {
		t.Errorf("Expected tapped index 5, got %d", tappedIndex)
	}
}

func TestItemContainer_TappedSecondary(t *testing.T) {
	var rightClickedIndex int
	var rightClickedPos fyne.Position
	var callbackCalled bool

	onRightClicked := func(index int, pos fyne.Position) {
		rightClickedIndex = index
		rightClickedPos = pos
		callbackCalled = true
	}

	ic := NewItemContainer(nil, onRightClicked)
	ic.SetIndex(3)

	// Simulate right-click event
	expectedPos := fyne.NewPos(200, 150)
	pe := &fyne.PointEvent{
		Position:         fyne.NewPos(20, 15),
		AbsolutePosition: expectedPos,
	}

	ic.TappedSecondary(pe)

	if !callbackCalled {
		t.Error("TappedSecondary callback was not called")
	}

	if rightClickedIndex != 3 {
		t.Errorf("Expected right-clicked index 3, got %d", rightClickedIndex)
	}

	if rightClickedPos != expectedPos {
		t.Errorf("Expected right-clicked position %v, got %v", expectedPos, rightClickedPos)
	}
}

func TestItemContainer_TappedWithNilCallback(t *testing.T) {
	ic := NewItemContainer(nil, nil)
	ic.SetIndex(1)

	// Should not panic when callbacks are nil
	pe := &fyne.PointEvent{
		Position:         fyne.NewPos(10, 10),
		AbsolutePosition: fyne.NewPos(100, 100),
	}

	// These should not panic
	ic.Tapped(pe)
	ic.TappedSecondary(pe)
}

func TestItemContainer_CreateRenderer(t *testing.T) {
	ic := NewItemContainer(nil, nil)

	renderer := ic.CreateRenderer()
	if renderer == nil {
		t.Error("CreateRenderer returned nil")
	}
}

func TestItemContainer_InterfaceCompliance(t *testing.T) {
	ic := NewItemContainer(nil, nil)

	// Test that ItemContainer implements required interfaces
	var _ fyne.Widget = ic
	var _ fyne.Tappable = ic
	var _ fyne.SecondaryTappable = ic
}

func TestItemContainer_WithTestApp(t *testing.T) {
	// Create a test app for more complete testing
	testApp := test.NewApp()
	defer testApp.Quit()

	var tappedIndex int
	onTapped := func(index int) {
		tappedIndex = index
	}

	ic := NewItemContainer(onTapped, nil)
	ic.SetText("Test Item")
	ic.SetIndex(7)

	// Test with test app
	testWindow := testApp.NewWindow("Test")
	testWindow.SetContent(ic)

	// Simulate tap
	test.Tap(ic)

	if tappedIndex != 7 {
		t.Errorf("Expected tapped index 7, got %d", tappedIndex)
	}
}
