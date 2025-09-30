package main

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/hanspr/tcell/v2"
	lua "github.com/yuin/gopher-lua"
)

// PreActionCall executes the lua pre callback if possible
func PreActionCall(funcName string, view *View, args ...any) bool {
	executeAction := true
	for pl := range loadedPlugins {
		if GetPluginOption(pl, "ftype") != "*" && (GetPluginOption(pl, "ftype") == nil || GetPluginOption(pl, "ftype").(string) != CurView().Buf.FileType()) {
			continue
		}
		ret, err := Call(pl+".pre"+funcName, append([]any{view}, args...)...)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			executeAction = false
		}
	}
	return executeAction
}

// PostActionCall executes the lua plugin callback if possible
func PostActionCall(funcName string, view *View, args ...any) bool {
	relocate := true
	for pl := range loadedPlugins {
		if GetPluginOption(pl, "ftype") != "*" && (GetPluginOption(pl, "ftype") == nil || GetPluginOption(pl, "ftype").(string) != CurView().Buf.FileType()) {
			continue
		}
		ret, err := Call(pl+".on"+funcName, append([]any{view}, args...)...)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
		if ret == lua.LFalse {
			relocate = false
		}
	}
	return relocate
}

func (v *View) deselect(index int) bool {
	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[index]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
		return true
	}
	return false
}

// MousePress is the event that should happen when a normal click happens
// This is almost always bound to left click
func (v *View) MousePress(usePlugin bool, e *tcell.EventMouse) bool {

	if usePlugin && !PreActionCall("MousePress", v, e) {
		return false
	}

	x, y := e.Position()
	x -= v.lineNumOffset - v.leftCol + v.x
	y += v.Topline - v.y

	// This is usually bound to left click
	v.MoveToMouseClick(x, y)
	if v.mouseReleased {
		if len(v.Buf.cursors) > 1 {
			for i := 1; i < len(v.Buf.cursors); i++ {
				v.Buf.cursors[i] = nil
			}
			v.Buf.cursors = v.Buf.cursors[:1]
			v.Buf.UpdateCursors()
			v.Cursor.ResetSelection()
			v.Relocate()
		}
		if time.Since(v.lastClickTime)/time.Millisecond < doubleClickThreshold && (x == v.lastLoc.X && y == v.lastLoc.Y) {
			if v.doubleClick {
				// Triple click
				v.lastClickTime = time.Now()

				v.tripleClick = true
				v.doubleClick = false
			} else {
				v.lastClickTime = time.Now()

				v.doubleClick = true
				v.tripleClick = false
			}
		} else {
			v.doubleClick = false
			v.tripleClick = false
			v.lastClickTime = time.Now()
		}
		v.mouseReleased = false
		if v.tripleClick {
			v.Cursor.SelectLine()
		} else if v.doubleClick {
			v.Cursor.SelectWord(true)
		} else {
			v.Cursor.OrigSelection[0] = v.Cursor.Loc
			v.Cursor.CurSelection[0] = v.Cursor.Loc
			v.Cursor.CurSelection[1] = v.Cursor.Loc
		}
	} else if !v.mouseReleased {
		if v.tripleClick {
			v.Cursor.AddLineToSelection()
		} else if v.doubleClick {
			v.Cursor.AddWordToSelection()
		} else {
			v.Cursor.SetSelectionEnd(v.Cursor.Loc)
		}
	}

	v.lastLoc = Loc{x, y}

	if usePlugin {
		PostActionCall("MousePress", v, e)
	}
	return false
}

// ScrollUpAction scrolls the view up
func (v *View) ScrollUpAction(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ScrollUp", v) {
			return false
		}

		scrollspeed := 1
		Dt := time.Since(scrollsince) / time.Millisecond
		if Dt > 40 && Dt < 90 {
			scrollspeed = scrollcount * v.Buf.End().Y / 100
			scrollcount += 3
		} else if Dt < 140 {
			scrollspeed = 4
		} else if Dt < 240 {
			scrollspeed = 2
		} else if Dt >= 240 {
			scrollcount = 1
		}
		v.ScrollUp(scrollspeed)
		scrollsince = time.Now()

		if usePlugin {
			PostActionCall("ScrollUp", v)
		}
	}
	return false
}

// ScrollDownAction scrolls the view up
func (v *View) ScrollDownAction(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ScrollDown", v) {
			return false
		}

		scrollspeed := 1
		Dt := time.Since(scrollsince) / time.Millisecond
		if Dt > 40 && Dt < 90 {
			scrollspeed = scrollcount * v.Buf.End().Y / 100
			scrollcount += 3
		} else if Dt < 140 {
			scrollspeed = 4
		} else if Dt < 240 {
			scrollspeed = 2
		} else if Dt >= 240 {
			scrollcount = 1
		}
		v.ScrollDown(scrollspeed)
		scrollsince = time.Now()

		if usePlugin {
			PostActionCall("ScrollDown", v)
		}
	}
	return false
}

// Center centers the view on the cursor
func (v *View) Center(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Center", v) {
		return false
	}

	v.Topline = v.Cursor.Y - v.Height/2
	if v.Topline+v.Height > v.Buf.NumLines {
		v.Topline = v.Buf.NumLines - v.Height
	}
	if v.Topline < 0 {
		v.Topline = 0
	}

	if usePlugin {
		return PostActionCall("Center", v)
	}
	return true
}

// CursorUp moves the cursor up
func (v *View) CursorUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorUp", v) {
		return false
	}

	v.deselect(0)
	v.Cursor.Up()
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorUp", v)
	}
	return true
}

// CursorDown moves the cursor down
func (v *View) CursorDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorDown", v) {
		return false
	}

	v.deselect(1)
	v.Cursor.Down()
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorDown", v)
	}
	return true
}

// CursorLeft moves the cursor left
func (v *View) CursorLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorLeft", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	} else {
		tabstospaces := v.Buf.Settings["tabstospaces"].(bool)
		tabmovement := v.Buf.Settings["tabmovement"].(bool)
		if tabstospaces && tabmovement {
			tabsize := int(v.Buf.Settings["tabsize"].(float64))
			line := v.Buf.Line(v.Cursor.Y)
			if v.Cursor.X-tabsize >= 0 && line[v.Cursor.X-tabsize:v.Cursor.X] == Spaces(tabsize) && IsStrWhitespace(line[0:v.Cursor.X-tabsize]) {
				for range tabsize {
					v.Cursor.Left()
				}
			} else {
				v.Cursor.Left()
			}
		} else {
			v.Cursor.Left()
		}
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorLeft", v)
	}
	return true
}

// CursorRight moves the cursor right
func (v *View) CursorRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorRight", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	} else {
		tabstospaces := v.Buf.Settings["tabstospaces"].(bool)
		tabmovement := v.Buf.Settings["tabmovement"].(bool)
		if tabstospaces && tabmovement {
			tabsize := int(v.Buf.Settings["tabsize"].(float64))
			line := v.Buf.Line(v.Cursor.Y)
			if v.Cursor.X+tabsize <= Count(line) && line[v.Cursor.X:v.Cursor.X+tabsize] == Spaces(tabsize) && IsStrWhitespace(line[0:v.Cursor.X]) {
				for range tabsize {
					v.Cursor.Right()
				}
			} else {
				v.Cursor.Right()
			}
		} else {
			v.Cursor.Right()
		}
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorRight", v)
	}
	return true
}

// WordRight moves the cursor one word to the right
func (v *View) WordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordRight", v) {
		return false
	}

	v.Cursor.WordRight()
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("WordRight", v)
	}
	return true
}

// WordLeft moves the cursor one word to the left
func (v *View) WordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("WordLeft", v) {
		return false
	}

	v.Cursor.WordLeft()
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("WordLeft", v)
	}
	return true
}

// SkipCloseBrace skiping all continouse close braces, improve typing speed
func (v *View) SkipCloseBrace(usePlugin bool) bool {
	beginSkiping := false
	x := 0
	line := v.Buf.LineRunes(v.Cursor.Y)
	n := len(line)
	for i, c := range line {
		x++
		if i <= v.Cursor.X+1 {
			continue
		}
		skip := (c == '\'' || c == '"' || c == '}' || c == ']' || c == ')')
		if skip && beginSkiping {
			continue
		} else if skip && !beginSkiping {
			beginSkiping = true
		} else if !beginSkiping {
			continue
		} else {
			break
		}
	}
	x--
	if x <= n {
		v.Cursor.GotoLoc(Loc{x, v.Cursor.Y})
	}
	return true
}

// SelectUp selects up one line
func (v *View) SelectUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectUp", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Up()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectUp", v)
	}
	return true
}

// SelectDown selects down one line
func (v *View) SelectDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectDown", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Down()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectDown", v)
	}
	return true
}

// SelectLeft selects the character to the left of the cursor
func (v *View) SelectLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectLeft", v) {
		return false
	}

	loc := v.Cursor.Loc
	count := v.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Left()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectLeft", v)
	}
	return true
}

// SelectRight selects the character to the right of the cursor
func (v *View) SelectRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectRight", v) {
		return false
	}

	loc := v.Cursor.Loc
	count := v.Buf.End()
	if loc.GreaterThan(count) {
		loc = count
	}
	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = loc
	}
	v.Cursor.Right()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectRight", v)
	}
	return true
}

// SelectWord selects the word to the right of the cursor
func (v *View) SelectWord(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWord", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.SelectWord(false)

	if usePlugin {
		return PostActionCall("SelectWord", v)
	}
	return true
}

// SelectWordRight selects the word to the right of the cursor
func (v *View) SelectWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordRight", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordRight()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectWordRight", v)
	}
	return true
}

// SelectWordLeft selects the word to the left of the cursor
func (v *View) SelectWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectWordLeft", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.WordLeft()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectWordLeft", v)
	}
	return true
}

// StartOfLine moves the cursor to the start of the line
func (v *View) StartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("StartOfLine", v) {
		return false
	}

	v.deselect(0)

	if v.Cursor.X != 0 {
		v.Cursor.Start()
	} else {
		v.Cursor.StartOfText()
	}

	if usePlugin {
		return PostActionCall("StartOfLine", v)
	}
	return true
}

// EndOfLine moves the cursor to the end of the line
func (v *View) EndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("EndOfLine", v) {
		return false
	}

	v.deselect(0)
	if v.Buf.Settings["rmtrailingws"] == true {
		v.Buf.RemoveTrailingSpace(Loc{0, v.Cursor.Loc.Y})
	}
	v.Cursor.End()
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("EndOfLine", v)
	}
	return true
}

// SelectLine selects the entire current line
func (v *View) SelectLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectLine", v) {
		return false
	}

	v.Cursor.SelectLine()

	if usePlugin {
		return PostActionCall("SelectLine", v)
	}
	return true
}

// SelectToStartOfLine selects to the start of the current line
func (v *View) SelectToStartOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStartOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.Start()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectToStartOfLine", v)
	}
	return true
}

// SelectToEndOfLine selects to the end of the current line
func (v *View) SelectToEndOfLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEndOfLine", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.End()
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectToEndOfLine", v)
	}
	return true
}

// ParagraphPrevious moves the cursor to the previous empty line, or beginning of the buffer if there's none
func (v *View) ParagraphPrevious(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ParagraphPrevious", v) {
		return false
	}
	var line int
	for line = v.Cursor.Y; line > 0; line-- {
		if len(v.Buf.lines[line].data) == 0 && line != v.Cursor.Y {
			v.Cursor.X = 0
			v.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == 0 {
		v.Cursor.Loc = v.Buf.Start()
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("ParagraphPrevious", v)
	}
	return true
}

// ParagraphNext moves the cursor to the next empty line, or end of the buffer if there's none
func (v *View) ParagraphNext(usePlugin bool) bool {
	if usePlugin && !PreActionCall("ParagraphNext", v) {
		return false
	}

	var line int
	for line = v.Cursor.Y; line < len(v.Buf.lines); line++ {
		if len(v.Buf.lines[line].data) == 0 && line != v.Cursor.Y {
			v.Cursor.X = 0
			v.Cursor.Y = line
			break
		}
	}
	// If no empty line found. move cursor to end of buffer
	if line == len(v.Buf.lines) {
		v.Cursor.Loc = v.Buf.End()
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("ParagraphNext", v)
	}
	return true
}

// Retab changes all tabs to spaces or all spaces to tabs depending
// on the user's settings
func (v *View) Retab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Retab", v) {
		return false
	}

	toSpaces := v.Buf.Settings["tabstospaces"].(bool)
	tabsize := int(v.Buf.Settings["tabsize"].(float64))
	dirty := false

	for i := range v.Buf.NumLines {
		l := v.Buf.Line(i)

		ws := GetLeadingWhitespace(l)
		if ws != "" {
			if toSpaces {
				ws = strings.Replace(ws, "\t", Spaces(tabsize), -1)
			} else {
				ws = strings.Replace(ws, Spaces(tabsize), "\t", -1)
			}
		}

		l = strings.TrimLeft(l, " \t")
		v.Buf.lines[i].data = []byte(ws + l)
		dirty = true
	}

	v.Buf.IsModified = dirty
	if usePlugin {
		return PostActionCall("Retab", v)
	}
	return true
}

// CursorStart moves the cursor to the start of the buffer
func (v *View) CursorStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorStart", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.X = 0
	v.Cursor.Y = 0
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorStart", v)
	}
	return true
}

// CursorEnd moves the cursor to the end of the buffer
func (v *View) CursorEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorEnd", v) {
		return false
	}

	v.deselect(0)

	v.Cursor.Loc = v.Buf.End()
	v.Cursor.StoreVisualX()
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorEnd", v)
	}
	return true
}

// SelectToStart selects the text from the cursor to the start of the buffer
func (v *View) SelectToStart(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToStart", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorStart(false)
	v.Cursor.SelectTo(v.Buf.Start())

	if usePlugin {
		return PostActionCall("SelectToStart", v)
	}
	return true
}

// SelectToEnd selects the text from the cursor to the end of the buffer
func (v *View) SelectToEnd(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectToEnd", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.CursorEnd(false)
	v.Cursor.SelectTo(v.Buf.End())

	if usePlugin {
		return PostActionCall("SelectToEnd", v)
	}
	return true
}

// InsertSpace inserts a space
func (v *View) InsertSpace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertSpace", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.Buf.Insert(v.Cursor.Loc, " ")

	if usePlugin {
		return PostActionCall("InsertSpace", v)
	}
	return true
}

// InsertNewline inserts a newline plus possible some whitespace if autoindent is on
func (v *View) InsertNewline(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertNewline", v) {
		return false
	}

	// Insert a newline
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}

	ws := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))

	// Autoclose for (),[],{}
	if v.Buf.Settings["autoclose"].(bool) && v.Cursor.X < len(v.Buf.LineRunes(v.Cursor.Y)) {
		cha := string(v.Buf.LineRunes(v.Cursor.Y)[v.Cursor.X : v.Cursor.X+1])
		if strings.Contains(autocloseNewLine, cha) {
			wsn := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
			v.Buf.Insert(v.Cursor.Loc, "\n"+wsn+"\n"+wsn)
			if v.Buf.Settings["smartindent"].(bool) {
				v.Cursor.Up()
				v.Buf.SmartIndent(v.Cursor.Loc, v.Cursor.Loc)
			}
		} else {
			v.Buf.Insert(v.Cursor.Loc, "\n")
			cSave := v.Cursor.Loc
			if v.Buf.Settings["smartindent"].(bool) {
				v.Buf.SmartIndent(v.Cursor.Loc, v.Cursor.Loc)
			}
			v.Cursor.GotoLoc(Loc{cSave.X + CountLeadingWhitespace(v.Buf.Line(v.Cursor.Y)), v.Cursor.Y})
		}
	} else {
		v.Buf.Insert(v.Cursor.Loc, "\n")
		if v.Buf.Settings["autoindent"].(bool) && !v.Buf.Settings["smartindent"].(bool) {
			wsn := GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y + 1))
			if len(wsn) > len(ws) {
				ws = wsn
			}
			v.Buf.Insert(v.Cursor.Loc, ws)

			// Remove the whitespaces if keepautoindent setting is off
			if IsSpacesOrTabs(v.Buf.Line(v.Cursor.Y-1)) && !v.Buf.Settings["keepautoindent"].(bool) {
				line := v.Buf.Line(v.Cursor.Y - 1)
				v.Buf.Remove(Loc{0, v.Cursor.Y - 1}, Loc{Count(line), v.Cursor.Y - 1})
			}
		} else if v.Buf.Settings["smartindent"].(bool) {
			v.Buf.SmartIndent(Loc{v.Cursor.Loc.X, v.Cursor.Loc.Y - 1}, Loc{v.Cursor.Loc.X, v.Cursor.Loc.Y - 1})
			v.Buf.SmartIndent(v.Cursor.Loc, v.Cursor.Loc)
		}
	}
	v.savedLoc = v.Cursor.Loc
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	if usePlugin {
		return PostActionCall("InsertNewline", v)
	}
	return true
}

// DeleteWord deletes the word under the cursor
func (v *View) DeleteWord(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWord", v) {
		return false
	}

	v.SelectWord(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("DeleteWord", v)
	}
	return true
}

// DeleteWordRight deletes the word to the right of the cursor
func (v *View) DeleteWordRight(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordRight", v) {
		return false
	}

	v.SelectWordRight(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("DeleteWordRight", v)
	}
	return true
}

// DeleteWordLeft deletes the word to the left of the cursor
func (v *View) DeleteWordLeft(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteWordLeft", v) {
		return false
	}

	v.SelectWordLeft(false)
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("DeleteWordLeft", v)
	}
	return true
}

// Backspace deletes the previous character
func (v *View) Backspace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Backspace", v) {
		return false
	}

	// Remove autoclose here
	if v.Buf.Settings["autoclose"].(bool) && v.Cursor.X > 0 && v.Cursor.X < len(v.Buf.LineRunes(v.Cursor.Y)) {
		n := strings.Index(autocloseClose, string(v.Buf.LineRunes(v.Cursor.Y)[v.Cursor.X:v.Cursor.X+1]))
		m := strings.Index(autocloseOpen, string(v.Buf.LineRunes(v.Cursor.Y)[v.Cursor.X-1:v.Cursor.X]))
		if n >= 0 && m == n {
			v.Delete(false)
		}
	}

	// Delete a character
	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else if v.Cursor.Loc.GreaterThan(v.Buf.Start()) {
		// Cursor Left, then Delete
		lineStart := sliceEnd(v.Buf.LineBytes(v.Cursor.Y), v.Cursor.X)
		tabSize := int(v.Buf.Settings["tabsize"].(float64))
		loc := v.Cursor.Loc
		if v.Buf.Settings["tabstospaces"].(bool) && IsSpaces(lineStart) && utf8.RuneCount(lineStart) != 0 && utf8.RuneCount(lineStart)%tabSize == 0 {
			nloc := new(Loc)
			nloc.X = loc.X
			nloc.Y = loc.Y
			for i := 1; i <= tabSize; i++ {
				*nloc = nloc.left(v.Buf)
			}
			v.Buf.Remove(*nloc, loc)
		} else {
			v.Buf.Remove(loc.left(v.Buf), loc)
		}
	}
	v.savedLoc = v.Cursor.Loc
	v.Cursor.LastVisualX = v.Cursor.GetVisualX()

	if usePlugin {
		return PostActionCall("Backspace", v)
	}
	return true
}

// Delete deletes the next character
func (v *View) Delete(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Delete", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
	} else {
		tabSize := int(v.Buf.Settings["tabsize"].(float64))
		loc := v.Cursor.Loc
		lineStart := sliceEnd(v.Buf.LineBytes(v.Cursor.Y), v.Cursor.X+tabSize)
		if v.Cursor.X%tabSize == 0 && v.Buf.Settings["tabstospaces"].(bool) && IsSpaces(lineStart) && utf8.RuneCount(lineStart)%tabSize == 0 {
			v.Buf.Remove(loc, loc.Move(tabSize, v.Buf))
		} else if loc.LessThan(v.Buf.End()) {
			v.Buf.Remove(loc, loc.Move(1, v.Buf))
		}
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("Delete", v)
	}
	return true
}

// IndentSelection indents the current selection
func (v *View) IndentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("IndentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0]
		end := v.Cursor.CurSelection[1]
		if end.Y < start.Y {
			start, end = end, start
			v.Cursor.SetSelectionStart(start)
			v.Cursor.SetSelectionEnd(end)
		}
		if v.Buf.Settings["smartindent"].(bool) {
			v.Buf.SmartIndent(start, end)
			v.Cursor.SetSelectionStart(start)
		} else {
			startY := start.Y
			endY := end.Move(-1, v.Buf).Y
			endX := end.Move(-1, v.Buf).X
			tabsize := len(v.Buf.IndentString())
			for y := startY; y <= endY; y++ {
				v.Buf.Insert(Loc{0, y}, v.Buf.IndentString())
				if y == startY && start.X > 0 {
					v.Cursor.SetSelectionStart(start.Move(tabsize, v.Buf))
				}
				if y == endY {
					v.Cursor.SetSelectionEnd(Loc{endX + tabsize + 1, endY})
				}
			}
		}
		v.Cursor.Relocate()
		v.savedLoc = v.Cursor.Loc

		if usePlugin {
			return PostActionCall("IndentSelection", v)
		}
		return true
	}
	return false
}

// OutdentLine moves the current line back one indentation
func (v *View) OutdentLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OutdentLine", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}

	for x := 0; x < len(v.Buf.IndentString()); x++ {
		if len(GetLeadingWhitespace(v.Buf.Line(v.Cursor.Y))) == 0 {
			break
		}
		v.Buf.Remove(Loc{0, v.Cursor.Y}, Loc{1, v.Cursor.Y})
	}
	v.Cursor.Relocate()
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("OutdentLine", v)
	}
	return true
}

// OutdentSelection takes the current selection and moves it back one indent level
func (v *View) OutdentSelection(usePlugin bool) bool {
	if usePlugin && !PreActionCall("OutdentSelection", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0]
		end := v.Cursor.CurSelection[1]
		if end.Y < start.Y {
			start, end = end, start
			v.Cursor.SetSelectionStart(start)
			v.Cursor.SetSelectionEnd(end)
		}

		startY := start.Y
		endY := end.Move(-1, v.Buf).Y
		for y := startY; y <= endY; y++ {
			for x := 0; x < len(v.Buf.IndentString()); x++ {
				if len(GetLeadingWhitespace(v.Buf.Line(y))) == 0 {
					break
				}
				v.Buf.Remove(Loc{0, y}, Loc{1, y})
			}
		}
		v.Cursor.Relocate()
		v.savedLoc = v.Cursor.Loc

		if usePlugin {
			return PostActionCall("OutdentSelection", v)
		}
		return true
	}
	return false
}

// InsertTab inserts a tab or spaces
func (v *View) InsertTab(usePlugin bool) bool {
	if usePlugin && !PreActionCall("InsertTab", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		return false
	}

	tabBytes := len(v.Buf.IndentString())
	bytesUntilIndent := tabBytes - (v.Cursor.GetVisualX() % tabBytes)
	if v.Buf.Settings["smartindent"].(bool) {
		v.Buf.RemoveTrailingSpace(Loc{0, v.Cursor.Loc.Y})
		cSave := v.Cursor.Loc
		spc := CountLeadingWhitespace(v.Buf.Line(v.Cursor.Y))
		v.Buf.SmartIndent(v.Cursor.Loc, v.Cursor.Loc)
		diff := CountLeadingWhitespace(v.Buf.Line(v.Cursor.Y)) - spc
		if diff != 0 && v.Buf.IsModified {
			v.Cursor.GotoLoc(Loc{cSave.X + diff, v.Cursor.Y})
		}
	} else if v.Buf.Settings["tabindents"].(bool) {
		v.Buf.Insert(Loc{0, v.Cursor.Loc.Y}, v.Buf.IndentString()[:bytesUntilIndent])
	} else {
		v.Buf.Insert(v.Cursor.Loc, v.Buf.IndentString()[:bytesUntilIndent])
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("InsertTab", v)
	}
	return true
}

// SaveAll saves all open buffers
func (v *View) SaveAll(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("SaveAll", v) {
			return false
		}
		for _, t := range tabs {
			for _, v := range t.Views {
				v.Save(false)
			}
		}

		if usePlugin {
			return PostActionCall("SaveAll", v)
		}
	}
	return false
}

// Save the buffer to disk
func (v *View) Save(usePlugin bool) bool {
	if !v.Buf.IsModified {
		// Do not save if buffer is clean
		messenger.Alert("info", Language.Translate("Buffer has no changes to save"))
		return false
	}
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Save", v) {
			return false
		}

		messenger.Alert("info", Language.Translate("Saving, please wait ...."))
		if v.Type.Scratch {
			// We can't save any view type with scratch set. eg help and log text
			return false
		}
		// If this is an empty buffer, ask for a filename
		if v.Buf.Path == "" {
			v.SaveAs(false)
		} else {
			v.saveToFile(v.Buf.AbsPath)
		}

		if v.Buf.RunFormatter() {
			v.ReOpen()
		}

		if usePlugin {
			return PostActionCall("Save", v)
		}
	}
	return false
}

// This function saves the buffer to `filename` and changes the buffer's path and name
// to `filename` if the save is successful
func (v *View) saveToFile(filename string) {
	err := v.Buf.SaveAs(filename)
	if err != nil {
		messenger.Alert("error", err.Error())
	} else {
		v.Buf.Path = filename
		v.Buf.name = filename
		v.Buf.AbsPath, _ = filepath.Abs(filename)
		v.Buf.Fname = filepath.Base(filename)
		messenger.Message(Language.Translate("Saved") + " " + filename)
		git.GitSetStatus()
	}
}

// SaveAsAnswer Dialog Save finished
func (v *View) SaveAsAnswer(values map[string]string) {
	if values["filename"] != "" {
		if values["encoding"] != v.Buf.encoder {
			v.Buf.encoding = true
			v.Buf.encoder = values["encoding"]
		}
		v.saveToFile(values["filename"])
		if values["usePlugin"] == "1" {
			PostActionCall("SaveAs", v)
		}
	}
}

// SaveAs saves the buffer to disk with the given name
func (v *View) SaveAs(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("SaveAs", v) {
			return false
		}
		micromenu.SaveAs(v.Buf, usePlugin, v.SaveAsAnswer)
	}
	git.GitSetStatus()
	return true
}

// FindNext searches forwards for the last used search term
func (v *View) FindNext(usePlugin bool) bool {

	if !searching {
		return false
	}

	if usePlugin && !PreActionCall("FindNext", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = v.Cursor.CurSelection[1]
	} else {
		searchStart = v.Cursor.Loc
	}
	if lastSearch == "" {
		return true
	}
	// messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, true)

	if usePlugin {
		return PostActionCall("FindNext", v)
	}
	return true
}

// FindPrevious searches backwards for the last used search term
func (v *View) FindPrevious(usePlugin bool) bool {

	if !searching {
		return false
	}

	if usePlugin && !PreActionCall("FindPrevious", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		searchStart = v.Cursor.CurSelection[0]
	} else {
		searchStart = v.Cursor.Loc
	}
	// messenger.Message("Finding: " + lastSearch)
	Search(lastSearch, v, false)

	if usePlugin {
		return PostActionCall("FindPrevious", v)
	}
	return true
}

// Undo undoes the last action
func (v *View) Undo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Undo", v) {
		return false
	}

	if v.Buf.curCursor == 0 {
		v.Buf.clearCursors()
	}

	v.Buf.Undo()
	if v.Cursor.Y >= v.Bottomline()-2 || v.Cursor.Y <= v.Topline+2 {
		v.Center(false)
	}

	// If current undo stack len == Last Saved Stack Len, buffer is not modified
	if v.Buf.UndoStack.Len() == v.Buf.UndoStackRef {
		v.Buf.IsModified = false
	}

	if usePlugin {
		return PostActionCall("Undo", v)
	}
	return true
}

// Redo redoes the last action
func (v *View) Redo(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Redo", v) {
		return false
	}

	if v.Buf.curCursor == 0 {
		v.Buf.clearCursors()
	}

	v.Buf.Redo()
	if v.Cursor.Y >= v.Bottomline()-2 || v.Cursor.Y <= v.Topline+2 {
		v.Center(false)
	}

	// If current undo stack len == Last Saved Stack Len, buffer is not modified
	if v.Buf.UndoStack.Len() == v.Buf.UndoStackRef {
		v.Buf.IsModified = false
	}

	if usePlugin {
		return PostActionCall("Redo", v)
	}
	return true
}

// MaxClipboardSize limit the size of selected text to avoid editor from freezing
const MaxClipboardSize int = 50000

// SelectionTooBig calculate size of clip and return true if too big
func (v *View) SelectionTooBig() bool {
	total := 0
	for _, l := range v.Cursor.CurSelection {
		total += len(v.Buf.LineBytes(l.Y))
	}
	return total > MaxClipboardSize
}

// Copy the selection to the clipboard
func (v *View) Copy(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Copy", v) {
			return false
		}

		if v.Cursor.HasSelection() {
			if v.SelectionTooBig() {
				messenger.Alert("error", Language.Translate("Selection is too big, can not copy. You may move (up/down) or delete only."))
			} else {
				loc := v.Cursor.CurSelection
				v.Cursor.CopySelection(currEnv.ClipWhere)
				v.freshClip = true
				v.Cursor.ResetSelection()
				v.Cursor.GotoLoc(loc[0])
			}
		}

		if usePlugin {
			return PostActionCall("Copy", v)
		}
	}
	return true
}

// CopyToCloud send clip to cloud service
func (v *View) CopyToCloud(usePlugin bool) bool {
	currEnv.ClipWhere = "cloud"
	v.Copy(usePlugin)
	currEnv.ClipWhere = "local"
	return true
}

// CutLine cuts the current line to the clipboard
func (v *View) CutLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CutLine", v) {
		return false
	}
	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	if v.SelectionTooBig() {
		messenger.Alert("error", Language.Translate("Line is too big, can not cut. You may move (up/down) or delete only."))
		return false
	}
	if v.freshClip {
		if v.Cursor.HasSelection() {
			text := v.Cursor.GetSelection()
			Clip.AppendTo(&text, "local")
		}
	} else if time.Since(v.lastCutTime)/time.Second > 10*time.Second || !v.freshClip {
		v.Copy(true)
	}
	v.freshClip = true
	v.lastCutTime = time.Now()
	v.DeleteLine(false)
	v.Cursor.GotoLoc(Loc{0, v.Cursor.Y})
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CutLine", v)
	}
	return true
}

// Cut the selection to the system clipboard
func (v *View) Cut(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Cut", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.SelectionTooBig() {
			messenger.Alert("error", Language.Translate("Selection is too big, can not copy. You may move (up/down) or delete only."))
			return false
		}
		v.Cursor.CopySelection(currEnv.ClipWhere)
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.freshClip = true

		if usePlugin {
			return PostActionCall("Cut", v)
		}
		return true
	}
	if len(v.Buf.LineBytes(v.Cursor.Loc.Y)) > MaxClipboardSize {
		messenger.Alert("error", Language.Translate("Line is too big to cut"))
	} else {
		return v.CutLine(usePlugin)
	}
	return true
}

// CutToCloud send clipboard to cloud service
func (v *View) CutToCloud(usePlugin bool) bool {
	currEnv.ClipWhere = "cloud"
	v.Cut(usePlugin)
	currEnv.ClipWhere = "local"
	return true
}

// DuplicateLine duplicates the current line or selection
func (v *View) DuplicateLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DuplicateLine", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.SelectionTooBig() {
			messenger.Alert("error", Language.Translate("Line is too big, can not duplicate. You may move (up/down) or delete only."))
			return false
		}
		v.Buf.Insert(v.Cursor.CurSelection[1], v.Cursor.GetSelection())
	} else {
		if len(v.Buf.LineBytes(v.Cursor.Loc.Y)) > MaxClipboardSize {
			messenger.Alert("error", Language.Translate("Line is too big, can not duplicate. You may move (up/down) or delete only."))
			return false
		}
		loc := v.Cursor.Loc
		loc.Y++
		v.Cursor.End()
		v.Buf.Insert(v.Cursor.Loc, "\n"+v.Buf.Line(v.Cursor.Y))
		v.Cursor.GotoLoc(loc)
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("DuplicateLine", v)
	}
	return true
}

// DeleteLine deletes the current line
func (v *View) DeleteLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("DeleteLine", v) {
		return false
	}

	v.Cursor.SelectLine()
	if !v.Cursor.HasSelection() {
		return false
	}
	v.Cursor.DeleteSelection()
	v.Cursor.ResetSelection()
	// Force cursor to begin of line, or it junps to the end of next line afterwards
	v.Cursor.GotoLoc(Loc{0, v.Cursor.Y})
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("DeleteLine", v)
	}
	return true
}

// MoveLinesUp moves up the current line or selected lines if any
func (v *View) MoveLinesUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("MoveLinesUp", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[0].Y == 0 {
			return true
		}
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Y
		if start > end {
			end, start = start, end
		}

		v.Buf.MoveLinesUp(
			start,
			end,
		)
		v.Cursor.CurSelection[1].Y--
	} else {
		if v.Cursor.Loc.Y == 0 {
			return true
		}
		v.Buf.MoveLinesUp(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
	}
	v.Buf.IsModified = true
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("MoveLinesUp", v)
	}
	return true
}

// MoveLinesDown moves down the current line or selected lines if any
func (v *View) MoveLinesDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("MoveLinesDown", v) {
		return false
	}

	if v.Cursor.HasSelection() {
		if v.Cursor.CurSelection[1].Y >= len(v.Buf.lines) {
			return true
		}
		start := v.Cursor.CurSelection[0].Y
		end := v.Cursor.CurSelection[1].Y
		if start > end {
			end, start = start, end
		}

		v.Buf.MoveLinesDown(
			start,
			end,
		)
	} else {
		if v.Cursor.Loc.Y >= len(v.Buf.lines)-1 {
			return true
		}
		v.Buf.MoveLinesDown(
			v.Cursor.Loc.Y,
			v.Cursor.Loc.Y+1,
		)
	}
	v.Buf.IsModified = true
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("MoveLinesDown", v)
	}
	return true
}

// Paste whatever is in the system clipboard into the buffer
// Delete and paste if the user has a selection
func (v *View) Paste(usePlugin bool) bool {
	if usePlugin && !PreActionCall("Paste", v) {
		return false
	}

	clip := Clip.ReadFrom(currEnv.ClipWhere, "clip")
	v.paste(clip)
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("Paste", v)
	}
	return true
}

// PasteCloud get last clip from cloud
func (v *View) PasteCloud(usePlugin bool) bool {
	currEnv.ClipWhere = "cloud"
	v.Paste(usePlugin)
	currEnv.ClipWhere = "local"
	return true
}

// JumpToMatchingBrace moves the cursor to the matching brace if it is
// currently on a brace
func (v *View) JumpToMatchingBrace(usePlugin bool) bool {
	if usePlugin && !PreActionCall("JumpToMatchingBrace", v) {
		return false
	}

	for _, bp := range bracePairs {
		r := v.Cursor.RuneUnder(v.Cursor.X)
		if r == bp[0] || r == bp[1] {
			matchingBrace := v.Buf.FindMatchingBrace(bp, v.Cursor.Loc)
			v.Cursor.GotoLoc(matchingBrace)
		}
	}
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("JumpToMatchingBrace", v)
	}
	return true
}

// SelectAll selects the entire buffer
func (v *View) SelectAll(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectAll", v) {
		return false
	}

	v.Cursor.SetSelectionStart(v.Buf.Start())
	v.Cursor.SetSelectionEnd(v.Buf.End())
	// Put the cursor at the beginning
	v.Cursor.X = 0
	v.Cursor.Y = 0

	if usePlugin {
		return PostActionCall("SelectAll", v)
	}
	return true
}

// OpenDirView Keybinding to open the DirView
func (v *View) OpenDirView(usePlugin bool) bool {
	micromenu.DirTreeView()
	return true
}

// OpenFile opens a new file in the buffer
func (v *View) OpenFile(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("OpenFile", v) {
			return false
		}

		if v.CanClose() {
			input, canceled := messenger.Prompt("> ", "open ", "Open", CommandCompletion)
			if !canceled {
				HandleCommand(input)
				if usePlugin {
					return PostActionCall("OpenFile", v)
				}
			}
		}
	}
	if git.CheckGit() {
		git.GitSetStatus()
	}

	return false
}

// Start moves the viewport to the start of the buffer
func (v *View) Start(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Start", v) {
			return false
		}

		v.Topline = 0

		if usePlugin {
			return PostActionCall("Start", v)
		}
	}
	return false
}

// End moves the viewport to the end of the buffer
func (v *View) End(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("End", v) {
			return false
		}

		if v.Height > v.Buf.NumLines {
			v.Topline = 0
		} else {
			v.Topline = v.Buf.NumLines - v.Height
		}

		if usePlugin {
			return PostActionCall("End", v)
		}
	}
	return false
}

// PageUp scrolls the view up a page
func (v *View) PageUp(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PageUp", v) {
			return false
		}

		if v.Topline > v.Height {
			v.ScrollUp(v.Height)
		} else {
			v.Topline = 0
		}
		v.savedLoc = v.Cursor.Loc

		if usePlugin {
			return PostActionCall("PageUp", v)
		}
	}
	return false
}

// PageDown scrolls the view down a page
func (v *View) PageDown(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PageDown", v) {
			return false
		}

		if v.Buf.NumLines-(v.Topline+v.Height) > v.Height {
			v.ScrollDown(v.Height)
		} else if v.Buf.NumLines >= v.Height {
			v.Topline = v.Buf.NumLines - v.Height
		}
		v.savedLoc = v.Cursor.Loc

		if usePlugin {
			return PostActionCall("PageDown", v)
		}
	}
	return false
}

// SelectPageUp selects up one page
func (v *View) SelectPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectPageUp", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.UpN(v.Height)
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectPageUp", v)
	}
	return true
}

// SelectPageDown selects down one page
func (v *View) SelectPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("SelectPageDown", v) {
		return false
	}

	if !v.Cursor.HasSelection() {
		v.Cursor.OrigSelection[0] = v.Cursor.Loc
	}
	v.Cursor.DownN(v.Height)
	v.Cursor.SelectTo(v.Cursor.Loc)

	if usePlugin {
		return PostActionCall("SelectPageDown", v)
	}
	return true
}

// CursorPageUp places the cursor a page up
func (v *View) CursorPageUp(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageUp", v) {
		return false
	}

	v.deselect(0)
	v.Buf.pasteLoc.X = -1

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[0]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	}

	diff := v.Cursor.Y - v.Topline
	lines := v.Height + diff
	if v.Cursor.Y-lines < 0 {
		v.Topline = 0
		diff = 0
	} else {
		v.Topline = v.Cursor.Y - lines
	}
	v.Cursor.UpN(lines)
	v.Cursor.GotoLoc(Loc{v.Cursor.X, v.Cursor.Y + Abs(diff)})
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorPageUp", v)
	}
	return true
}

// CursorPageDown places the cursor a page up
func (v *View) CursorPageDown(usePlugin bool) bool {
	if usePlugin && !PreActionCall("CursorPageDown", v) {
		return false
	}

	v.deselect(0)
	v.Buf.pasteLoc.X = -1

	if v.Cursor.HasSelection() {
		v.Cursor.Loc = v.Cursor.CurSelection[1]
		v.Cursor.ResetSelection()
		v.Cursor.StoreVisualX()
	}

	diff := v.Topline - v.Cursor.Y
	lines := v.Height + diff
	if v.Cursor.Y+lines < v.Buf.End().Y {
		v.Topline = v.Cursor.Y + lines
	}
	v.Cursor.DownN(lines)
	if v.Cursor.Y+Abs(diff) > v.Buf.End().Y {
		diff = v.Buf.End().Y - v.Cursor.Y
	}
	v.Cursor.GotoLoc(Loc{v.Cursor.X, v.Cursor.Y + Abs(diff)})
	v.savedLoc = v.Cursor.Loc

	if usePlugin {
		return PostActionCall("CursorPageDown", v)
	}
	return true
}

// ToggleRuler turns line numbers off and on
func (v *View) ToggleRuler(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleRuler", v) {
			return false
		}

		if v.Buf.Settings["ruler"] == false {
			v.Buf.Settings["ruler"] = true
		} else {
			v.Buf.Settings["ruler"] = false
		}

		if usePlugin {
			return PostActionCall("ToggleRuler", v)
		}
	}
	return false
}

// ToggleSoftWrap change softwrap status
func (v *View) ToggleSoftWrap(usePlugin bool) bool {
	if v.Buf.Settings["softwrap"] == false {
		v.Buf.Settings["softwrap"] = true
	} else {
		v.Buf.Settings["softwrap"] = false
	}
	return false
}

// JumpLine jumps to a line and moves the view accordingly.
func (v *View) JumpLine(usePlugin bool) bool {
	if usePlugin && !PreActionCall("JumpLine", v) {
		return false
	}

	// Prompt for line number
	message := fmt.Sprintf(Language.Translate("Jump to line: (1 - %v) :"), v.Buf.NumLines)
	input, canceled := messenger.Prompt(message, "", "LineNumber", NoCompletion)
	if canceled {
		return false
	}
	var lineInt int
	var colInt int
	var err error
	if strings.Contains(input, ":") {
		split := strings.Split(input, ":")
		lineInt, err = strconv.Atoi(split[0])
		if err != nil {
			messenger.Message(Language.Translate("Invalid line number"))
			return false
		}
		colInt, err = strconv.Atoi(split[1])
		if err != nil {
			messenger.Message(Language.Translate("Invalid line number"))
			return false
		}
	} else {
		lineInt, err = strconv.Atoi(input)
		if err != nil {
			messenger.Message(Language.Translate("Invalid line number"))
			return false
		}
	}
	lineInt--
	// Move cursor and view if possible.
	if lineInt < v.Buf.NumLines && lineInt >= 0 {
		v.Cursor.X = colInt
		v.Cursor.Y = lineInt
		v.Center(false)
		v.savedLoc = v.Cursor.Loc

		if usePlugin {
			return PostActionCall("JumpLine", v)
		}
		return true
	}
	messenger.Message(Language.Translate("Invalid line number"))
	return false
}

// ClearStatus clears the messenger bar
func (v *View) ClearStatus(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ClearStatus", v) {
			return false
		}

		messenger.Message("")

		if usePlugin {
			return PostActionCall("ClearStatus", v)
		}
	}
	return false
}

// ToggleHelp toggles the help screen
func (v *View) ToggleHelp(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleHelp", v) {
			return false
		}

		if v.Type != vtHelp {
			// Open the default help
			v.openHelp("help")
		} else {
			v.Quit(true)
		}

		if usePlugin {
			return PostActionCall("ToggleHelp", v)
		}
	}
	return true
}

// CommandMode lets the user enter a command
func (v *View) CommandMode(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("CommandMode", v) {
			return false
		}

		input, canceled := messenger.Prompt("> ", "", "Command", CommandCompletion)
		if !canceled {
			HandleCommand(input)
			if usePlugin {
				return PostActionCall("CommandMode", v)
			}
		}
	}

	return false
}

// ToggleOverwriteMode lets the user toggle the text overwrite mode
func (v *View) ToggleOverwriteMode(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("ToggleOverwriteMode", v) {
			return false
		}

		v.isOverwriteMode = !v.isOverwriteMode
		v.SetCursorColorShape()

		if usePlugin {
			return PostActionCall("ToggleOverwriteMode", v)
		}
	}
	return false
}

// Escape leaves current mode
func (v *View) Escape(usePlugin bool) bool {
	if v.mainCursor() {
		// check if user is searching, or the last search is still active
		if searching || lastSearch != "" {
			ExitSearch(v)
			return true
		}
		// check if a prompt is shown, hide it and don't quit
		if messenger.hasPrompt {
			messenger.Reset()
			return true
		}
	}

	return false
}

// SafeQuit allways close none editor views first to avoid user mistakes
func (v *View) SafeQuit(usePlugin bool) bool {
	for _, vo := range tabs[curTab].Views {
		if vo.Type == vtLog || vo.Type == vtTerm || vo.Type == vtHelp {
			vo.Quit(false)
			if CurView().Type.Kind == 0 && navigationMode {
				navigationMode = false
			}
			return false
		}
	}
	v.Quit(true)
	return true
}

// QuitOthers Binding to close oposite window
func (v *View) QuitOthers(usePlugin bool) bool {
	if v.Type == vtLog || v.Type == vtTerm || v.Type == vtHelp {
		return false
	}
	curloc := v.Cursor.Loc
	if len(tabs[curTab].Views) > 1 {
		for _, vo := range tabs[curTab].Views {
			if vo.Type == vtLog || vo.Type == vtTerm || vo.Type == vtHelp {
				vo.Quit(false)
				v.Cursor.GotoLoc(curloc)
				return false
			} else if v.Num != vo.Num {
				// Avoid on focus to change location when closing
				v.savedLoc = curloc
				vo.Quit(true)
				tabs[curTab].CurView = v.Num
				v.Cursor.GotoLoc(curloc)
				return false
			}
		}
	}
	v.Quit(true)
	return true
}

// Quit this will close the current tab or view that is open
func (v *View) Quit(usePlugin bool) bool {
	if v == HelperWindow {
		HelperWindow = nil
	}

	if v.mainCursor() {
		if usePlugin && !PreActionCall("Quit", v) {
			return false
		}

		// Make sure not to quit if there are unsaved changes
		if v.CanClose() {
			cc.closeCodeComplete(v.Buf.FileType())
			LastView = -1
			if len(tabs[curTab].Views) > 1 {
				pos := v.splitNode.GetViewNumPosition(v.Num)
				v.splitNode.Delete()
				tabs[v.TabNum].Cleanup()
				tabs[v.TabNum].Resize()
				tabs[curTab].CurView = v.splitNode.GetPositionViewNum(pos)
			} else if len(tabs) > 1 {
				if len(tabs[v.TabNum].Views) == 1 {
					// Make sure we release this pointer
					tabs[curTab] = nil
					tabs = tabs[:v.TabNum+copy(tabs[v.TabNum:], tabs[v.TabNum+1:])]
					for i, t := range tabs {
						t.SetNum(i)
					}
					if curTab >= len(tabs) {
						curTab--
					}
					if curTab == 0 {
						CurView().AddTabbarSpace()
					}
				}
			} else {
				if usePlugin {
					PostActionCall("Quit", v)
				}
				messenger.SaveHistory()
				Finish(0)
			}
		}

		if usePlugin {
			return PostActionCall("Quit", v)
		}
	}
	return false
}

// QuitAll quits the whole editor; all splits and tabs
func (v *View) QuitAll(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("QuitAll", v) {
			return false
		}

		closeAll := true
		for _, tab := range tabs {
			for _, v := range tab.Views {
				if !v.CanClose() {
					closeAll = false
				}
			}
		}

		if closeAll {
			if usePlugin {
				PostActionCall("QuitAll", v)
			}
			messenger.SaveHistory()
			Finish(0)
		}
	}

	return false
}

// AddTab adds a new tab with an empty buffer
func (v *View) AddTab(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("AddTab", v) {
			return false
		} else if usePlugin {
			PostActionCall("DisplayBlur", v)
		}
		tab := NewTabFromView(NewView(NewBufferFromString("", "")))
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

		if usePlugin {
			return PostActionCall("AddTab", v)
		}
	}
	return true
}

// PreviousTab switches to the previous tab in the tab list
func (v *View) PreviousTab(usePlugin bool) bool {
	if curTab == 0 {
		return false
	}
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PreviousTab", v) {
			return false
		} else if usePlugin {
			PostActionCall("DisplayBlur", v)
		}

		if curTab > 0 {
			curTab--
		}

		if CurView().Type.Readonly {
			navigationMode = true
		} else {
			navigationMode = false
		}

		if usePlugin {
			if CurView().Type.Kind == 0 {
				PostActionCall("DisplayFocus", CurView())
			}
			return PostActionCall("PreviousTab", v)
		}
	}
	return false
}

// NextTab switches to the next tab in the tab list
func (v *View) NextTab(usePlugin bool) bool {
	if curTab == len(tabs)-1 {
		return false
	}
	if v.mainCursor() {
		if usePlugin && !PreActionCall("NextTab", v) {
			return false
		} else if usePlugin {
			PostActionCall("DisplayBlur", v)
		}

		if curTab < len(tabs)-1 {
			curTab++
		}

		if CurView().Type.Readonly {
			navigationMode = true
		} else {
			navigationMode = false
		}

		if usePlugin {
			if CurView().Type.Kind == 0 {
				PostActionCall("DisplayFocus", CurView())
			}
			return PostActionCall("NextTab", v)
		}
	}
	return false
}

// VSplitBinding opens vertical split
func (v *View) VSplitBinding(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("VSplit", v) {
			return false
		}
		if GetGlobalOption("splitempty").(bool) {
			// Open empty
			v.VSplit(NewBufferFromString("", ""))
		} else {
			// Split current Buffer
			v.VSplit(v.Buf)
		}

		v.mouseReleased = false
		if usePlugin {
			return PostActionCall("VSplit", v)
		}
	}
	return false
}

// HSplitBinding opens an horizontal split
func (v *View) HSplitBinding(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("HSplit", v) {
			return false
		}

		if GetGlobalOption("splitempty").(bool) {
			// Open empty
			v.HSplit(NewBufferFromString("", ""))
		} else {
			// Split current Buffer
			v.HSplit(v.Buf)
			v.frozen = false
			v.Relocate()
			v.frozen = true
			CurView().Relocate()
		}
		v.mouseReleased = false
		if usePlugin {
			return PostActionCall("HSplit", v)
		}
	}
	return false
}

// Unsplit closes all splits in the current tab except the active one
func (v *View) Unsplit(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("Unsplit", v) {
			return false
		}

		curView := tabs[curTab].CurView
		for i := len(tabs[curTab].Views) - 1; i >= 0; i-- {
			view := tabs[curTab].Views[i]
			if view != nil && view.Num != curView {
				view.Quit(true)
				// messenger.Message("Quit ", view.Buf.Path)
			}
		}

		if usePlugin {
			return PostActionCall("Unsplit", v)
		}
	}
	return false
}

// NextSplit changes the view to the next split
func (v *View) NextSplit(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("NextSplit", v) {
			return false
		} else if usePlugin {
			PostActionCall("DisplayBlur", v)
		}
		tab := tabs[curTab]
		pType := CurView().Type
		// Save cursor location and line reference
		v.savedLoc = v.Cursor.Loc
		v.savedLine = SubstringSafe(v.Buf.Line(v.Cursor.Loc.Y), 0, 20)
		// Find next View parsing tree_split downward
		tab.CurView = v.splitNode.GetNextPrevView(1)
		if CurView().Type.Readonly {
			navigationMode = true
		} else if pType.Readonly {
			navigationMode = false
		}
		if usePlugin {
			if CurView().Type.Kind == 0 {
				PostActionCall("DisplayFocus", CurView())
			}
			return PostActionCall("NextSplit", v)
		}
	}
	return false
}

// PreviousSplit changes the view to the previous split
func (v *View) PreviousSplit(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("PreviousSplit", v) {
			return false
		} else if usePlugin {
			PostActionCall("DisplayBlur", v)
		}

		tab := tabs[curTab]
		pType := CurView().Type
		// Save cursor location and line reference
		v.savedLoc = v.Cursor.Loc
		v.savedLine = SubstringSafe(v.Buf.Line(v.Cursor.Loc.Y), 0, 20)
		// Find next View parsing tree_split upward
		tab.CurView = v.splitNode.GetNextPrevView(-1)
		if CurView().Type.Readonly {
			navigationMode = true
		} else if pType.Readonly {
			navigationMode = false
		}
		if usePlugin {
			if CurView().Type.Kind == 0 {
				PostActionCall("DisplayFocus", CurView())
			}
			return PostActionCall("PreviousSplit", v)
		}
	}
	return false
}

// SpawnMultiCursor creates a new multiple cursor at the next occurrence of the current selection or current word
func (v *View) SpawnMultiCursor(usePlugin bool) bool {
	spawner := v.Buf.cursors[len(v.Buf.cursors)-1]
	// You can only spawn a cursor from the main cursor
	if v.Cursor == spawner {
		if usePlugin && !PreActionCall("SpawnMultiCursor", v) {
			return false
		}

		if !spawner.HasSelection() {
			spawner.SelectWord(false)
		} else {
			c := &Cursor{
				buf: v.Buf,
			}

			sel := spawner.GetSelection()

			searchStart = spawner.CurSelection[1]
			v.Cursor = c
			Search(regexp.QuoteMeta(sel), v, true)

			if v.Cursor.CurSelection == [2]Loc{c.Loc, v.Buf.cursors[0].Loc} {
				messenger.Information("done")
				return false
			}
			v.Buf.cursors = append(v.Buf.cursors, c)
			v.Buf.UpdateCursors()
			v.Relocate()
			v.Cursor = spawner
		}

		if usePlugin {
			PostActionCall("SpawnMultiCursor", v)
		}
	}

	return false
}

// SpawnMultiCursorSelect adds a cursor at the beginning of each line of a selection
func (v *View) SpawnMultiCursorSelect(usePlugin bool) bool {
	if v.Cursor == &v.Buf.Cursor {
		if usePlugin && !PreActionCall("SpawnMultiCursorSelect", v) {
			return false
		}

		// Avoid cases where multiple cursors already exist, that would create problems
		if len(v.Buf.cursors) > 1 {
			return false
		}

		var startLine int
		var endLine int

		a, b := v.Cursor.CurSelection[0].Y, v.Cursor.CurSelection[1].Y
		if a > b {
			startLine, endLine = b, a
		} else {
			startLine, endLine = a, b
		}

		if v.Cursor.HasSelection() {
			v.Cursor.ResetSelection()
			v.Cursor.GotoLoc(Loc{0, startLine})

			for i := startLine; i < endLine; i++ {
				c := &Cursor{
					buf: v.Buf,
				}
				c.GotoLoc(Loc{0, i})
				v.Buf.cursors = append(v.Buf.cursors, c)
			}
			v.Buf.MergeCursors()
			v.Buf.UpdateCursors()
		} else {
			return false
		}

		if usePlugin {
			PostActionCall("SpawnMultiCursorSelect", v)
		}

	}
	return false
}

// MouseMultiCursor is a mouse action which puts a new cursor at the mouse position
func (v *View) MouseMultiCursor(usePlugin bool, e *tcell.EventMouse) bool {
	if v.Cursor == &v.Buf.Cursor {
		if usePlugin && !PreActionCall("SpawnMultiCursorAtMouse", v, e) {
			return false
		}
		x, y := e.Position()
		x -= v.lineNumOffset - v.leftCol + v.x
		y += v.Topline - v.y

		c := &Cursor{
			buf: v.Buf,
		}
		v.Cursor = c
		v.MoveToMouseClick(x, y)
		v.Relocate()
		v.Cursor = &v.Buf.Cursor

		v.Buf.cursors = append(v.Buf.cursors, c)
		v.Buf.MergeCursors()
		v.Buf.UpdateCursors()

		if usePlugin {
			PostActionCall("SpawnMultiCursorAtMouse", v)
		}
	}
	return false
}

// SkipMultiCursor moves the current multiple cursor to the next available position
func (v *View) SkipMultiCursor(usePlugin bool) bool {
	cursor := v.Buf.cursors[len(v.Buf.cursors)-1]

	if v.mainCursor() {
		if usePlugin && !PreActionCall("SkipMultiCursor", v) {
			return false
		}
		sel := cursor.GetSelection()

		searchStart = cursor.CurSelection[1]
		v.Cursor = cursor
		Search(regexp.QuoteMeta(sel), v, true)
		v.Relocate()
		v.Cursor = cursor

		if usePlugin {
			PostActionCall("SkipMultiCursor", v)
		}
	}
	return false
}

// RemoveMultiCursor removes the latest multiple cursor
func (v *View) RemoveMultiCursor(usePlugin bool) bool {
	end := len(v.Buf.cursors)
	if end > 1 {
		if v.mainCursor() {
			if usePlugin && !PreActionCall("RemoveMultiCursor", v) {
				return false
			}

			v.Buf.cursors[end-1] = nil
			v.Buf.cursors = v.Buf.cursors[:end-1]
			v.Buf.UpdateCursors()
			v.Relocate()

			if usePlugin {
				return PostActionCall("RemoveMultiCursor", v)
			}
			return true
		}
	} else {
		v.RemoveAllMultiCursors(usePlugin)
	}
	return false
}

// RemoveAllMultiCursors removes all cursors except the base cursor
func (v *View) RemoveAllMultiCursors(usePlugin bool) bool {
	if v.mainCursor() {
		if usePlugin && !PreActionCall("RemoveAllMultiCursors", v) {
			return false
		}

		v.Buf.clearCursors()
		v.Relocate()

		if usePlugin {
			return PostActionCall("RemoveAllMultiCursors", v)
		}
		return true
	}
	return false
}

// ToggleCase change case
func (v *View) ToggleCase(usePlugin bool) bool {
	if v.Cursor.HasSelection() {
		text := v.Cursor.GetSelection()
		re, _ := regexp.Compile(`\p{Lu}`)
		if re.MatchString(text) {
			text = strings.ToLower(text)
		} else {
			text = strings.ToUpper(text)
		}
		sStart := v.Cursor.CurSelection[0]
		sEnd := v.Cursor.CurSelection[1]
		v.Cursor.DeleteSelection()
		v.Cursor.ResetSelection()
		v.Buf.Insert(v.Cursor.Loc, text)
		v.Cursor.SetSelectionStart(sStart)
		v.Cursor.SetSelectionEnd(sEnd)
	}
	return true
}

// SEARCH REPLACE

// SearchDialog create search dialog box
func (v *View) SearchDialog(usePlugin bool) bool {
	v.searchSave = v.Cursor.Loc
	micromenu.SearchReplace(v.SearchDialogFinished)
	return true
}

// SearchDialogFinished Call command:Replace
func (v *View) SearchDialogFinished(values map[string]string) {
	searchStart = v.searchSave
	v.Cursor.ResetSelection()
	if values["search"] == "" {
		v.Cursor.GotoLoc(v.searchSave)
		v.Relocate()
		return
	}
	replacing = true
	replaceTime = time.Now()
	Replace([]string{values["search"], values["replace"], values["a"], values["i"], values["l"]})
	replacing = false
}

// FindDialog create Find dialog
func (v *View) FindDialog(usePlugin bool) bool {
	v.searchSave = v.Cursor.Loc
	micromenu.Search(v.FindDialogFinished)
	return true
}

// FindDialogFinished get expression to find and search
func (v *View) FindDialogFinished(values map[string]string) {
	searchStart = v.searchSave
	v.Cursor.ResetSelection()
	if values["find"] == "" {
		v.Cursor.GotoLoc(v.searchSave)
		v.Relocate()
		return
	}
	search := values["find"]
	mods := "(?m)"
	if values["i"] == "i" {
		mods = mods + "(?i)"
	}
	search = mods + search
	lastSearch = search
	if Search(search, v, true) {
		StartSearchMode()
	} else {
		v.Cursor.GotoLoc(v.searchSave)
		v.Relocate()
	}
}

// Replace runs search and replace
func Replace(args []string) {

	if len(args) < 2 || args[0] == args[1] {
		// We need to find both a search and replace expression
		messenger.Alert("error", Language.Translate("Invalid replace statement:"), " "+strings.Join(args, " "))
		return
	}

	all := false
	noRegex := false
	mods := "(?m)"
	if len(args) > 2 {
		for _, arg := range args[2:] {
			switch arg {
			case "a":
				all = true
			case "i":
				mods = mods + "(?i)"
			case "s":
				mods = mods + "(?s)"
			case "l":
				noRegex = true
			case "":
			default:
				messenger.Alert("error", Language.Translate("Invalid flag:"), " ", arg)
				return
			}
		}
	}

	search := string(args[0])

	if noRegex {
		search = regexp.QuoteMeta(search)
	} else {
		search = mods + search
	}

	replace := string(args[1])
	replace = ExpandString(replace)
	regex, err := regexp.Compile(search)
	if err != nil {
		// There was an error with the user's regex
		messenger.Alert("error", err.Error())
		return
	}

	view := CurView()
	startLine := view.searchSave.Y
	found := 0
	view.searchLoops = 0

	for {
		var choice rune
		var canceled bool
		// The 'check' flag was used
		Search(search, view, true)
		if !view.Cursor.HasSelection() {
			break
		}
		view.Relocate()
		RedrawAll(true)
		if view.searchLoops > 0 && view.Cursor.Y > startLine {
			view.searchLoops = 0
			if view.Cursor.HasSelection() {
				view.Cursor.Loc = view.Cursor.CurSelection[0]
				view.Cursor.ResetSelection()
			}
			messenger.Reset()
			break
		}
		y := []rune(Language.Translate("y"))[0]
		n := []rune(Language.Translate("n"))[0]
		q := []rune(Language.Translate("q"))[0]
		I := []rune(Language.Translate("!"))[0]
		if all && !freeze {
			freeze = true
		} else if !all {
			choice, canceled = messenger.LetterPrompt(true, Language.Translate("Perform replacement? (y,n,q,!)"), y, n, q, I)
		}
		if canceled {
			if view.Cursor.HasSelection() {
				view.Cursor.Loc = view.Cursor.CurSelection[0]
				view.Cursor.ResetSelection()
			}
			messenger.Reset()
			break
		} else if choice == y || choice == I || all {
			sel := view.Cursor.GetSelection()
			rep := regex.ReplaceAllString(sel, replace)
			if Count(rep) > Count(sel) {
				searchStart = Loc{view.Cursor.CurSelection[1].X + 1, view.Cursor.Loc.Y}
			} else {
				searchStart = Loc{view.Cursor.CurSelection[1].X - 1, view.Cursor.Loc.Y}
			}
			if searchStart.X > len(view.Buf.LineBytes(searchStart.Y))-1 {
				searchStart = Loc{0, searchStart.Y + 1}
				if searchStart.Y > view.Buf.Len()-1 {
					searchStart.Y = view.Buf.Len() - 1
				}
			}
			view.Cursor.DeleteSelection()
			view.Buf.Insert(view.Cursor.Loc, rep)
			if !all {
				messenger.Reset()
			}
			if choice == I && !all {
				all = true
			}
			found++
		} else if choice == q {
			if view.Cursor.HasSelection() {
				view.Cursor.Loc = view.Cursor.CurSelection[0]
				view.Cursor.ResetSelection()
			}
			messenger.Reset()
			break
		} else {
			searchStart = view.Cursor.CurSelection[1]
		}
	}
	replacing = false
	freeze = false
	view.Cursor.Relocate()

	if found > 1 {
		messenger.Message(Language.Translate("Replaced"), " ", found, " "+Language.Translate("occurrences of"), " ", search)
	} else if found == 1 {
		messenger.Message(Language.Translate("Replaced"), " ", found, " "+Language.Translate("occurrence of "), search)
	} else {
		messenger.Message(Language.Translate("Nothing matched"), " ", search)
	}
}

// mi-ide Services

// UploadToCloud uload file to the cloud
func (v *View) UploadToCloud(plugin bool) bool {
	b := v.Buf
	if b.Len() == 0 {
		return true
	}
	file := b.String()
	msg := Clip.WriteTo(&file, "cloud", "file")
	if msg == "" {
		messenger.Alert("success", Language.Translate("File uploaded to cloud"))
	} else {
		messenger.Alert("error", Language.Translate(msg))
	}
	return true
}

// DownloadFromCloud download last saved file from cloud service
func (v *View) DownloadFromCloud(plugin bool) bool {
	text := Clip.ReadFrom("cloud", "file")
	if text == "" {
		messenger.Message(Language.Translate("No files to download"))
		return true
	}
	tab := NewTabFromView(NewView(NewBufferFromString(text, "")))
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
	return true
}

// CloudSettings open cloud settings dialog
func (v *View) CloudSettings(plugin bool) bool {
	micromenu.MiCloudSync()
	return true
}

// ToggleMouse enable disable mouse
func (v *View) ToggleMouse(usePlugin bool) bool {
	if mouseEnabled {
		screen.DisableMouse()
	} else {
		screen.EnableMouse()
	}
	mouseEnabled = !mouseEnabled
	previousMouseStatus = mouseEnabled
	return true
}

// MouseOnOff enable disable mouse
func MouseOnOff(b bool) {
	if b {
		screen.EnableMouse()
	} else {
		screen.DisableMouse()
	}
	mouseEnabled = b
}

// BufferSettings show Buffersettings
func (v *View) BufferSettings(usePlugin bool) bool {
	micromenu.SelLocalSettings(v.Buf)
	return true
}

// NavigationMode Toggle Navigation Mode
func (v *View) NavigationMode(usePlugin bool) bool {
	if v.Type.Readonly {
		navigationMode = true
		return true
	}
	MouseOnOff(false)
	navigationMode = !navigationMode
	return true
}

// MultiComment Add Multicomments
func (v *View) MultiComment(usePlugin bool) bool {
	if v.Buf.Settings["comment"].(string) == "" {
		messenger.Alert("warning", "settings comment not configured")
		return true
	}
	if v.Cursor.HasSelection() {
		start := v.Cursor.CurSelection[0]
		end := v.Cursor.CurSelection[1]
		if end.Y < start.Y {
			start, end = end, start
			v.Cursor.SetSelectionStart(start)
			v.Cursor.SetSelectionEnd(end)
		}
		v.Buf.AddMultiComment(start, end)
		v.Cursor.SetSelectionStart(start)
		v.Cursor.Relocate()
		v.savedLoc = v.Cursor.Loc
		return true
	}
	v.Buf.AddMultiComment(v.Cursor.Loc, v.Cursor.Loc)
	v.Cursor.Start()
	return true
}

// FindFunctionDeclaration Find function and show full file on split window
func (v *View) FindFunctionDeclaration(usePlugin bool) bool {
	messenger.Message("")
	ok, where, word, line := v.SearchFunction(false)
	if !ok {
		messenger.Warning("function not found : ", word)
		return true
	}
	v.VSplit(v.Buf)
	if where == "b" {
		CurView().Cursor.GotoLoc(Loc{0, line})
		CurView().Center(false)
		CurView().PreviousSplit(false)
		v.Cursor.Loc = v.savedLoc
		v.NextSplit(false)
		messenger.Success("function found in current file : ", word)
		return true
	}
	// search in files in the current directory
	CurView().Open(where)
	CurView().savedLoc = Loc{0, line}
	messenger.Success("function found on file (", where, ") : ", word)
	return true
}

// HintFunction Find function and show a small hint on window, with relevant information
func (v *View) HintFunction(usePlugin bool) bool {
	messenger.Message("")
	ok, data, word, _ := v.SearchFunction(true)
	if !ok {
		messenger.Warning("function not found : ", word)
		return true
	}
	v.OpenHelperView("h", v.Buf.Settings["filetype"].(string), data)
	return true
}

// SearchFunction :
// Search for word under cursor in the current buffer or the file system
// If searching for hint, return previous comments, and first line of function
func (v *View) SearchFunction(hint bool) (bool, string, string, int) {
	loc := v.Cursor.Loc
	v.Cursor.SelectWord(false)
	word := v.Cursor.GetSelection()
	v.Cursor.ResetSelection()
	v.Cursor.Loc = loc
	if word == "" {
		return false, "", word, 0
	}
	exp := `^\s*(?:local )?(?:func(?:tion)?|def(?:n|un|ine)?|fn|sub|let|\w+\s+(?:\(.*?\)))\s+` + word + `\s*(?:(?:\(.*?\))|\s*\W)|` + word + `\s*:?=\s*(?:func(?:tion|fn|sub)[\s\{\(])`
	if v.Buf.Settings["findfuncregex"].(string) != "" {
		exp = strings.ReplaceAll(v.Buf.Settings["findfuncregex"].(string), "%word%", word)
	}
	r := regexp.MustCompile(exp)
	//search function definition in current buffer
	line, ok := FindLineWith(r, v, v.Cursor.Loc, v.Buf.End(), false)
	if ok {
		if !hint {
			return true, "b", word, line
		}
		comment := regexp.MustCompile(`^\s*(?:#|//|(?:<!)?--|/\*)`)
		data := ""
		for l := line - 5; l < line+4; l++ {
			if l < 0 {
				continue
			} else if l > v.Buf.End().Y {
				break
			}
			d := v.Buf.Line(l)
			if l < line {
				if comment.MatchString(d) {
					data = data + d + "\n"
				}
				continue
			}
			data = data + d + "\n"
		}
		return true, data, word, line
	}
	// search in files in the current directory
	messenger.Message("Searching in file system, wait ....")
	data, line, ok := FindFileWith(r, filepath.Dir(v.Buf.Path), v.Buf.FileType(), path.Ext(v.Buf.Fname), 2, hint)
	messenger.Message("")
	if ok {
		return true, data, word, line
	}
	return false, "", word, 0
}

// ComboKeyActive check Ctrl-k pressed
var ComboKeyActive = false

// ComboKey Concatenates special key combinations
func (v *View) ComboKey(usePlugin bool) bool {
	ComboKeyActive = !ComboKeyActive
	return true
}
