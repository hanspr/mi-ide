package main

import (
	"path/filepath"
	"sort"

	"github.com/hanspr/tcell"
)

const (
	tabOpen       string = "|"
	tabClose      string = "|"
	tabMenuSymbol string = "µ𝑒 "
	toolBar       string = " ↴  ◨  ⬓  🔎  ℁  ❎ "
)

var tabBarOffset int

// A Tab holds an array of views and a splitTree to determine how the
// views should be arranged
type Tab struct {
	// This contains all the views in this tab
	// There is generally only one view per tab, but you can have
	// multiple views with splits
	Views []*View
	// This is the current view for this tab
	CurView int

	tree *SplitTree
}

// NewTabFromView creates a new tab and puts the given view in the tab
func NewTabFromView(v *View) *Tab {
	t := new(Tab)
	t.Views = append(t.Views, v)
	t.Views[0].Num = 0

	t.tree = new(SplitTree)
	t.tree.kind = VerticalSplit
	t.tree.children = []Node{NewLeafNode(t.Views[0], t.tree)}

	w, h := screen.Size()
	t.tree.width = w
	t.tree.height = h

	if globalSettings["infobar"].(bool) {
		t.tree.height--
	}

	t.Resize()
	return t
}

// SetNum sets all this tab's views to have the correct tab number
func (t *Tab) SetNum(num int) {
	t.tree.tabNum = num
	for _, v := range t.Views {
		v.TabNum = num
	}
}

// Cleanup cleans up the tree (for example if views have closed)
func (t *Tab) Cleanup() {
	t.tree.Cleanup()
}

// Resize handles a resize event from the terminal and resizes
// all child views correctly
func (t *Tab) Resize() {
	w, h := screen.Size()
	t.tree.width = w
	t.tree.height = h

	if globalSettings["infobar"].(bool) {
		t.tree.height--
	}

	t.tree.ResizeSplits()

	for i, v := range t.Views {
		v.Num = i
		if v.Type == vtTerm {
			v.term.Resize(v.Width, v.Height)
		}
	}
}

// CurView returns the current view
func CurView() *View {
	curTab := tabs[curTab]
	return curTab.Views[curTab.CurView]
}

// TabbarString returns the string that should be displayed in the tabbar
// It also returns a map containing which indicies correspond to which tab number
// This is useful when we know that the mouse click has occurred at an x location
// but need to know which tab that corresponds to to accurately change the tab
func TabbarString() (string, map[int]int) {
	var cv int

	str := tabMenuSymbol + toolBar
	indicies := make(map[int]int)
	unique := make(map[string]int)

	for _, t := range tabs {
		unique[filepath.Base(t.Views[t.CurView].Buf.GetName())]++
	}

	for i, t := range tabs {
		buf := t.Views[t.CurView].Buf
		name := filepath.Base(buf.GetName())
		if name == "fileviewer" {
			buf = t.Views[t.CurView+1].Buf
			name = filepath.Base(buf.GetName())
			cv = t.CurView + 1
		} else {
			cv = t.CurView
		}

		if i == curTab {
			str += tabOpen + " "
		} else {
			str += "  "
		}
		if unique[name] == 1 {
			str += name
		} else {
			str += buf.GetName()
		}
		if t.Views[cv].Type.Kind == 0 && buf.Modified() {
			str += "*"
		}
		if i == curTab {
			str += " " + tabClose
		} else {
			str += "  "
		}
		//		indicies[Count(str)+1] = i + 1
		indicies[Count(str)] = i + 1
	}
	return str, indicies
}

// TabbarHandleMouseEvent checks the given mouse event if it is clicking on the tabbar
// If it is it changes the current tab accordingly
// This function returns true if the tab is changed
func TabbarHandleMouseEvent(event tcell.Event) bool {
	var tabnum int
	var keys []int

	// For the moment, ignore mouse events if only one file open in tab
	// When menu is added will be removed
	//if len(tabs) <= 1 {
	//	return false
	//}

	switch e := event.(type) {
	case *tcell.EventMouse:
		button := e.Buttons()
		// Accepted Mouse Click Events
		if button == tcell.Button1 || button == tcell.Button3 {
			// Find where is the mouse click
			x, y := e.Position()
			// ignore if not in tab line
			if y != 0 {
				return false
			}
			if e.HasMotion() == true {
				return true
			}
			str, indicies := TabbarString()
			// ignore if past last tab
			if x+tabBarOffset >= Count(str) {
				return true
			}
			if x < 3 {
				// Click on Menu Icon
				//micromenu.Start()
				micromenu.Test()
				return true
			}
			if x < 21 {
				// Click on Toolbar
				if button == tcell.Button1 {
					ToolbarHandleMouseEvent(x)
				}
				return true
			}
			// Find which tab was clicked
			for k := range indicies {
				keys = append(keys, k)
			}
			sort.Ints(keys)
			for _, k := range keys {
				if x+tabBarOffset <= k {
					tabnum = indicies[k] - 1
					break
				}
			}
			// Ignore on current tab and not to close
			if button == tcell.Button1 && curTab == tabnum {
				return true
			}
			// Change tab
			dirview.onTabChange()
			if button == tcell.Button1 {
				// Left click = Select tab and display
				curTab = tabnum
				return true
			} else if button == tcell.Button3 {
				// We allow to close only the current Tab, so user knows what he is doing
				if curTab == tabnum {
					// Right click = Close selected tab
					CurView().Unsplit(false)
					CurView().Quit(false)
				} else {
					// If not current, make current so he can click again
					curTab = tabnum
					return true
				}
			}
			return true
		}
	}
	return false
}

// Handle ToolBarClick
// Save, Split Vertical, Split Horizontal, Find, Search, Close selected View
func ToolbarHandleMouseEvent(x int) {
	v := CurView()
	switch {
	case x < 6:
		// Save
		v.Buf.Save()
	case x < 9:
		// Split Vertical
		v.VSplitBinding(true)
	case x < 12:
		// Split Horizontal
		v.HSplitBinding(true)
	case x < 15:
		// Search
		v.FindDialog(true)
	case x < 18:
		//Search Replace
		v.SearchDialog(true)
	default:
		//Close View
		v.SafeQuit(true)
	}
}

func FixTabsIconArea() {
	// prefill the first 21 spaces with black background
	var tStyle tcell.Style
	tStyle = StringToStyle("normal #000000,#000000")
	for x := 0; x < 22; x++ {
		screen.SetContent(x, 0, '.', nil, tStyle)
	}
	screen.Show()
}

// DisplayTabs; always displays the tabbar at the top of the editor
// The bar will eventually handle a menu
func DisplayTabs() {
	var tStyle tcell.Style
	var tabActive bool = false

	str, indicies := TabbarString()

	tabBarStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["tabbar"]; ok {
		tabBarStyle = style
	}
	// Maybe there is a unicode filename?
	fileRunes := []rune(str)
	w, _ := screen.Size()
	tooWide := (w < len(fileRunes))

	// if the entire tab-bar is longer than the screen is wide,
	// then it should be truncated appropriately to keep the
	// active tab visible on the UI.
	if tooWide == true {
		// first we have to work out where the selected tab is
		// out of the total length of the tab bar. this is done
		// by extracting the hit-areas from the indicies map
		// that was constructed by `TabbarString()`
		var keys []int
		for offset := range indicies {
			keys = append(keys, offset)
		}
		// sort them to be in ascending order so that values will
		// correctly reflect the displayed ordering of the tabs
		sort.Ints(keys)
		// record the offset of each tab and the previous tab so
		// we can find the position of the tab's hit-box.
		previousTabOffset := 0
		currentTabOffset := 0
		for _, k := range keys {
			tabIndex := indicies[k] - 1
			if tabIndex == curTab {
				currentTabOffset = k
				break
			}
			// this is +2 because there are two padding spaces that aren't accounted
			// for in the display. please note that this is for cosmetic purposes only.
			previousTabOffset = k + 2
		}
		// get the width of the hitbox of the active tab, from there calculate the offsets
		// to the left and right of it to approximately center it on the tab bar display.
		centeringOffset := (w - (currentTabOffset - previousTabOffset))
		leftBuffer := previousTabOffset - (centeringOffset / 2)
		rightBuffer := currentTabOffset + (centeringOffset / 2)

		// check to make sure we haven't overshot the bounds of the string,
		// if we have, then take that remainder and put it on the left side
		overshotRight := rightBuffer - len(fileRunes)
		if overshotRight > 0 {
			leftBuffer = leftBuffer + overshotRight
		}

		overshotLeft := leftBuffer - 0
		if overshotLeft < 0 {
			leftBuffer = 0
			rightBuffer = leftBuffer + (w - 1)
		} else {
			rightBuffer = leftBuffer + (w - 2)
		}

		if rightBuffer > len(fileRunes)-1 {
			rightBuffer = len(fileRunes) - 1
		}

		// construct a new buffer of text to put the
		// newly formatted tab bar text into.
		var displayText []rune

		// if the left-side of the tab bar isn't at the start
		// of the constructed tab bar text, then show that are
		// more tabs to the left by displaying a "+"
		if leftBuffer != 0 {
			displayText = append(displayText, '+')
		}
		// copy the runes in from the original tab bar text string
		// into the new display buffer
		for x := leftBuffer; x < rightBuffer; x++ {
			displayText = append(displayText, fileRunes[x])
		}
		// if there is more text to the right of the right-most
		// column in the tab bar text, then indicate there are more
		// tabs to the right by displaying a "+"
		if rightBuffer < len(fileRunes)-1 {
			displayText = append(displayText, '+')
		}

		// now store the offset from zero of the left-most text
		// that is being displayed. This is to ensure that when
		// clicking on the tab bar, the correct tab gets selected.
		tabBarOffset = leftBuffer

		// use the constructed buffer as the display buffer to print
		// onscreen.
		fileRunes = displayText
	} else {
		tabBarOffset = 0
	}

	// iterate over the width of the terminal display and for each column,
	// write a character into the tab display area with the appropriate style.
	for x := 0; x < w; x++ {
		if x < len(fileRunes) {
			if x < 3 {
				// Menu icon
				if x == 0 {
					tStyle = StringToStyle("bold #ffd700,#121212")
				}
			} else if x < 21 {
				// Toolbar icons
				if x == 3 {
					tStyle = StringToStyle("#ffffff")
				}
			} else if string(fileRunes[x]) == tabOpen && tabActive == false {
				tStyle = StringToStyle("bold #ffd700,#000087")
				tabActive = true
			} else if tabActive {
				if string(fileRunes[x]) == tabClose {
					tabActive = false
				}
			} else {
				tStyle = tabBarStyle
			}
			screen.SetContent(x, 0, fileRunes[x], nil, tStyle)
		} else {
			screen.SetContent(x, 0, ' ', nil, tabBarStyle)
		}
	}
}
