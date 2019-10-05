package main

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hanspr/tcell"
)

// The ViewType defines what kind of view this is
type ViewType struct {
	Kind     int
	Readonly bool // The file cannot be edited
	Scratch  bool // The file cannot be saved
}

var (
	vtDefault = ViewType{0, false, false}
	vtHelp    = ViewType{1, true, true}
	vtLog     = ViewType{2, true, true}
	vtScratch = ViewType{3, false, true}
	vtRaw     = ViewType{4, true, true}
	vtTerm    = ViewType{5, true, true}
)

var LastView int = -1
var LastTab int = -1

// The View struct stores information about a view into a buffer.
// It stores information about the cursor, and the viewport
// that the user sees the buffer from.
type View struct {
	// A pointer to the buffer's cursor for ease of access
	Cursor *Cursor

	// The topmost line, used for vertical scrolling
	Topline int
	// The leftmost column, used for horizontal scrolling
	leftCol int

	// Specifies whether or not this view holds a help buffer
	Type ViewType

	// Actual width and height
	Width  int
	Height int

	LockWidth  bool
	LockHeight bool

	// Where this view is located
	x, y int

	// How much to offset because of line numbers
	lineNumOffset int

	// Holds the list of gutter messages
	messages map[string][]GutterMessage

	// This is the index of this view in the views array
	Num int
	// What tab is this view stored in
	TabNum int

	// The buffer
	Buf *Buffer
	// The statusline
	sline *Statusline

	// Since tcell doesn't differentiate between a mouse release event
	// and a mouse move event with no keys pressed, we need to keep
	// track of whether or not the mouse was pressed (or not released) last event to determine
	// mouse release events
	mouseReleased bool

	// We need to keep track of insert key press toggle
	isOverwriteMode bool
	// This stores when the last click was
	// This is useful for detecting double and triple clicks
	lastClickTime time.Time

	// Saved cursor locations for view jumps and search
	lastLoc    Loc
	savedLoc   Loc
	savedLine  string
	searchSave Loc

	// lastCutTime stores when the last ctrl+k was issued.
	// It is used for clearing the clipboard to replace it with fresh cut lines.
	lastCutTime time.Time

	// freshClip returns true if the clipboard has never been pasted.
	freshClip bool

	// Was the last mouse event actually a double click?
	// Useful for detecting triple clicks -- if a double click is detected
	// but the last mouse event was actually a double click, it's a triple click
	doubleClick bool
	// Same here, just to keep track for mouse move events
	tripleClick bool

	// The cellview used for displaying and syntax highlighting
	cellview *CellView

	splitNode *LeafNode

	// Virtual terminal
	term *Terminal
}

// NewView returns a new fullscreen view
func NewView(buf *Buffer) *View {
	screenW, screenH := screen.Size()
	return NewViewWidthHeight(buf, screenW, screenH)
}

// NewViewWidthHeight returns a new view with the specified width and height
// Note that w and h are raw column and row values
func NewViewWidthHeight(buf *Buffer, w, h int) *View {
	v := new(View)

	v.x, v.y = 0, 0

	v.Width = w
	v.Height = h
	v.cellview = new(CellView)

	v.AddTabbarSpace()

	v.OpenBuffer(buf)

	v.messages = make(map[string][]GutterMessage)

	v.sline = &Statusline{
		view:    v,
		hotspot: make(map[string]Loc),
	}

	v.term = new(Terminal)

	for pl := range loadedPlugins {
		if GetPluginOption(pl, "ftype") != "*" && (GetPluginOption(pl, "ftype") == nil || GetPluginOption(pl, "ftype").(string) != CurView().Buf.FileType()) {
			continue
		}
		_, err := Call(pl+".onViewOpen", v)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
	}

	if v.Type.Kind == 0 && v.Buf.RO {
		v.Type.Readonly = true
	}
	return v
}

// StartTerminal execs a command in this view
func (v *View) StartTerminal(execCmd []string, wait bool, getOutput bool, luaCallback string) error {
	err := v.term.Start(execCmd, v, getOutput)
	v.term.wait = wait
	v.term.callback = luaCallback
	if err == nil {
		v.term.Resize(v.Width, v.Height)
		v.Type = vtTerm
	}
	return err
}

// CloseTerminal shuts down the tty running in this view
// and returns it to the default view type
func (v *View) CloseTerminal() {
	v.term.Stop()
}

// AddTabbarSpace creates an extra row for the tabbar if necessary
func (v *View) AddTabbarSpace() {
	if v.y == 0 {
		// Include one line for the tab bar at the top
		v.Height--
		v.y = 1
	}
}

func (v *View) paste(clip string) {

	if len(clip) == 0 {
		return
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	Start := v.Cursor.Loc
	v.Buf.pasteLoc = v.Cursor.Loc

	if v.Buf.Settings["rmtrailingws"] == true {
		v.Buf.RemoveTrailingSpace(v.Cursor.Loc)
	}
	v.Cursor.Loc = Start

	if v.Buf.Settings["smartindent"].(bool) {
		v.Buf.Insert(v.Cursor.Loc, clip)
		x := v.Cursor.Loc.X
		spc := CountLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
		v.Buf.SmartIndent(Start, v.Cursor.Loc, false)
		if strings.Contains(clip, "\n") {
			// Multiline paste, move cursor to begging of last line, is the safest option
			v.Cursor.GotoLoc(Loc{0, v.Cursor.Loc.Y})
		} else {
			// One line paste, check if it was indented
			spc2 := CountLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
			if spc != spc2 {
				// Was indented move cursor location to pasted end string
				v.Cursor.GotoLoc(Loc{x + spc2, v.Cursor.Loc.Y})
			}
		}
	} else {
		if v.Buf.Settings["smartpaste"].(bool) {
			if v.Cursor.X > 0 && GetLeadingWhitespace(strings.TrimLeft(clip, "\r\n")) == "" {
				leadingWS := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
				clip = strings.Replace(clip, "\n", "\n"+leadingWS, -1)
			}
		}
		v.Buf.Insert(v.Cursor.Loc, clip)
	}

	v.freshClip = false
}

// ScrollUp scrolls the view up n lines (if possible)
func (v *View) ScrollUp(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.Topline-n >= 0 {
		v.Topline -= n
	} else if v.Topline > 0 {
		v.Topline--
	}
}

// ScrollDown scrolls the view down n lines (if possible)
func (v *View) ScrollDown(n int) {
	// Try to scroll by n but if it would overflow, scroll by 1
	if v.Topline+n <= v.Buf.NumLines {
		v.Topline += n
	} else if v.Topline < v.Buf.NumLines-1 {
		v.Topline++
	}
}

// CanClose returns whether or not the view can be closed
// If there are unsaved changes, the user will be asked if the view can be closed
// causing them to lose the unsaved changes
func (v *View) CanClose() bool {
	// Check if another clone is open, if it is, close without saving
	for _, t := range tabs {
		for _, view := range t.Views {
			if v.Buf.AbsPath == view.Buf.AbsPath && v.Num != view.Num {
				return true
			} else if v.Buf.AbsPath == view.Buf.AbsPath && v.TabNum != view.TabNum {
				return true
			}
		}
	}
	if v.Type == vtDefault && v.Buf.Modified() {
		var choice bool
		var canceled bool
		if v.Buf.Settings["autosave"].(bool) {
			choice = true
		} else {
			choice, canceled = messenger.YesNoPrompt(Language.Translate("Save changes to") + " " + v.Buf.GetName() + " " + Language.Translate("before closing? (y,n,esc)"))
		}
		if !canceled {
			//if char == 'y' {
			if choice {
				v.Save(true)
			}
		} else {
			return false
		}
	}
	return true
}

// OpenBuffer opens a new buffer in this view.
// This resets the topline, event handler and cursor.
func (v *View) OpenBuffer(buf *Buffer) {
	screen.Clear()
	v.CloseBuffer()
	v.Buf = buf
	v.Cursor = &buf.Cursor
	v.Topline = 0
	v.leftCol = 0
	v.Cursor.ResetSelection()
	v.Relocate()
	v.Center(false)
	v.messages = make(map[string][]GutterMessage)

	// Set mouseReleased to true because we assume the mouse is not being pressed when
	// the editor is opened
	v.mouseReleased = true
	// Set isOverwriteMode to false, because we assume we are in the default mode when editor
	// is opened
	v.isOverwriteMode = false
	v.lastClickTime = time.Time{}

	GlobalPluginCall("onBufferOpen", v.Buf)
	GlobalPluginCall("onViewOpen", v)
}

// Open opens the given file in the view
func (v *View) Open(path string) {
	buf, err := NewBufferFromFile(path)
	if err != nil {
		messenger.Error(err)
		return
	}
	v.OpenBuffer(buf)
}

// CloseBuffer performs any closing functions on the buffer
func (v *View) CloseBuffer() {
	if v.Buf != nil {
		v.Buf.Serialize()
	}
}

// ReOpen reloads the current buffer
func (v *View) ReOpen() {
	if v.CanClose() {
		screen.Clear()
		v.Buf.ReOpen()
		v.Relocate()
	}
}

// HSplit opens a horizontal split with the given buffer
func (v *View) HSplit(buf *Buffer) {
	i := 0
	if v.Buf.Settings["splitbottom"].(bool) {
		i = 1
	}
	v.savedLoc = v.Cursor.Loc
	v.savedLine = SubtringSafe(v.Buf.Line(v.Cursor.Loc.Y), 0, 20)
	v.splitNode.HSplit(buf, CurView().Num+i)
}

// VSplit opens a vertical split with the given buffer
func (v *View) VSplit(buf *Buffer) {
	i := 0
	if v.Buf.Settings["splitright"].(bool) {
		i = 1
	}
	v.savedLoc = v.Cursor.Loc
	v.savedLine = SubtringSafe(v.Buf.Line(v.Cursor.Loc.Y), 0, 20)
	v.splitNode.VSplit(buf, CurView().Num+i)
}

// HSplitIndex opens a horizontal split with the given buffer at the given index
func (v *View) HSplitIndex(buf *Buffer, splitIndex int) {
	v.splitNode.HSplit(buf, splitIndex)
}

// VSplitIndex opens a vertical split with the given buffer at the given index
func (v *View) VSplitIndex(buf *Buffer, splitIndex int) {
	v.splitNode.VSplit(buf, splitIndex)
}

// GetSoftWrapLocation gets the location of a visual click on the screen and converts it to col,line
func (v *View) GetSoftWrapLocation(vx, vy int) (int, int) {
	if !v.Buf.Settings["softwrap"].(bool) {
		if vy >= v.Buf.NumLines {
			vy = v.Buf.NumLines - 1
		}
		vx = v.Cursor.GetCharPosInLine(vy, vx)
		return vx, vy
	}

	screenX, screenY := 0, v.Topline
	for lineN := v.Topline; lineN < v.Bottomline(); lineN++ {
		line := v.Buf.Line(lineN)
		if lineN >= v.Buf.NumLines {
			return 0, v.Buf.NumLines - 1
		}

		colN := 0
		for _, ch := range line {
			if screenX >= v.Width-v.lineNumOffset {
				screenX = 0
				screenY++
			}

			if screenX == vx && screenY == vy {
				return colN, lineN
			}

			if ch == '\t' {
				screenX += int(v.Buf.Settings["tabsize"].(float64)) - 1
			}

			screenX++
			colN++
		}
		if screenY == vy {
			return colN, lineN
		}
		screenX = 0
		screenY++
	}

	return 0, 0
}

// Bottomline returns the line number of the lowest line in the view
// You might think that this is obviously just v.Topline + v.Height
// but if softwrap is enabled things get complicated since one buffer
// line can take up multiple lines in the view
func (v *View) Bottomline() int {
	if !v.Buf.Settings["softwrap"].(bool) {
		return v.Topline + v.Height
	}

	screenX, screenY := 0, 0
	numLines := 0
	for lineN := v.Topline; lineN < v.Topline+v.Height; lineN++ {
		line := v.Buf.Line(lineN)

		colN := 0
		for _, ch := range line {
			if screenX >= v.Width-v.lineNumOffset {
				screenX = 0
				screenY++
			}

			if ch == '\t' {
				screenX += int(v.Buf.Settings["tabsize"].(float64)) - 1
			}

			screenX++
			colN++
		}
		screenX = 0
		screenY++
		numLines++

		if screenY >= v.Height {
			break
		}
	}
	return numLines + v.Topline
}

// Relocate moves the view window so that the cursor is in view, only if out of view
// This is useful if the user has scrolled far away, and then starts typing
func (v *View) Relocate() bool {
	scrollmargin := int(v.Buf.Settings["scrollmargin"].(float64))
	cy := v.Cursor.Y
	height := v.Bottomline() - v.Topline
	ret := false
	if cy < v.Topline+scrollmargin && cy > scrollmargin-1 {
		v.Topline = cy - scrollmargin
		ret = true
	} else if cy < v.Topline {
		v.Topline = cy
		ret = true
	}
	if cy > v.Topline+height-1-scrollmargin && cy < v.Buf.NumLines-scrollmargin {
		v.Topline = cy - height + 1 + scrollmargin
		ret = true
	} else if cy >= v.Buf.NumLines-scrollmargin && cy >= height {
		v.Topline = v.Buf.NumLines - height
		ret = true
	}

	if !v.Buf.Settings["softwrap"].(bool) {
		// HPR
		// Force go all the way to the left when visual X in range to fit at the begging
		// This avoids having shifted lines close to the begining of the window
		cx := v.Cursor.GetVisualX()
		if cx < v.Width*10/11 {
			cx = 0
		}
		if cx < v.leftCol {
			v.leftCol = cx
			ret = true
		}
		if cx+v.lineNumOffset+1 > v.leftCol+v.Width {
			v.leftCol = cx - v.Width + v.lineNumOffset + 1
			ret = true
		}
	}
	return ret
}

func (v *View) GetMouseRelativePositon(x, y int) (int, int) {
	x -= v.x - v.leftCol
	y = y - v.y + 1
	return x, y
}

// GetMouseClickLocation gets the location in the buffer from a mouse click
// on the screen
func (v *View) GetMouseClickLocation(x, y int) (int, int) {
	x -= v.lineNumOffset - v.leftCol + v.x
	y += v.Topline - v.y

	if y-v.Topline > v.Height-1 {
		v.ScrollDown(1)
		y = v.Height + v.Topline - 1
	}
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}

	newX, newY := v.GetSoftWrapLocation(x, y)
	if newX > Count(v.Buf.Line(newY)) {
		newX = Count(v.Buf.Line(newY))
	}

	return newX, newY
}

// MoveToMouseClick moves the cursor to location x, y assuming x, y were given
// by a mouse click
func (v *View) MoveToMouseClick(x, y int) {
	if y-v.Topline > v.Height-1 {
		v.ScrollDown(1)
		y = v.Height + v.Topline - 1
	}
	if y < 0 {
		y = 0
	}
	if x < 0 {
		x = 0
	}

	x, y = v.GetSoftWrapLocation(x, y)
	if x > Count(v.Buf.Line(y)) {
		x = Count(v.Buf.Line(y))
	}
	v.Cursor.X = x
	v.Cursor.Y = y
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()
}

// Execute actions executes the supplied actions
func (v *View) ExecuteActions(actions []func(*View, bool) bool) bool {
	relocate := false
	readonlyBindingsList := []string{"Delete", "Insert", "Backspace", "Cut", "Play", "Paste", "Move", "Add", "DuplicateLine"}
	for _, action := range actions {
		readonlyBindingsResult := false
		funcName := ShortFuncName(action)
		// Check check for not allowed actions
		if v.Type.Readonly == true {
			// check for readonly and if true only let key bindings get called if they do not change the contents.
			for _, readonlyBindings := range readonlyBindingsList {
				if strings.Contains(funcName, readonlyBindings) {
					readonlyBindingsResult = true
				}
			}
		}
		if !readonlyBindingsResult {
			// call the key binding
			relocate = action(v, true) || relocate
		} else {
			messenger.Error(Language.Translate("File is readonly"))
		}
	}

	return relocate
}

// SetCursor sets the view's and buffer's cursor
func (v *View) SetCursor(c *Cursor) bool {
	if c == nil {
		return false
	}
	v.Cursor = c
	v.Buf.curCursor = c.Num

	return true
}

const autocloseOpen = "\"'`({["
const autocloseClose = "\"'`)}]"
const autocloseNewLine = ")}]"

// HandleEvent handles an event passed by the main loop
func (v *View) HandleEvent(event tcell.Event) {
	if v.Type == vtTerm {
		v.term.HandleEvent(event)
		return
	}

	if v.Type == vtRaw {
		v.Buf.Insert(v.Cursor.Loc, reflect.TypeOf(event).String()[7:])
		v.Buf.Insert(v.Cursor.Loc, fmt.Sprintf(": %q\n", event.EscSeq()))

		switch e := event.(type) {
		case *tcell.EventKey:
			if e.Key() == tcell.KeyCtrlQ {
				v.Quit(true)
			}
		}

		return
	}

	// This bool determines whether the view is relocated at the end of the function
	// By default it's true because most events should cause a relocate
	relocate := true

	v.Buf.CheckModTime()

	switch e := event.(type) {
	case *tcell.EventRaw:
		for key, actions := range bindings {
			if key.keyCode == -1 {
				if e.EscSeq() == key.escape {
					for _, c := range v.Buf.cursors {
						ok := v.SetCursor(c)
						if !ok {
							break
						}
						relocate = false
						relocate = v.ExecuteActions(actions) || relocate
					}
					v.SetCursor(&v.Buf.Cursor)
					v.Buf.MergeCursors()
					break
				}
			}
		}
	case *tcell.EventKey:
		// Check first if input is a key binding, if it is we 'eat' the input and don't insert a rune
		isBinding := false
		for key, actions := range bindings {
			if e.Key() == key.keyCode {
				if e.Key() == tcell.KeyRune {
					if e.Rune() != key.r {
						continue
					}
				}
				if e.Modifiers() == key.modifiers {
					var cursors []*Cursor
					if len(v.Buf.cursors) > 1 && e.Name() == "Enter" {
						// Multicursor, newline. Reverse cursor order so it works
						for i := len(v.Buf.cursors) - 1; i >= 0; i-- {
							cursors = append(cursors, v.Buf.cursors[i])
						}
					} else {
						cursors = v.Buf.cursors
					}
					for _, c := range cursors {
						ok := v.SetCursor(c)
						if !ok {
							break
						}
						relocate = false
						isBinding = true
						relocate = v.ExecuteActions(actions) || relocate
					}
					v.SetCursor(&v.Buf.Cursor)
					v.Buf.MergeCursors()
					break
				}
			}
		}

		if !isBinding && e.Key() == tcell.KeyRune {
			// Check viewtype if readonly don't insert a rune (readonly help and log view etc.)
			if v.Type.Readonly == true {
				messenger.Error(Language.Translate("File is readonly"))
			} else {
				for _, c := range v.Buf.cursors {
					v.SetCursor(c)

					// Insert a character
					if v.Cursor.HasSelection() {
						v.Cursor.DeleteSelection()
						v.Cursor.ResetSelection()
					}

					if v.isOverwriteMode {
						next := v.Cursor.Loc
						next.X++
						v.Buf.Replace(v.Cursor.Loc, next, string(e.Rune()))
					} else {
						v.Buf.Insert(v.Cursor.Loc, string(e.Rune()))
					}

					// Autoclose here for : {}[]()''""``
					if v.Buf.Settings["autoclose"].(bool) {
						n := strings.Index(autocloseOpen, string(e.Rune()))
						m := strings.Index(autocloseClose, string(e.Rune()))
						if n != -1 || m != -1 {
							if n < 3 || m > 2 {
								// Test we do not duplicate closing char
								x := v.Cursor.X
								if v.Cursor.X < len(v.Buf.LineRunes(v.Cursor.Y)) && v.Cursor.X > 1 {
									chb := string(v.Buf.LineRunes(v.Cursor.Y)[x : x+1])
									if chb == string(e.Rune()) {
										v.Delete(false)
										n = -1
									}
								}
								if n >= 0 && n < 3 && v.Cursor.X < len(v.Buf.LineRunes(v.Cursor.Y))-1 && v.Cursor.X > 1 {
									// Test surrounding chars
									ch1 := string(v.Buf.LineRunes(v.Cursor.Y)[x-2 : x-1])
									ch2 := string(v.Buf.LineRunes(v.Cursor.Y)[x : x+1])
									messenger.AddLog(noAutoCloseChar(ch1), " || ", noAutoCloseChar(ch2))
									if noAutoCloseChar(ch1) || noAutoCloseChar(ch2) {
										n = -1
									}
								}
							}
							if n >= 0 {
								v.Buf.Insert(v.Cursor.Loc, autocloseClose[n:n+1])
								v.Cursor.Left()
							}
						}
					}

					for pl := range loadedPlugins {
						if GetPluginOption(pl, "ftype") != "*" && (GetPluginOption(pl, "ftype") == nil || GetPluginOption(pl, "ftype").(string) != CurView().Buf.FileType()) {
							continue
						}
						_, err := Call(pl+".onRune", string(e.Rune()), v)
						if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
							TermMessage(err)
						}
					}

				}
				v.SetCursor(&v.Buf.Cursor)
			}
		}
	case *tcell.EventPaste:
		// Check viewtype if readonly don't paste (readonly help and log view etc.)
		if v.Type.Readonly == false {
			if !PreActionCall("Paste", v) {
				break
			}

			for _, c := range v.Buf.cursors {
				v.SetCursor(c)
				v.paste(e.Text())
			}
			v.SetCursor(&v.Buf.Cursor)

			PostActionCall("Paste", v)
		} else {
			messenger.Error(Language.Translate("File is readonly"))
		}
	case *tcell.EventMouse:
		// range validation
		// Mouse clicks outside range are handled at micro.go main event loop
		// That allows to change views, and restrict mouse events here based on the view size

		button := e.Buttons()

		// This events are relative to each view dimentions
		if e.Buttons() == tcell.Button1 || button == tcell.Button3 || button == tcell.Button2 {
			rx, ry := v.GetMouseRelativePositon(e.Position())
			if v.Type.Kind == 0 {
				//messenger.Message(ry, "?", v.Height, ": rx", rx, "?", v.lineNumOffset)
				if ry <= 0 {
					// ignore actions on tabbar or upper views
					return
				} else if ry == v.Height+1 {
					// On status line.
					v.sline.MouseEvent(e, rx, ry)
					return
				} else if ry > v.Height+1 {
					// ignore actions in messenger area or lower views
					return
				} else if (button == tcell.Button3 || button == tcell.Button2) && rx < v.lineNumOffset && v.Buf.Settings["ruler"] == true {
					v.Buf.Settings["ruler"] = false
				} else if (button == tcell.Button3 || button == tcell.Button2) && rx < 3 && v.Buf.Settings["ruler"] == false {
					v.Buf.Settings["ruler"] = true
				}
			} else {
				if ry >= v.Height {
					// ignore actions in messenger area or lower views
					return
				}
			}
		}
		// End of range validation

		// Don't relocate for mouse events
		relocate = false
		for key, actions := range bindings {
			if button == key.buttons && e.Modifiers() == key.modifiers {
				for _, c := range v.Buf.cursors {
					ok := v.SetCursor(c)
					if !ok {
						break
					}
					relocate = v.ExecuteActions(actions) || relocate
				}
				v.SetCursor(&v.Buf.Cursor)
				v.Buf.MergeCursors()
			}
		}

		for key, actions := range mouseBindings {
			if button == key.buttons && e.Modifiers() == key.modifiers {
				for _, action := range actions {
					action(v, true, e)
				}
			}
		}
		switch button {
		case tcell.ButtonNone:
			// Mouse event with no click
			if !v.mouseReleased {
				// Mouse was just released

				x, y := e.Position()
				x -= v.lineNumOffset - v.leftCol + v.x
				y += v.Topline - v.y
				v.MoveToMouseClick(x, y)
				v.savedLoc = Loc{x, y}
				v.mouseReleased = true
			}
		}
	}

	if relocate {
		v.Relocate()

		// HP, does not seems to be needed any more

		// We run relocate again because there's a bug with relocating with softwrap
		// when for example you jump to the bottom of the buffer and it tries to
		// calculate where to put the topline so that the bottom line is at the bottom
		// of the terminal and it runs into problems with visual lines vs real lines.
		// This is (hopefully) a temporary solution

		// v.Relocate()
	}
}

func (v *View) mainCursor() bool {
	return v.Buf.curCursor == len(v.Buf.cursors)-1
}

// GutterMessage creates a message in this view's gutter
func (v *View) GutterMessage(section string, lineN int, msg string, kind int) {
	lineN--
	gutterMsg := GutterMessage{
		lineNum: lineN,
		msg:     msg,
		kind:    kind,
	}
	for _, v := range v.messages {
		for _, gmsg := range v {
			if gmsg.lineNum == lineN {
				return
			}
		}
	}
	messages := v.messages[section]
	v.messages[section] = append(messages, gutterMsg)
}

// ClearGutterMessages clears all gutter messages from a given section
func (v *View) ClearGutterMessages(section string) {
	v.messages[section] = []GutterMessage{}
}

// ClearAllGutterMessages clears all the gutter messages
func (v *View) ClearAllGutterMessages() {
	for k := range v.messages {
		v.messages[k] = []GutterMessage{}
	}
}

// Opens the given help page in a new horizontal split
func (v *View) openHelp(helpPage string) {
	if data, err := FindRuntimeFile(RTHelp, helpPage).Data(); err != nil {
		TermMessage("Unable to load help text", helpPage, "\n", err)
	} else {
		helpBuffer := NewBufferFromString(string(data), helpPage+".md")
		helpBuffer.name = "Help"

		if v.Type == vtHelp {
			v.OpenBuffer(helpBuffer)
		} else {
			v.HSplit(helpBuffer)
			CurView().Type = vtHelp
			v.Relocate()
		}
	}
}

func (v *View) FindCurLine(n int, s string) Loc {
	var newLoc Loc
	var s2 string

	Y := v.savedLoc.Y + n
	Min := 0
	Max := v.Buf.LinesNum() - 1
	newLoc.Y = -1
	newLoc.X = v.savedLoc.X
	if n < 0 {
		for N := Y; N >= Min; N-- {
			s2 = SubtringSafe(v.Buf.Line(N), 0, 10)
			if s == s2 {
				newLoc.Y = N
				if newLoc.X > len(v.Buf.Line(N))-1 {
					newLoc.X = len(v.Buf.Line(N)) - 1
				}
				return newLoc
			}
		}
	} else {
		for N := Y; N <= Max; N++ {
			s2 = SubtringSafe(v.Buf.Line(N), 0, 10)
			if s == s2 {
				newLoc.Y = N
				if newLoc.X > len(v.Buf.Line(N))-1 {
					newLoc.X = len(v.Buf.Line(N)) - 1
				}
				return newLoc
			}
		}
	}
	return newLoc
}

func (v *View) SetCursorEscapeString() {
	cursorshape := v.Buf.Settings["cursorshape"].(string)
	cursorcolor := v.Buf.Settings["cursorcolor"].(string)
	ci := CursorInsert
	co := CursorOverwrite

	if cursorcolor == "disabled" {
		CursorInsert.color = ""
		CursorOverwrite.color = ""
		if CursorHadColor {
			screen.SetCursorColorShape("white", "")
			CursorHadColor = false
		}
	} else {
		CursorInsert.color = cursorcolor
		CursorOverwrite.color = "red"
		CursorHadColor = true
	}
	if cursorshape == "disabled" {
		CursorInsert.shape = ""
		CursorOverwrite.shape = ""
		if CursorHadShape {
			screen.SetCursorColorShape("", "block")
			CursorHadShape = false
		}
	} else {
		CursorHadShape = true
		CursorInsert.shape = cursorshape
		if cursorshape == "underline" {
			CursorOverwrite.shape = "block"
		} else {
			CursorOverwrite.shape = "underline"
		}
	}
	if ((ci.color != CursorInsert.color || ci.shape != CursorInsert.shape) && v.isOverwriteMode == false) || ((co.color != CursorOverwrite.color || co.shape != CursorOverwrite.shape) && v.isOverwriteMode) {
		v.SetCursorColorShape()
	}
}

func (v *View) SetCursorColorShape() {
	if v.isOverwriteMode {
		screen.SetCursorColorShape(CursorOverwrite.color, CursorOverwrite.shape)
	} else {
		screen.SetCursorColorShape(CursorInsert.color, CursorInsert.shape)
	}
}

// DisplayView draws the view to the screen
func (v *View) DisplayView() {
	ActiveView := true
	if CurView().Num != v.Num {
		ActiveView = false
	}
	if v.Type == vtTerm {
		v.term.Display()
		return
	}

	if v.Buf.Settings["softwrap"].(bool) && v.leftCol != 0 {
		v.leftCol = 0
	}

	if v.Type == vtLog || v.Type == vtRaw {
		// Log or raw views should always follow the cursor...
		v.Relocate()
	}

	if CurView().Type.Kind == 0 && LastView != CurView().Num && CurView().Cursor.Loc != CurView().savedLoc && Mouse.Click == false {
		// HP : Set de cursor in last known position for this view
		// It happens when 2+ views point to same buffer
		// Set into current view boudaries
		if CurView().savedLoc.Y > CurView().Buf.End().Y {
			CurView().savedLoc.Y = CurView().Buf.End().Y
			if CurView().savedLoc.X > CurView().Buf.End().X {
				CurView().savedLoc.X = CurView().Buf.End().X
			}
		}
		currLine := SubtringSafe(CurView().Buf.Line(CurView().savedLoc.Y), 0, 20)
		if currLine != CurView().savedLine {
			var newLoc Loc
			// Line has moved, find new position using as a reference the last known line beggining
			newLoc = CurView().FindCurLine(-1, CurView().savedLine)
			if newLoc.Y < 0 {
				newLoc = CurView().FindCurLine(1, CurView().savedLine)
				if newLoc.Y < 0 {
					newLoc = CurView().savedLoc
				}
			}
			CurView().Cursor.GotoLoc(newLoc)
		} else {
			CurView().Cursor.GotoLoc(CurView().savedLoc)
		}
		CurView().Relocate()
		if CurView().Cursor.Loc.Y > CurView().Height-3 {
			//CurView().Center(false)
		}
	}

	// We need to know the string length of the largest line number
	// so we can pad appropriately when displaying line numbers
	maxLineNumLength := len(strconv.Itoa(v.Buf.NumLines))

	if v.Buf.Settings["ruler"] == true {
		// + 1 for the little space after the line number
		v.lineNumOffset = maxLineNumLength + 1
	} else {
		v.lineNumOffset = 0
	}

	// We need to add to the line offset if there are gutter messages
	var hasGutterMessages bool
	for _, v := range v.messages {
		if len(v) > 0 {
			hasGutterMessages = true
		}
	}
	if hasGutterMessages {
		v.lineNumOffset += 2
	}

	divider := 0
	if v.x != 0 {
		// One space for the extra split divider
		v.lineNumOffset++
		divider = 1
	}

	xOffset := v.x + v.lineNumOffset
	yOffset := v.y

	height := v.Height
	width := v.Width
	left := v.leftCol
	top := v.Topline

	WindowOffset := width / 4

	if ActiveView {
		// Have a window margin on edges for long lines if the windows is not wide enough
		if v.Buf.Settings["softwrap"].(bool) == false && StringWidth(v.Buf.Line(v.Cursor.Loc.Y), int(v.Buf.Settings["tabsize"].(float64))) > width-v.lineNumOffset {
			shift := 0
			if v.Cursor.GetVisualX()+1 < width-v.lineNumOffset && v.Cursor.GetVisualX()+1 > width-v.lineNumOffset-WindowOffset && left == 0 {
				shift = WindowOffset - (width - v.lineNumOffset - v.Cursor.GetVisualX())
			} else if v.Cursor.GetVisualX()+1 >= width-v.lineNumOffset && v.Cursor.GetVisualX()-WindowOffset > left+WindowOffset {
				shift = WindowOffset
			} else if (v.Cursor.GetVisualX()+1 >= width-v.lineNumOffset && v.Cursor.GetVisualX()-WindowOffset <= left+WindowOffset) || left > 0 {
				shift = v.Cursor.GetVisualX() - left - WindowOffset
			}
			left += shift
		}
	}

	v.cellview.Draw(v.Buf, top, height, left, width-v.lineNumOffset, ActiveView)
	_, bgDisabled, _ := defStyle.Decompose()
	if ActiveView == false {
		if style, ok := colorscheme["unfocused"]; ok {
			_, bgDisabled, _ = style.Decompose()
		} else {
			bgDisabled = tcell.Color236
		}
	}

	screenX := v.x
	realLineN := top - 1
	visualLineN := 0
	var line []*Char
	for visualLineN, line = range v.cellview.lines {
		var firstChar *Char
		if len(line) > 0 {
			firstChar = line[0]
		}

		var softwrapped bool
		if firstChar != nil {
			if firstChar.realLoc.Y == realLineN {
				softwrapped = true
			}
			realLineN = firstChar.realLoc.Y
		} else {
			realLineN++
		}

		colorcolumn := int(v.Buf.Settings["colorcolumn"].(float64))
		if colorcolumn != 0 && xOffset+colorcolumn-v.leftCol < v.Width {
			style := GetColor("color-column")
			fg, _, _ := style.Decompose()
			st := defStyle.Background(fg)
			screen.SetContent(xOffset+colorcolumn-v.leftCol, yOffset+visualLineN, ' ', nil, st)
		}

		screenX = v.x

		// If there are gutter messages we need to display the '>>' symbol here
		if hasGutterMessages {
			// msgOnLine stores whether or not there is a gutter message on this line in particular
			msgOnLine := false
			for k := range v.messages {
				for _, msg := range v.messages[k] {
					if msg.lineNum == realLineN {
						msgOnLine = true
						gutterStyle := defStyle
						switch msg.kind {
						case GutterInfo:
							if style, ok := colorscheme["gutter-info"]; ok {
								gutterStyle = style
							}
						case GutterWarning:
							if style, ok := colorscheme["gutter-warning"]; ok {
								gutterStyle = style
							}
						case GutterError:
							if style, ok := colorscheme["gutter-error"]; ok {
								gutterStyle = style
							}
						}
						screen.SetContent(screenX, yOffset+visualLineN, '>', nil, gutterStyle)
						screenX++
						screen.SetContent(screenX, yOffset+visualLineN, '>', nil, gutterStyle)
						screenX++
						if v.Cursor.Y == realLineN && !messenger.hasPrompt {
							messenger.Message(msg.msg)
							messenger.gutterMessage = true
						}
					}
				}
			}
			// If there is no message on this line we just display an empty offset
			if !msgOnLine {
				screen.SetContent(screenX, yOffset+visualLineN, ' ', nil, defStyle)
				screenX++
				screen.SetContent(screenX, yOffset+visualLineN, ' ', nil, defStyle)
				screenX++
				if v.Cursor.Y == realLineN && messenger.gutterMessage {
					messenger.Reset()
					messenger.gutterMessage = false
				}
			}
		}

		lineNumStyle := defStyle
		if v.Buf.Settings["ruler"] == true {
			// Write the line number
			if ActiveView {
				if style, ok := colorscheme["line-number"]; ok {
					lineNumStyle = style
				}
			} else {
				if style, ok := colorscheme["unfocused"]; ok {
					lineNumStyle = style
				}
			}
			if style, ok := colorscheme["current-line-number"]; ok {
				if realLineN == v.Cursor.Y && ActiveView && !v.Cursor.HasSelection() {
					lineNumStyle = style
				}
			}

			lineNum := strconv.Itoa(realLineN + 1)

			// Write the spaces before the line number if necessary
			for i := 0; i < maxLineNumLength-len(lineNum); i++ {
				screen.SetContent(screenX+divider, yOffset+visualLineN, ' ', nil, lineNumStyle)
				screenX++
			}
			if softwrapped && visualLineN != 0 {
				// Pad without the line number because it was written on the visual line before
				for range lineNum {
					screen.SetContent(screenX+divider, yOffset+visualLineN, ' ', nil, lineNumStyle)
					screenX++
				}
			} else {
				// Write the actual line number
				for _, ch := range lineNum {
					screen.SetContent(screenX+divider, yOffset+visualLineN, ch, nil, lineNumStyle)
					screenX++
				}
			}

			// Write the extra space
			screen.SetContent(screenX+divider, yOffset+visualLineN, ' ', nil, lineNumStyle)
			screenX++
		}

		// Cursor
		var lastChar *Char
		cursorSet := false
		for _, char := range line {
			if char != nil {
				lineStyle := char.style

				colorcolumn := int(v.Buf.Settings["colorcolumn"].(float64))
				if colorcolumn != 0 && char.visualLoc.X == colorcolumn {
					style := GetColor("color-column")
					fg, _, _ := style.Decompose()
					lineStyle = lineStyle.Background(fg)
				}

				charLoc := char.realLoc
				for _, c := range v.Buf.cursors {
					v.SetCursor(c)
					if ActiveView && v.Cursor.HasSelection() &&
						(charLoc.GreaterEqual(v.Cursor.CurSelection[0]) && charLoc.LessThan(v.Cursor.CurSelection[1]) ||
							charLoc.LessThan(v.Cursor.CurSelection[0]) && charLoc.GreaterEqual(v.Cursor.CurSelection[1])) {
						// The current character is selected
						lineStyle = defStyle.Reverse(true)

						if style, ok := colorscheme["selection"]; ok {
							lineStyle = style
						}
					}
				}
				if ActiveView {
					v.SetCursor(&v.Buf.Cursor)
				}

				if v.Buf.Settings["cursorline"].(bool) && ActiveView &&
					!v.Cursor.HasSelection() && v.Cursor.Y == realLineN {
					style := GetColor("cursor-line")
					fg, _, _ := style.Decompose()
					lineStyle = lineStyle.Background(fg)
				} else if ActiveView == false {
					lineStyle = lineStyle.Background(bgDisabled)
				}

				screen.SetContent(xOffset+char.visualLoc.X, yOffset+char.visualLoc.Y, char.drawChar, nil, lineStyle)

				for i, c := range v.Buf.cursors {
					v.SetCursor(c)
					if ActiveView && !v.Cursor.HasSelection() &&
						v.Cursor.Y == char.realLoc.Y && v.Cursor.X == char.realLoc.X && (!cursorSet || i != 0) {
						ShowMultiCursor(xOffset+char.visualLoc.X, yOffset+char.visualLoc.Y, i)
						cursorSet = true
					}
				}
				if ActiveView {
					v.SetCursor(&v.Buf.Cursor)
				}

				lastChar = char
			}
		}

		lastX := 0
		var realLoc Loc
		var visualLoc Loc
		var cx, cy int
		if lastChar != nil {
			lastX = xOffset + lastChar.visualLoc.X + lastChar.width
			for i, c := range v.Buf.cursors {
				v.SetCursor(c)
				if ActiveView && !v.Cursor.HasSelection() &&
					v.Cursor.Y == lastChar.realLoc.Y && v.Cursor.X == lastChar.realLoc.X+1 {
					ShowMultiCursor(lastX, yOffset+lastChar.visualLoc.Y, i)
					cx, cy = lastX, yOffset+lastChar.visualLoc.Y
				}
			}
			v.SetCursor(&v.Buf.Cursor)
			realLoc = Loc{lastChar.realLoc.X + 1, realLineN}
			visualLoc = Loc{lastX - xOffset, lastChar.visualLoc.Y}
		} else if len(line) == 0 {
			for i, c := range v.Buf.cursors {
				v.SetCursor(c)
				if ActiveView && !v.Cursor.HasSelection() &&
					v.Cursor.Y == realLineN {
					ShowMultiCursor(xOffset, yOffset+visualLineN, i)
					cx, cy = xOffset, yOffset+visualLineN
				}
			}
			v.SetCursor(&v.Buf.Cursor)
			lastX = xOffset
			realLoc = Loc{0, realLineN}
			visualLoc = Loc{0, visualLineN}
		}

		if ActiveView && v.Cursor.HasSelection() &&
			(realLoc.GreaterEqual(v.Cursor.CurSelection[0]) && realLoc.LessThan(v.Cursor.CurSelection[1]) ||
				realLoc.LessThan(v.Cursor.CurSelection[0]) && realLoc.GreaterEqual(v.Cursor.CurSelection[1])) {
			// The current character is selected
			selectStyle := defStyle.Reverse(true)

			if style, ok := colorscheme["selection"]; ok {
				selectStyle = style
			}
			screen.SetContent(xOffset+visualLoc.X, yOffset+visualLoc.Y, ' ', nil, selectStyle)
		}

		if v.Buf.Settings["cursorline"].(bool) && ActiveView &&
			!v.Cursor.HasSelection() && v.Cursor.Y == realLineN {
			for i := lastX; i < xOffset+v.Width-v.lineNumOffset; i++ {
				style := GetColor("cursor-line")
				fg, _, _ := style.Decompose()
				style = style.Background(fg)
				if !(ActiveView && !v.Cursor.HasSelection() && i == cx && yOffset+visualLineN == cy) {
					screen.SetContent(i, yOffset+visualLineN, ' ', nil, style)
				}
			}
		}

		// Fill trailing space with inactive background
		if ActiveView == false {
			for i := lastX; i < xOffset+v.Width-v.lineNumOffset; i++ {
				screen.SetContent(i, yOffset+visualLineN, ' ', nil, defStyle.Background(bgDisabled))
			}
		}
	}

	if divider != 0 {
		dividerStyle := defStyle
		if style, ok := colorscheme["divider"]; ok {
			dividerStyle = style
		}
		for i := 0; i < v.Height; i++ {
			screen.SetContent(v.x, yOffset+i, tcell.RuneVLine, nil, dividerStyle.Reverse(true))
		}
	}

	// ---------------------------------------------------
	// onDisplayFocus Event
	// ---------------------------------------------------
	if (CurView().Type.Kind == 0 && LastView != CurView().Num) || (LastTab != curTab) {
		LastView = CurView().Num
		LastTab = curTab

		for pl := range loadedPlugins {
			if GetPluginOption(pl, "ftype") != "*" && (GetPluginOption(pl, "ftype") == nil || GetPluginOption(pl, "ftype").(string) != CurView().Buf.FileType()) {
				continue
			}
			_, err := Call(pl+".onDisplayFocus", CurView())
			if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
				TermMessage(err)
				continue
			}
		}
	}
}

// ShowMultiCursor will display a cursor at a location
// If i == 0 then the terminal cursor will be used
// Otherwise a fake cursor will be drawn at the position
func ShowMultiCursor(x, y, i int) {
	if i == 0 {
		screen.ShowCursor(x, y)
	} else {
		r, _, _, _ := screen.GetContent(x, y)
		screen.SetContent(x, y, r, nil, defStyle.Reverse(true))
	}
}

// Display renders the view, the cursor, and statusline
func (v *View) Display() {
	if globalSettings["termtitle"].(bool) {
		screen.SetTitle("micro: " + v.Buf.GetName())
	}
	v.DisplayView()
	// Don't draw the cursor if it is out of the viewport or if it has a selection
	if v.Num == tabs[curTab].CurView && (v.Cursor.Y-v.Topline < 0 || v.Cursor.Y-v.Topline > v.Height-1 || v.Cursor.HasSelection()) {
		screen.HideCursor()
	}
	v.sline.Display()
}
