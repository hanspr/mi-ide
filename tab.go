package main

import (
	"strings"

	"github.com/hanspr/tcell/v2"
)

const (
	bufDirty      string = "✸"
	tabOpen       string = "|"
	tabClose      string = "|"
	tabMenuSymbol string = "ɱ  "
)

var tabBarOffset int
var toolBarOffset int

// -------------------------------------------
// Toolbar
// -------------------------------------------

// ToolBar structure for a toolbar
type ToolBar struct {
	active      bool
	icons       []rune
	callback    []func()
	luacallback string
}

// NewToolBar create a new toolbar with predefined icons
func NewToolBar() *ToolBar {
	t := new(ToolBar)
	// t.AddIcon('⌹', t.DirView, "")
	// t.AddIcon('↴', t.Save, "")
	// t.AddIcon('◨', t.VSplit, "")
	// t.AddIcon('⬓', t.HSplit, "")
	// t.AddIcon('🔎', t.Find, "")
	// t.AddIcon('℁', t.Replace, "")
	// t.AddIcon('↑', t.CloudUpload, "")
	// t.AddIcon('↓', t.CloudDownload, "")
	// t.AddIcon('❎', t.Quit, "")
	t.active = true
	return t
}

// AddIcon adds an icon to the toolbar
func (t *ToolBar) AddIcon(icon rune, cb func(), luacallback string) {
	if cb == nil {
		if luacallback == "" {
			return
		}
		n := len(t.icons) - 1
		cb = t.PluginCallBack
		t.icons = append(t.icons, ' ')
		copy(t.icons[n+1:], t.icons[n:])
		t.icons[n] = icon
		t.callback = append(t.callback, nil)
		copy(t.callback[n+1:], t.callback[n:])
		t.callback[n] = cb
		t.luacallback = luacallback
	} else {
		t.icons = append(t.icons, icon)
		t.callback = append(t.callback, cb)
	}
}

// Runes convert toolbar into runes
func (t *ToolBar) Runes() []rune {
	var str []rune

	toolBarOffset = len(t.icons)*3 + 3
	for _, ic := range t.icons {
		str = append(str, ' ', ic, ' ')
	}
	str = append([]rune(tabMenuSymbol), str...)
	return str
}

// PluginCallBack call back function for plugins
func (t *ToolBar) PluginCallBack() {
	Call(t.luacallback)
}

// Save Save icon callback
// func (t *ToolBar) Save() {
// 	CurView().Buf.Save()
// }

// VSplit icon callback
// func (t *ToolBar) VSplit() {
// 	CurView().VSplitBinding(true)
// }

// HSplit icon callback
// func (t *ToolBar) HSplit() {
// 	CurView().HSplitBinding(true)
// }

// Find icon callback
// func (t *ToolBar) Find() {
// 	CurView().FindDialog(true)
// }

// Replace icon callback
// func (t *ToolBar) Replace() {
// 	CurView().SearchDialog(true)
// }

// DirView icon callback
// func (t *ToolBar) DirView() {
// 	micromenu.DirTreeView()
// }

// Void unimplemented
func (t *ToolBar) Void() {
}

// Quit icon callback
// func (t *ToolBar) Quit() {
// 	CurView().Quit(true)
// }

// CloudUpload icon callback
// func (t *ToolBar) CloudUpload() {
// 	CurView().UploadToCloud(false)
// }

// CloudDownload icon callback
// func (t *ToolBar) CloudDownload() {
// 	CurView().DownloadFromCloud(false)
// }

// ToolbarHandleMouseEvent Handle ToolBarClick
func (t *ToolBar) ToolbarHandleMouseEvent(x int) {
	var pos int

	if !t.active {
		return
	}
	pos = x / 3
	t.callback[pos-1]()
}

// FixTabsIconArea fix the icon area filling with black background to avoid wide chars leaving transparent cells
func (t *ToolBar) FixTabsIconArea() {
	// prefill the length with with black background spaces
	var tStyle tcell.Style

	if !t.active {
		return
	}
	tStyle = StringToStyle("normal #000000,#000000")
	for x := 0; x < toolBarOffset; x++ {
		screen.SetContent(x, 0, '.', nil, tStyle)
	}
	screen.Show()
}

// -------------------------------------------
// Tabs
// -------------------------------------------

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
	t.tree.height = h - 1

	//t.tree.height--

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
	t.tree.height = h - 1

	t.tree.ResizeSplits()

	for i, v := range t.Views {
		v.Num = i
	}
	MicroToolBar.FixTabsIconArea()
}

// CurView returns the current view
func CurView() *View {
	if len(tabs) == 0 {
		return nil
	}
	curtab := tabs[curTab]
	return curtab.Views[curtab.CurView]
}

// TabbarHandleMouseEvent checks the given mouse event if it is clicking on the tabbar
// If it is it changes the current tab accordingly
// This function returns true if the tab is changed
func TabbarHandleMouseEvent(event tcell.Event) {
	var tabnum int

	toffset := toolBarOffset
	if !MicroToolBar.active {
		toffset = 0
	}
	if Mouse.Button == 1 || Mouse.Button == 3 {
		x := Mouse.Pos.X
		if toffset > 0 {
			if x < 3 {
				// Click on Menu Icon
				micromenu.Menu()
				return
			}
			if x < toffset {
				// Click on Toolbar
				if Mouse.Button == 1 {
					MicroToolBar.ToolbarHandleMouseEvent(x)
				}
				return
			}
		}
		// Get the indices to know the hotspot for each tab
		str, indicies := TabbarString(toffset)
		// ignore if past last tab
		if x >= Count(str)+toffset {
			return
		}
		// Find which tab was clicked
		for i := range tabs {
			if x+tabBarOffset < indicies[i]+toffset {
				tabnum = i
				break
			}
		}
		// Ignore on current tab and not to close
		if Mouse.Button == 1 && curTab == tabnum {
			return
		}
		// Change tab
		if Mouse.Button == 1 {
			// Left click = Select tab and display
			curTab = tabnum
			return
		} else if Mouse.Button == 3 {
			// We allow to close only the current Tab, so user knows what he is doing
			if curTab == tabnum {
				// Right click = Close selected tab
				CurView().Unsplit(false)
				CurView().Quit(false)
			} else {
				// If not current, make current so he can click again
				curTab = tabnum
				return
			}
		}
		return
	}
}

// -------------------------------------------
// Tabbar drawing
// -------------------------------------------

// TabbarString returns the string that should be displayed in the tabbar
// It also returns a map containing which indicies correspond to which tab number
// This is useful when we know that the mouse click has occurred at an x location
func TabbarString(toffset int) (string, map[int]int) {
	var cv int

	w, _ := screen.Size()
	str := ""
	middle := 0
	tabIndex := 0 // Reference to the first tab being displayed
	curTabDisplayed := false
	indicies := make(map[int]int)

	for i, t := range tabs {
		buf := t.Views[t.CurView].Buf
		name := buf.Fname
		if Count(str)+Count(name)+toffset+10 > w {
			// Tabbar is longer than screen width
			if middle == 0 {
				middle = (i - tabIndex) / 2
			}
			if curTab > middle {
				// We are past the midfdle of the tabbar, we have to shift everything to the left
				// to leave some root to see next available tabs
				offset := indicies[tabIndex]
				olen := Count(str)
				str = "←" + str[offset+1:]
				diff := olen - Count(str)
				for a := range tabs {
					if a < tabIndex {
						indicies[a] = 0
						continue
					} else if a > i {
						break
					}
					indicies[a] -= diff
				}
				tabIndex++
			}
			if curTabDisplayed && curTab-tabIndex <= middle {
				// We have the current tab on view and in the middle, no need to do more processing
				str = str + "→ "
				indicies[i] = Count(str)
				break
			}
		}
		cv = t.CurView

		if i == curTab {
			str += tabOpen + " "
			curTabDisplayed = true
		} else {
			str += "  "
		}
		str += buf.Fname
		if t.Views[cv].Type.Kind == 0 && buf.Modified() {
			str += bufDirty
		} else {
			str += " "
		}
		if i == curTab {
			str += tabClose
		} else {
			str += " "
		}
		indicies[i] = Count(str)
	}
	return strings.Clone(str), indicies
}

// DisplayTabs Display a Menu, Icons ToolBar, and Tabs
func DisplayTabs() {
	var tStyle tcell.Style
	var tabActive = false

	// Display ToolBar
	w, _ := screen.Size()
	toffset := 0
	if w > 60 {
		MicroToolBar.active = true
		toffset = toolBarOffset
		toolbarRunes := MicroToolBar.Runes()
		for x := 0; x < len(toolbarRunes); x++ {
			if x < 3 {
				// Menu icon
				if x == 0 {
					tStyle = StringToStyle("bold #ffd700,#121212")
				}
			} else if x < toffset {
				// Toolbar icons
				if x == 3 {
					tStyle = StringToStyle("#ffffff")
				}
			}
			screen.SetContent(x, 0, toolbarRunes[x], nil, tStyle)
		}
	} else {
		MicroToolBar.active = false
	}

	// Get the current Tabs to display
	str, _ := TabbarString(toffset)

	tabBarStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["tabbar"]; ok {
		tabBarStyle = style
	}
	tabsRunes := []rune(str)

	// Display Tabs
	for x := 0; x < w; x++ {
		if x < len(tabsRunes) {
			if string(tabsRunes[x]) == tabOpen && !tabActive {
				// Hightlight the current tab
				tStyle = StringToStyle("bold #ffd700,#000087")
				tabActive = true
			} else if tabActive {
				tStyle = StringToStyle("bold #ffd700,#000087")
				if string(tabsRunes[x]) == bufDirty {
					tStyle = tStyle.Foreground(tcell.ColorRed)
				}
				if string(tabsRunes[x]) == tabClose {
					// Hightlight off, end of current tab
					tabActive = false
				}
			} else {
				tStyle = tabBarStyle
				if string(tabsRunes[x]) == bufDirty {
					tStyle = tStyle.Foreground(tcell.ColorRed)
				}
			}
			screen.SetContent(x+toffset, 0, tabsRunes[x], nil, tStyle)
		} else {
			screen.SetContent(x+toffset, 0, ' ', nil, tabBarStyle)
		}
	}
}

// NewTab opens the given file in a new tab
func NewTab(args []string) {
	if len(args) == 0 {
		CurView().AddTab(true)
	} else {
		buf, err := NewBufferFromFile(args[0])
		if err != nil {
			messenger.Alert("error", err)
			return
		}

		tab := NewTabFromView(NewView(buf))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		curTab = len(tabs) - 1
		if len(tabs) == 2 {
			for _, t := range tabs {
				for _, v := range t.Views {
					v.AddTabbarSpace()
				}
			}
		}
	}
}
