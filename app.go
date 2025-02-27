// app
package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hanspr/tcell"
)

// Opt option element of a select
type Opt struct {
	value string
	label string
	style *AppStyle
}

// AppElement an element from the frame
type AppElement struct {
	name        string // element name
	label       string // element label if it apply
	form        string // types: box, textbox, textarea, label, checkbox, radio, button
	value       string
	valueType   string                                              // string, number
	pos         Loc                                                 // Top left corner where to position element
	aposb       Loc                                                 // hotspot top,left
	apose       Loc                                                 // hotspot bottom,right
	cursor      Loc                                                 // cursor location inside text elements
	width       int                                                 // text element width, text area width
	height      int                                                 // Textbox: maxlength, text area have no predefined maxlength
	index       int                                                 // order in which to draw
	callback    func(string, string, string, string, int, int) bool // (element.name, element.value, event, when, x, y)
	style       tcell.Style                                         // color style for this element
	checked     bool                                                // Checkbox, Radio checked. Select is open
	gname       string                                              // Original name, required for radio buttons
	offset      int                                                 // Select = indexSelected, Text = offset from the age id box smaller than maxlength
	visible     bool                                                // Set element visibility attribute
	iKey        int                                                 // Used to store temporary integer value
	opts        []Opt                                               // Hold split select data
	microapp    *MicroApp
	frame       *Frame
	luacallback string
}

// AppStyle defines a style
type AppStyle struct {
	name  string
	style tcell.Style
}

// Frame defines a frame
type Frame struct {
	name      string
	visible   bool
	top       int
	left      int
	right     int
	bottom    int
	position  string
	otop      int
	oleft     int
	owidth    int
	oheight   int
	elements  map[string]*AppElement
	microapp  *MicroApp
	maxindex  int
	maxheight int
}

// MicroApp application object structure
type MicroApp struct {
	name             string
	screen           tcell.Screen
	defStyle         tcell.Style
	styles           map[string]AppStyle
	Finish           func(string)
	WindowMouseEvent func(string, int, int) // event, x, y
	WindowKeyEvent   func(string, int, int) // key, x, y
	WindowFinish     func(map[string]string)
	cursor           Loc // Relative position inside frame
	activeElement    string
	lockActive       bool
	mousedown        bool
	lastloc          Loc
	lastbutton       string
	lastclick        time.Time
	frames           map[string]*Frame
	activeFrame      string
	mouseOver        string
	debug            bool
}

// ------------------------------------------------
// App Build Methods
// ------------------------------------------------

// AddFrame creates a new frame object
func (a *MicroApp) AddFrame(name string, top, left, width, height int, position string) *Frame {
	var f Frame

	f.elements = make(map[string]*AppElement)
	w, h := a.screen.Size()
	f.otop = top
	f.oleft = left
	f.owidth = width
	f.oheight = height
	if top+left+width+height == 0 {
		width = w
		height = h
	} else {
		if left < 0 {
			left = w/2 - width/2
		}
		if top < 0 {
			top = h/2 - height/2
		}
	}
	if position != "fixed" && position != "relative" && position != "close" {
		position = "close"
	}
	f.top = top
	f.left = left
	f.right = left + width
	f.bottom = top + height
	f.position = position
	f.visible = false
	f.name = name
	f.microapp = a
	f.maxheight = height
	f.maxindex = 2
	if a.activeFrame == "" {
		a.activeFrame = name
		f.visible = true
	}
	a.frames[name] = &f
	return &f
}

// ChangeFrame modify frame data
func (f *Frame) ChangeFrame(top, left, width, height int, position string) {
	a := f.microapp
	w, h := a.screen.Size()
	f.otop = top
	f.oleft = left
	f.owidth = width
	f.oheight = height
	if top+left+width+height == 0 {
		width = w
		height = h
	} else {
		if left < 0 {
			left = w/2 - width/2
		}
		if top < 0 {
			top = h/2 - height/2
		}
	}
	if position != "fixed" && position != "relative" {
		position = "relative"
	}
	f.top = top
	f.left = left
	f.right = left + width
	f.bottom = top + height
	f.position = position
}

// Show a frame
func (f *Frame) Show(v bool) {
	f.visible = v
}

// AddStyle add a new style definition to this frame
func (a *MicroApp) AddStyle(name, style string) {
	var s AppStyle

	s.name = name
	s.style = StringToStyle(style)
	a.styles[name] = s
}

// AddWindowElement add an element to a frame
func (a *MicroApp) AddWindowElement(frame, name, label, form, value, valueType string, x, y, w, h int, chk bool, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	var e AppElement

	f := a.frames[frame]
	e.name = name
	e.label = label
	e.form = form
	e.value = value
	e.valueType = valueType
	e.pos = Loc{x, y}
	e.cursor = Loc{0, 0}
	e.width = w
	e.height = h
	e.index = 1
	e.callback = callback
	e.luacallback = luacallback
	e.checked = chk
	e.offset = 0
	e.microapp = a
	e.frame = f
	e.visible = true
	re := regexp.MustCompile(`{/?\w+}`)
	label = re.ReplaceAllString(label, "")
	if style == "" {
		e.style = a.defStyle
	} else if val, ok := a.styles[style]; ok {
		e.style = val.style
	} else {
		e.style = StringToStyle(style)
	}
	if form == "box" {
		e.index = 0
		if y+h > f.maxheight {
			f.maxheight = y + h
		}
	} else if form == "radio" || form == "checkbox" {
		n := 0
		for _, e := range f.elements {
			if e.gname == name {
				n++
			}
		}
		e.gname = name
		e.name = name + strconv.Itoa(n)
	}
	// Calculate begin and end coordinates for the hotspot for element type
	r := []rune(label)
	lblwidth := len(r)
	if form == "label" || form == "button" {
		e.aposb = Loc{x, y}
		e.apose = Loc{x + lblwidth, y}
	} else if form == "radio" || form == "checkbox" {
		e.aposb = Loc{x, y}
		e.apose = Loc{x + lblwidth + 2, y}
	} else if form == "textbox" {
		e.aposb = Loc{x + lblwidth, y}
		e.apose = Loc{x + lblwidth + w - 1, y}
	} else if form == "select" {
		var vp Opt
		e.index++
		h--
		e.offset = -1
		e.cursor.Y = 0
		e.aposb = Loc{x + lblwidth, y}
		e.apose = Loc{x + lblwidth + w, y + h}
		opts := strings.Split(e.valueType, "|")
		e.cursor.X = len(opts)
		for i := 0; i < len(opts); i++ {
			opt := strings.Split(opts[i], "]")
			if len(opt) == 1 {
				opt = append(opt, opt[0])
			}
			re := regexp.MustCompile(`{/?(\w+)}`)
			name := re.FindStringSubmatch(opt[1])
			vp.style = nil
			if len(name) > 0 {
				if val, ok := a.styles[name[1]]; ok {
					vp.style = &val
				}
			}
			opt[1] = re.ReplaceAllString(opt[1], "")
			if opt[1] == "" {
				opt[1] = opt[0]
			}
			vp.value = opt[0]
			vp.label = opt[1]
			e.opts = append(e.opts, vp)
			if opt[0] == e.value {
				e.offset = i
			}
		}
		e.valueType = ""
	} else if form == "textarea" {
		e.aposb = Loc{x + 1, y + 1}
		e.apose = Loc{x + w - 1, y + h - 1}
	}
	a.frames[frame].elements[e.name] = &e
}

// AddWindowBox add a Box object to frame
func (f *Frame) AddWindowBox(name, title string, x, y, width, height int, border bool, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	if width < 1 || height < 1 {
		return
	}
	a.AddWindowElement(f.name, name, title, "box", "", "", x, y, width, height, border, callback, style, luacallback)
}

// AddWindowLabel add a label to a frame
func (f *Frame) AddWindowLabel(name, label string, x, y int, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	a.AddWindowElement(f.name, name, label, "label", "", "", x, y, 0, 0, false, callback, style, luacallback)
}

// AddWindowMenuLabel add a label for menu drawing
func (f *Frame) AddWindowMenuLabel(name, label, kind string, x, y int, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	if kind == "r" {
		a.AddWindowElement(f.name, name, " "+label+string(tcell.RuneVLine), "label", "", "", x, y, 0, 0, false, callback, style, luacallback)
	} else if kind == "l" {
		a.AddWindowElement(f.name, name, string(tcell.RuneVLine)+label+" ", "label", "", "", x, y, 0, 0, false, callback, style, luacallback)
	} else if kind == "cl" {
		a.AddWindowElement(f.name, name, string('┐')+label+string(tcell.RuneVLine), "label", "", "", x, y, 0, 0, false, callback, style, luacallback)
	} else {
		a.AddWindowElement(f.name, name, string(tcell.RuneVLine)+label+string(tcell.RuneVLine), "label", "", "", x, y, 0, 0, false, callback, style, luacallback)
	}
}

// AddWindowMenuTop add top of menu label with border
func (f *Frame) AddWindowMenuTop(name, label string, x, y int, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	label = strings.ReplaceAll(label, " ", string(tcell.RuneHLine))
	a.AddWindowElement(f.name, name, string(tcell.RuneLLCorner)+label+string(tcell.RuneURCorner), "label", "", "", x, y, 0, 0, false, callback, style, luacallback)
}

// AddWindowMenuBottom add bottom menu label with border
func (f *Frame) AddWindowMenuBottom(name, label string, x, y int, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	label = strings.ReplaceAll(label, " ", string(tcell.RuneHLine))
	a.AddWindowElement(f.name, name, string(tcell.RuneLLCorner)+label+string(tcell.RuneLRCorner), "label", "", "", x, y, 0, 0, false, callback, style, luacallback)
}

// AddWindowTextBox add a textbox object to frame
func (f *Frame) AddWindowTextBox(name, label, value, valueType string, x, y, width, maxlength int, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	if width < 1 {
		return
	} else if width <= 3 && maxlength > 3 {
		maxlength = width
	}
	a.AddWindowElement(f.name, name, label, "textbox", value, valueType, x, y, width, maxlength, false, callback, style, luacallback)
}

// AddWindowCheckBox add a checkbox to the frame
func (f *Frame) AddWindowCheckBox(name, label, value string, x, y int, checked bool, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	a.AddWindowElement(f.name, name, label, "checkbox", value, "", x, y, 0, 0, checked, callback, style, luacallback)
}

// AddWindowRadio add a radio element to the frame
func (f *Frame) AddWindowRadio(name, label, value string, x, y int, checked bool, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	a.AddWindowElement(f.name, name, label, "radio", value, "", x, y, 0, 0, checked, callback, style, luacallback)
}

// AddWindowTextArea add a textarea element to the frame
func (f *Frame) AddWindowTextArea(name, label, value string, x, y, columns, rows int, readonly bool, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	if columns < 5 || rows < 2 {
		return
	}
	a.AddWindowElement(f.name, name, label, "textarea", value, "", x, y, columns+2, rows+2, readonly, callback, style, luacallback)
}

// AddWindowSelect add a select element to the frame
func (f *Frame) AddWindowSelect(name, label, value string, options string, x, y, width, height int, callback func(string, string, string, string, int, int) bool, style, luacallback string) bool {
	a := f.microapp
	if options == "" {
		return false
	}
	if width == 0 {
		opts := strings.Split(options, "|")
		for _, opt := range opts {
			if strings.Contains(opt, "]") {
				v := strings.Split(opt, "]")
				if v[1] == "" {
					v[1] = v[0]
				}
				if Count(v[1]) > width {
					width = Count(v[1])
				}
			} else {
				if Count(opt) > width {
					width = Count(opt)
				}
			}
		}
	}
	a.AddWindowElement(f.name, name, label, "select", value, options, x, y, width, height, false, callback, style, luacallback)
	return true
}

// AddWindowButton add a button to the frame
func (f *Frame) AddWindowButton(name, label, buttonType string, x, y int, callback func(string, string, string, string, int, int) bool, style, luacallback string) {
	a := f.microapp
	a.AddWindowElement(f.name, name, label, "button", "", buttonType, x, y, 0, 0, false, callback, style, luacallback)
}

// ------------------------------------------------
// Element Methods
// ------------------------------------------------

// SetIndex set the index layer property of an element (0 - 5 layers)
func (f *Frame) SetIndex(k string, v int) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	if v > 5 {
		v = 5
	} else if v < 0 {
		v = 0
	}
	if v > f.maxindex {
		f.maxindex = v
	}
	e.index = v
}

// GetVisible get visible property of an element
func (f *Frame) GetVisible(k string) bool {
	e, ok := f.elements[k]
	if !ok {
		return false
	}
	return e.visible
}

// SetVisible set visible property of an element
func (f *Frame) SetVisible(k string, v bool) {
	a := f.microapp
	e, ok := f.elements[k]
	if !ok {
		return
	}
	if e.visible == v {
		return
	}
	e.visible = v
	if v {
		e.Draw()
	} else {
		a.ResetFrames()
		if a.activeElement == "" {
			a.cursor = Loc{0, 0}
			a.screen.HideCursor()
		} else {
			a.screen.ShowCursor(a.cursor.X+f.left, a.cursor.Y+f.top)
		}
	}
	a.screen.Show()
}

// GetgName get group name of this element
func (f *Frame) GetgName(k string) string {
	e, ok := f.elements[k]
	if !ok {
		return ""
	}
	return e.gname
}

// SetgName Set a group name for this element
func (f *Frame) SetgName(k, v string) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	e.gname = v
}

// GetiKey get the integer user value to this element
func (f *Frame) GetiKey(k string) int {
	e, ok := f.elements[k]
	if !ok {
		return 0
	}
	return e.iKey
}

// SetiKey set the user integer value of this element
func (f *Frame) SetiKey(k string, v int) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	e.iKey = v
}

// GetValue get elements value property
func (f *Frame) GetValue(k string) string {
	e, ok := f.elements[k]
	if !ok {
		return ""
	}
	return e.value
}

// SetValue set element value property
func (f *Frame) SetValue(k, v string) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	a := f.microapp
	e.value = v
	if e.form == "select" {
		for i := 0; i < len(e.opts); i++ {
			if e.opts[i].value == v {
				e.offset = i
				break
			}
		}
	} else if e.form == "textbox" || e.form == "textarea" {
		f.SetFocus(k, "end")
	}
	e.Draw()
	a.screen.Show()
}

// InsertStringAt Inserts string in current value at position X
func (e *AppElement) InsertStringAt(x int, clip string) {
	a := e.microapp
	var nr []rune
	if len(clip)+len(e.value) > e.height {
		clip = SubstringSafe(clip, 0, e.height-len(e.value)-3)
	}
	r := []rune(e.value)
	cr := []rune(clip)
	nr = append(nr, r[:e.cursor.X]...)
	nr = append(nr, cr...)
	nr = append(nr, r[e.cursor.X:]...)
	e.value = string(nr)
	e.cursor.X += len(cr)
	a.cursor.X += len(cr)
	e.DrawTextBox()
	a.screen.ShowCursor(a.cursor.X+e.frame.left, a.cursor.Y+e.frame.top)
	a.screen.Show()
}

// GetChecked get checked property
func (f *Frame) GetChecked(k string) bool {
	e, ok := f.elements[k]
	if !ok {
		return false
	}
	return e.checked
}

// SetChecked set checked property
func (f *Frame) SetChecked(k string, v bool) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	a := f.microapp
	e.checked = v
	e.Draw()
	a.screen.Show()
}

// GetPos get the element coordinates of this element
func (f *Frame) GetPos(k string) Loc {
	e, ok := f.elements[k]
	if !ok {
		return Loc{0, 0}
	}
	return e.pos
}

// SetPos set a new position for this element
func (f *Frame) SetPos(k string, v Loc) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	a := f.microapp
	e.pos = v
	e.Draw()
	a.screen.Show()
}

// GetLabel set the label property of an element
func (f *Frame) GetLabel(k string) string {
	e, ok := f.elements[k]
	if !ok {
		return ""
	}
	return e.label
}

// SetLabel set the label property of an element
func (f *Frame) SetLabel(k, v string) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	a := f.microapp
	re, _ := regexp.Compile(`\{\/?\w+\}`)
	vx := re.ReplaceAllString(v, "")
	vx = strings.TrimRight(vx, " ")
	lx := re.ReplaceAllString(e.label, "")
	lx = strings.TrimRight(lx, " ")
	if Count(vx) < Count(lx) {
		// First overwrite n spaces to erease current label
		e.label = strings.Repeat(" ", Count(lx))
		e.Draw()
	}
	// Now add new label
	e.label = v
	e.Draw()
	a.screen.Show()
}

// SetFocus set the cursor over this element
func (f *Frame) SetFocus(k, where string) {
	e, ok := f.elements[k]
	if !ok {
		return
	}
	a := f.microapp
	if f.elements[k].index == 0 || (f.elements[k].form != "textbox" && f.elements[k].form != "textarea") {
		return
	}
	a.activeElement = k
	a.cursor = e.aposb
	if where == "B" {
		where = "Home"
	} else {
		where = "End"
	}
	if e.form == "textbox" {
		e.TextBoxKeyEvent(where, e.aposb.X, e.aposb.Y)
	} else {
		e.TextAreaKeyEvent(where, e.aposb.X, e.aposb.Y)
	}
}

// SetFocusPreviousInputElement move cursor to previous text element
func (f *Frame) SetFocusPreviousInputElement(k string) {
	var next AppElement
	var last AppElement

	me, ok := f.elements[k]
	if !ok {
		return
	}
	last.pos.Y = -1
	next.pos.Y = -1
	next.pos.X = -1
	for _, e := range f.elements {
		if e.index == 0 || e.name == me.name || (e.form != "textbox" && e.form != "textarea") {
			continue
		}
		if e.pos.Y == me.pos.Y && e.pos.X < me.pos.X {
			next = *e
			break
		}
		if e.pos.Y < me.pos.Y && e.pos.Y >= next.pos.Y && e.pos.X >= next.pos.X {
			next = *e
			continue
		}
		if e.pos.Y > last.pos.Y {
			last = *e
			continue
		}
		if e.pos.Y == last.pos.Y && e.pos.X > last.pos.X {
			last = *e
			continue
		}
	}
	if next.name != "" {
		f.SetFocus(next.name, "E")
	} else if last.name != "" {
		f.SetFocus(last.name, "E")
	}
}

// SetFocusNextInputElement move cursor to next text element
func (f *Frame) SetFocusNextInputElement(k string) {
	var next AppElement
	var first AppElement

	me, ok := f.elements[k]
	if !ok {
		return
	}
	first.pos.Y = 99999
	next.pos.Y = 99999
	next.pos.X = 99999
	for _, e := range f.elements {
		if e.index == 0 || e.name == me.name || (e.form != "textbox" && e.form != "textarea") {
			continue
		}
		if e.pos.Y == me.pos.Y && e.pos.X > me.pos.X {
			next = *e
			break
		}
		if e.pos.Y > me.pos.Y && e.pos.Y <= next.pos.Y && e.pos.X <= next.pos.X {
			next = *e
			continue
		}
		if e.pos.Y < first.pos.Y {
			first = *e
			continue
		}
		if e.pos.Y == first.pos.Y && e.pos.X < first.pos.X {
			first = *e
			continue
		}
	}
	if next.name != "" {
		f.SetFocus(next.name, "E")
	} else if first.name != "" {
		f.SetFocus(first.name, "E")
	}
}

// DeleteElement remove element from frame
func (f *Frame) DeleteElement(k string) {
	delete(f.elements, k)
	f.microapp.DrawAll()
}

// ------------------------------------------------
// Drawing Methods
// ------------------------------------------------

// Element Drawing

// Draw an element into the screen
func (e *AppElement) Draw() {
	if !e.visible {
		return
	}
	if e.form == "box" {
		e.DrawBox()
	} else if e.form == "label" {
		e.DrawLabel()
	} else if e.form == "textbox" {
		e.DrawTextBox()
	} else if e.form == "radio" {
		e.DrawRadio()
	} else if e.form == "checkbox" {
		e.DrawCheckBox()
	} else if e.form == "select" {
		e.DrawSelect()
	} else if e.form == "textarea" {
		e.DrawTextArea()
	} else if e.form == "button" {
		e.DrawButton()
	}
}

// Hide not implemented, hide an element
func (e *AppElement) Hide() {
}

// DrawBox draw a box element
func (e *AppElement) DrawBox() {
	a := e.microapp
	f := e.frame
	Hborder := tcell.RuneHLine
	Vborder := tcell.RuneVLine
	ULC := tcell.RuneULCorner
	URC := tcell.RuneURCorner
	LLC := tcell.RuneLLCorner
	LRC := tcell.RuneLRCorner
	if !e.checked {
		Hborder = ' '
		Vborder = ' '
		ULC = ' '
		URC = ' '
		LLC = ' '
		LRC = ' '
	}
	x1 := e.pos.X + f.left
	y1 := e.pos.Y + f.top
	x2 := e.pos.X + f.left + e.width
	y2 := e.pos.Y + f.top + e.height
	if y2 < y1 {
		y1, y2 = y2, y1
	}
	if x2 < x1 {
		x1, x2 = x2, x1
	}

	for col := x1; col <= x2; col++ {
		a.screen.SetContent(col, y1, Hborder, nil, e.style)
		a.screen.SetContent(col, y2, Hborder, nil, e.style)
	}
	for row := y1 + 1; row < y2; row++ {
		a.screen.SetContent(x1, row, Vborder, nil, e.style)
		a.screen.SetContent(x2, row, Vborder, nil, e.style)
	}
	if y1 != y2 && x1 != x2 {
		// Only add corners if we need to
		a.screen.SetContent(x1, y1, ULC, nil, e.style)
		a.screen.SetContent(x2, y1, URC, nil, e.style)
		a.screen.SetContent(x1, y2, LLC, nil, e.style)
		a.screen.SetContent(x2, y2, LRC, nil, e.style)
	}
	for row := y1 + 1; row < y2; row++ {
		for col := x1 + 1; col < x2; col++ {
			a.screen.SetContent(col, row, ' ', nil, e.style)
		}
	}
	if e.label != "" {
		e.frame.PrintStyle(e.label, e.pos.X+1, e.pos.Y, &e.style)
	}
}

// DrawLabel draw a label element
func (e *AppElement) DrawLabel() {
	e.frame.PrintStyle(e.label, e.pos.X, e.pos.Y, &e.style)
}

// DrawTextBox draw a textbox element
func (e *AppElement) DrawTextBox() {
	var r rune

	a := e.microapp
	f := e.frame
	val := []rune(e.value)
	if e.offset > 0 || (e.height > e.width && len(val) >= e.width) {
		// possible overflow, find offset
		if a.cursor.X == e.apose.X {
			e.offset = e.cursor.X - e.width + 1
		} else if a.cursor.X == e.aposb.X && e.offset > 0 {
			e.offset--
			if e.cursor.X-e.width+1 > 0 {
				e.offset++
			}
		}
		if e.offset < 0 || (a.cursor.X == e.aposb.X && e.offset > 0 && len(val) <= e.width) {
			if e.offset > 0 && a.cursor.X == e.aposb.X {
				a.cursor.X = e.aposb.X + e.cursor.X
				if a.cursor.X > e.apose.X {
					a.cursor.X = e.apose.X
					e.cursor.X--
				}
			}
			e.offset = 0
		}
	}
	e.frame.PrintStyle(e.label, e.pos.X, e.pos.Y, &e.style)
	style := e.style.Underline(true)
	for W := 0; W < e.width; W++ {
		if W+e.offset < len(val) {
			r = val[W+e.offset]
		} else {
			r = ' '
		}
		a.screen.SetContent(e.aposb.X+W+f.left, e.pos.Y+f.top, r, nil, style)
	}
}

// DrawRadio draw a radio element
func (e *AppElement) DrawRadio() {
	var radio string
	s := tcell.Style.Foreground(e.microapp.defStyle, tcell.Color168)

	radio = "◎ "
	if e.checked {
		s = tcell.Style.Foreground(e.microapp.defStyle, tcell.Color107)
		radio = "◉ "
	}
	e.frame.PrintStyle(radio, e.pos.X, e.pos.Y, &s)
	e.frame.PrintStyle(e.label, e.pos.X+2, e.pos.Y, &e.style)
}

// DrawCheckBox draw a checkbox element
func (e *AppElement) DrawCheckBox() {
	s := tcell.Style.Foreground(e.microapp.defStyle, tcell.Color168)
	check := "✗ "

	if e.checked {
		s = tcell.Style.Foreground(e.microapp.defStyle, tcell.Color107)
		check = "✔ "
	}
	e.frame.PrintStyle(check, e.pos.X, e.pos.Y, &s)
	e.frame.PrintStyle(e.label, e.pos.X+2, e.pos.Y, &e.style)
}

// DrawSelect draw a select element
func (e *AppElement) DrawSelect() {
	var style tcell.Style

	a := e.microapp
	f := e.frame
	Y := -1
	chr := ""
	ft := "%-" + strconv.Itoa(e.width) + "s"
	e.frame.PrintStyle(e.label, e.pos.X, e.pos.Y, &e.style)
	if e.offset >= 0 && e.offset < e.cursor.Y && e.height > 1 {
		e.cursor.Y = e.offset
	} else if e.height > 1 && e.offset >= e.height+e.cursor.Y {
		// Overflow, find the starting point
		if e.offset >= e.height+e.cursor.Y {
			e.cursor.Y = e.offset - e.height + 1
		}
		if e.cursor.Y >= len(e.opts) {
			e.cursor.Y = len(e.opts) - 1
		}
	}
	start := e.cursor.Y
	for i := start; i < len(e.opts); i++ {
		if i == e.offset {
			if a.activeElement == e.name {
				// Show active
				style = style.Foreground(tcell.ColorBlack).Background(tcell.Color220)
			} else {
				// Show selected not active
				style = style.Reverse(true).Bold(true)
			}
		} else if e.height == 1 {
			continue
		} else {
			// Show with selected style
			if e.opts[i].style != nil {
				style = e.opts[i].style.style
			} else {
				style = a.defStyle.Foreground(tcell.ColorWhite).Background(tcell.Color234)
			}
		}
		if e.height == 1 {
			Y = 0
			style = style.Bold(false).Background(tcell.ColorDarkSlateGrey).Foreground(tcell.ColorWhite).Reverse(false)
			chr = "▾"
		} else {
			Y++
			if start > 0 && i == start {
				chr = "▲"
			} else if Y+1 == e.height && i+1 < len(e.opts) {
				chr = "▼"
			} else {
				chr = " "
			}
		}
		if Y >= e.height {
			break
		}
		label := []rune(fmt.Sprintf(ft, e.opts[i].label) + chr)
		for N := 0; N < len(label); N++ {
			a.screen.SetContent(e.aposb.X+N+f.left, e.aposb.Y+Y+f.top, label[N], nil, style)
		}
		if e.height == 1 {
			return
		}
	}
}

// DrawButton draw a button element
func (e *AppElement) DrawButton() {
	f := e.frame
	style := e.style
	label := []rune(" " + e.label + " ")
	if e.valueType == "cancel" {
		style = e.style.Background(tcell.ColorDarkRed).Foreground(tcell.ColorWhite).Bold(true)
	} else if e.valueType == "ok" {
		style = e.style.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite).Bold(true)
	}
	for x := 0; x < len(label); x++ {
		e.microapp.screen.SetContent(e.pos.X+x+f.left, e.pos.Y+f.top, label[x], nil, style)
	}
}

// DrawTextArea draw a text area element
func (e *AppElement) DrawTextArea() {
	y := e.aposb.Y
	e.DrawBox()
	str := e.value
	str = WordWrap(str, e.width-1)
	lines := strings.Split(str, "\\N")
	for _, line := range lines {
		e.frame.Print(line, e.aposb.X, y, nil)
		y++
		if y > e.apose.Y {
			break
		}
	}
}

// WordWrap utility to add new lines to string, to implement wrapping
func WordWrap(str string, w int) string {
	nstr := str
	if Count(str) <= w {
		return str
	}
	lastspc := runeLastIndex(str, " ")
	for lastspc >= w {
		nstr = nstr[:lastspc-1]
		lastspc = runeLastIndex(nstr, " ")
	}
	if lastspc < 0 {
		lastspc = w + 1
	}
	rstr := []rune(str)
	str1 := string(rstr[:lastspc+1])
	str2 := string(rstr[lastspc+1:])
	return str1 + "\\N" + WordWrap(str2, w)
}

// Get the cursor position from an absolute cursor position
func (e *AppElement) getECursorFromACursor() int {
	a := e.microapp
	f := e.frame
	X := 0
	Y := 0
	ac := 0
	offset := a.cursor.X - e.aposb.X
	width := e.width - 1
	str := WordWrap(e.value, width)
	lines := strings.Split(str, "\\N")
	for i, line := range lines {
		Y = i
		X = Count(line)
		if i+e.aposb.Y == a.cursor.Y {
			newx := ac + offset
			if offset > X {
				newx = newx - Abs(X-offset)
				a.cursor.X = e.aposb.X + X
			}
			a.screen.ShowCursor(a.cursor.X+f.left, a.cursor.Y+f.top)
			return newx
		}
		ac = ac + X
	}
	a.cursor.Y = e.aposb.Y + Y
	a.cursor.X = e.aposb.X + X
	a.screen.ShowCursor(a.cursor.X+f.left, a.cursor.Y+f.top)
	return ac
}

// Set cursor absolute position from an element cursor position
func (e *AppElement) setACursorFromECursor() {
	a := e.microapp
	f := e.frame
	ex := e.cursor.X
	ax := 0
	ay := 0
	ac := 0
	width := e.width - 1
	str := WordWrap(e.value, width)
	lines := strings.Split(str, "\\N")
	for i, line := range lines {
		ay = i
		ll := Count(line)
		ac = ac + ll
		ax = e.aposb.X + ll
		if ex <= ac {
			ax = e.aposb.X + ll - Abs(ac-ex)
			break
		}
	}
	if ax > e.apose.X {
		ax = e.apose.X
	}
	a.cursor.X = ax
	a.cursor.Y = ay + e.aposb.Y
	a.screen.ShowCursor(a.cursor.X+f.left, a.cursor.Y+f.top)
}

func runeLastIndex(str, s string) int {
	rstr := []rune(str)
	rs := []rune(s)
	last := len(rstr) - 1
	for i := last; i >= 0; i-- {
		if rstr[i] == rs[0] {
			return i
		}
	}
	return -1
}

// ------------------------------------------------
// MicroApp Drawing Methods
// ------------------------------------------------

// DrawAll draw all frames and its elements
func (a *MicroApp) DrawAll() {
	// Draw all elements in index order from 0 to current max index (normally 2)
	for _, f := range a.frames {
		if f.visible {
			for i := 0; i <= f.maxindex; i++ {
				for _, e := range f.elements {
					if e.index == i {
						e.Draw()
					}
				}
			}
		}
	}
	if a.activeElement == "" {
		a.screen.HideCursor()
	}
	a.screen.Show()
}

// Resize handle resize events
func (a *MicroApp) Resize() {
	for _, t := range tabs {
		t.Resize()
	}
	for _, f := range a.frames {
		if f.position == "relative" {
			w, h := a.screen.Size()
			if f.otop+f.oleft+f.owidth+f.oheight == 0 {
				f.right = w
				f.bottom = h
			} else {
				if f.oleft < 0 {
					f.left = w/2 - f.owidth/2
				}
				if f.otop < 0 {
					f.top = h/2 - f.oheight/2
				}
				f.right = f.left + f.owidth
				f.bottom = f.top + f.oheight
			}
		} else if f.position == "close" {
			MicroAppStop()
			RedrawAll(true)
			return
		}
	}
	RedrawAll(false)
	a.DrawAll()
	if a.activeElement == "" {
		a.cursor = Loc{0, 0}
		a.screen.HideCursor()
	} else {
		e := a.GetActiveElement(a.activeElement)
		f := e.microapp.frames[e.frame.name]
		a.screen.ShowCursor(a.cursor.X+f.left, a.cursor.Y+f.top)
	}
	a.screen.Show()
}

// Clear Routines

// ClearScreen Clear the entire screen from any app drawings
func (a *MicroApp) ClearScreen() {
	RedrawAll(false)
	a.screen.HideCursor()
	a.screen.Show()
}

// ResetFrames Redraw Frame
func (a *MicroApp) ResetFrames() {
	RedrawAll(false)
	a.screen.HideCursor()
	a.DrawAll()
}

// Print Routines

// Debug works with absolute coordinates
func (a *MicroApp) Debug(msg string, x, y int) {
	a.PrintAbsolute(msg, x, y, nil)
	a.screen.Show()
}

// Print works with relative coordinates
func (f *Frame) Print(msg string, x, y int, style *tcell.Style) {
	x += f.left
	y += f.top
	f.microapp.PrintAbsolute(msg, x, y, style)
}

// PrintStyle works with relative coordinates
func (f *Frame) PrintStyle(msg string, x, y int, estyle *tcell.Style) {
	var style *tcell.Style
	var defstyle *tcell.Style

	a := f.microapp
	if estyle == nil {
		estyle = &a.defStyle
	} else {
		defstyle = estyle
	}
	var re = new(regexp.Regexp)
	reclose := regexp.MustCompile(`{(/\w+)}`)
	ends := reclose.FindAllString(msg, -1)
	parts := regexp.MustCompile(`{/\w+}`).Split(msg, -1)
	cont := 0
	for _, part := range parts {
		if part == "" {
			continue
		}
		if len(ends) > 0 && len(ends)-1 >= cont {
			beg := ends[cont]
			beg = strings.ReplaceAll(beg, "/", "(")
			beg = strings.ReplaceAll(beg, "}", ")}")
			re = regexp.MustCompile(beg)
			cont++
		} else {
			re = regexp.MustCompile(`!trash¡`)
		}
		found := false
		var txt []string
		name := re.FindStringSubmatch(part)
		if len(name) == 0 {
			style = estyle
		} else if val, ok := a.styles[name[1]]; ok {
			found = true
			style = &val.style
		} else {
			style = estyle
		}
		if found {
			txt = re.Split(part, -1)
		} else {
			txt = append(txt, part)
		}
		if txt[0] != "" {
			f.Print(txt[0], x, y, defstyle)
			x = x + Count(txt[0])
		}
		if len(txt) > 1 {
			f.Print(txt[1], x, y, style)
			x = x + Count(txt[1])
		}
	}
}

// PrintAbsolute works with absolute coordinates
func (a *MicroApp) PrintAbsolute(msg string, x, y int, style *tcell.Style) {
	if style == nil {
		style = &a.defStyle
	}
	msgr := []rune(msg)
	for p := 0; p < len(msgr); p++ {
		a.screen.SetContent(x+p, y, msgr[p], nil, *style)
	}
	a.screen.Show()
}

// ------------------------------------------------
// Mouse Events
// ------------------------------------------------

// TextAreaClickEvent Handle clic event
func (e *AppElement) TextAreaClickEvent(event string, x, y int) {
	a := e.microapp
	a.activeElement = e.name
	a.cursor.X = x
	a.cursor.Y = y
	e.cursor.X = e.getECursorFromACursor()
	a.screen.Show()
}

// TextBoxClickEvent handle clic event
func (e *AppElement) TextBoxClickEvent(event string, x, y int) {
	a := e.microapp
	f := e.frame
	a.activeElement = e.name
	w := Count(e.value)
	if x > e.aposb.X+w {
		x = e.aposb.X + w
		a.cursor.X = x
	}
	e.cursor.X = e.offset + a.cursor.X - e.aposb.X
	a.screen.ShowCursor(x+f.left, y+f.top)
	a.screen.Show()
}

// SelectClickEvent handle select click event
func (e *AppElement) SelectClickEvent(event string, x, y int) {
	a := e.microapp

	f := e.frame
	a.activeElement = e.name
	// SELECT COMBO BEGIN
	if e.height == 1 && !e.checked {
		// Open select. Set height, and new hotspot, save
		e.height = len(e.opts) + 1
		e.apose = Loc{e.apose.X, e.apose.Y + e.height - 1}
		e.cursor.Y = 0
		e.checked = true
		// Lock events to this element until closed
		a.lockActive = true
		a.activeElement = e.name
		a.DrawAll()
		return
	} else if e.checked {
		// Close select. Find element being clicked and saved as selected value
		for i := 0; i < e.cursor.X; i++ {
			if e.aposb.Y+i == y {
				if i < e.cursor.X {
					e.value = e.opts[i].value
					e.offset = i
				}
				break
			}
		}

		// Reset to height=1, hotspot, save
		if e.aposb.Y+e.height > f.maxheight {
			RedrawAll(false)
		}
		e.height = 1
		e.apose = Loc{e.apose.X, e.aposb.Y}
		// release event locking
		a.lockActive = false
		a.activeElement = ""
		e.checked = false
		// Needs to be all to erase list
		a.DrawAll()
		return
	}
	// SELECT END

	// LIST
	for i := 0; i < e.cursor.X; i++ {
		if e.aposb.Y+i == y {
			e.offset = i + e.cursor.Y
			e.value = e.opts[e.offset].value
			break
		}
	}
	e.Draw()
	a.screen.Show()
}

// RadioCheckboxClickEvent handle radio and checkboxes clic events
func (e *AppElement) RadioCheckboxClickEvent(event string, x, y int) {
	a := e.microapp
	f := a.frames[e.frame.name]
	if e.form == "radio" && e.checked {
		return
	}
	if e.checked {
		e.checked = false
	} else {
		e.checked = true
	}
	if e.form == "radio" {
		e.DrawRadio()
		for _, r := range f.elements {
			if r.form == "radio" && r.gname == e.gname && r.name != e.name && e.checked {
				r.checked = false
				r.DrawRadio()
			}
		}
	} else {
		e.DrawCheckBox()
	}
	a.screen.Show()
}

// ProcessElementClick find what kind of element received a click event
func (e *AppElement) ProcessElementClick(event string, x, y int) {
	a := e.microapp
	name := e.name
	if a.activeElement != "" {
		a.screen.HideCursor()
		a.DrawAll()
	}
	check := false
	if e.form == "checkbox" || e.form == "radio" {
		name = e.gname
		check = true
	}
	if e.callback != nil {
		if !e.callback(name, e.value, event, "PRE", x, y) {
			return
		}
	} else if e.luacallback != "" {
		v, err := Call(e.luacallback, name, e.value, event, "PRE", x, y)
		if err == nil && v.String() == "false" {
			return
		}
	}
	if event == "mouse-click1" {
		a.activeElement = name
		if check {
			e.RadioCheckboxClickEvent(event, x, y)
		} else if e.form == "textbox" {
			e.TextBoxClickEvent(event, x, y)
		} else if e.form == "textarea" {
			e.TextAreaClickEvent(event, x, y)
		} else if e.form == "select" {
			e.SelectClickEvent(event, x, y)
		} else {
			a.activeElement = ""
		}
	}
	if e.callback != nil || e.luacallback != "" {
		value := e.value
		if check {
			value = "false"
			if e.checked {
				value = "true"
			}
		}
		if e.callback != nil {
			e.callback(name, value, event, "POST", x, y)
		} else {
			Call(e.luacallback, name, value, event, "POST", x, y)
		}
	}
}

// ProcessElementMouseMove find element to send a mouse move event
func (e *AppElement) ProcessElementMouseMove(event string, x, y int) {
	if e.callback != nil {
		if !e.callback(e.name, e.value, event, "", x, y) {
			return
		}
	} else if e.luacallback != "" {
		v, err := Call(e.luacallback, e.name, e.value, event, "", x, y)
		if err == nil && v.String() == "false" {
			return
		}
	}
}

// ProcessElementMouseDown not implemented
func (e *AppElement) ProcessElementMouseDown(event string, x, y int) {
}

// SelectWheelEvent dispatch event to element
func (e *AppElement) SelectWheelEvent(event string, x, y int) {
	a := e.microapp
	if event == "mouse-wheelUp" {
		if e.offset < 1 {
			return
		}
		e.offset--
	} else if event == "mouse-wheelDown" {
		if e.cursor.X <= e.offset+1 {
			return
		}
		e.offset++
	}
	e.value = e.opts[e.offset].value
	e.Draw()
	a.screen.Show()
}

// ProcessElementMouseWheel dispatch mousewheel event to element
func (e *AppElement) ProcessElementMouseWheel(event string, x, y int) {
	if e.form == "select" {
		e.SelectWheelEvent(event, x, y)
	}
}

// ------------------------------------------------
// Key Events
// ------------------------------------------------

func (e *AppElement) SetNavModeBindings(key string) string {
	// Navmode binding
	if key == "i" {
		key = "Up"
	} else if key == "k" {
		key = "Down"
	} else if key == "y" {
		key = "PgUp"
	} else if key == "l" {
		key = "Right"
	} else if key == "j" {
		key = "Left"
	} else if key == "h" {
		key = "PgDn"
	} else if key == "u" {
		key = "Home"
	} else if key == "o" {
		key = "End"
	}
	return key
}

// SelectKeyEvent handle key event
func (e *AppElement) SelectKeyEvent(key string, x, y int) {
	key = e.SetNavModeBindings(key)
	a := e.microapp
	f := e.frame
	r := []rune(key)
	if len(r) <= 1 {
		return
	}
	// Process Control Keys
	if key == "Up" {
		if e.offset < 1 {
			return
		}
		e.offset--
	} else if key == "Down" {
		if e.cursor.X <= e.offset+1 {
			return
		}
		e.offset++
	} else if key == "PgUp" {
		e.offset -= e.height / 2
		if e.offset < 0 {
			e.offset = 0
		}
	} else if key == "PgDn" {
		e.offset += e.height / 2
		if e.offset >= e.cursor.X {
			e.offset = e.cursor.X - 1
		}
	} else if key == "Home" {
		e.offset = 0
	} else if key == "End" {
		e.offset = e.cursor.X - 1
	} else if key == "Enter" {
		a.activeElement = ""
		if a.lockActive || e.checked {
			if e.aposb.Y+e.height > f.maxheight {
				RedrawAll(false)
			}
			a.activeElement = ""
			e.height = 1
			e.apose = Loc{e.apose.X, e.aposb.Y}
			a.lockActive = false
			e.checked = false
			a.DrawAll()
		}
	}
	e.value = e.opts[e.offset].value
	e.Draw()
	a.screen.Show()
}

// TextAreaKeyEvent handle key event
func (e *AppElement) TextAreaKeyEvent(key string, x, y int) {
	a := e.microapp
	if e.apose.Y == y && e.apose.X == x {
		return
	}
	r := []rune(key)
	b := []rune(e.value)
	if len(r) > 1 {
		// Process Control Keys
		if key == "Backspace2" {
			if e.cursor.X <= 0 {
				return
			}
			e.cursor.X--
			b = a.removeCharAt(b, e.cursor.X)
			e.value = string(b)
		} else if key == "Delete" || key == "Ctrl+U" {
			b = a.removeCharAt(b, e.cursor.X)
			e.value = string(b)
		} else if key == "Left" || key == "Alt+j" {
			if e.cursor.X-1 < 0 {
				return
			}
			e.cursor.X--
			e.setACursorFromECursor()
		} else if key == "Right" || key == "Alt+l" {
			if e.cursor.X+1 > len(b) || (x+1 > e.apose.X && y == e.apose.Y) {
				return
			}
			e.cursor.X++
			e.setACursorFromECursor()
		} else if key == "Up" || key == "Alt+i" {
			if a.cursor.Y-1 < e.aposb.Y {
				return
			}
			a.cursor.Y--
			e.cursor.X = e.getECursorFromACursor()
			a.screen.Show()
			return
		} else if key == "Down" || key == "Alt+k" {
			if a.cursor.Y+1 > e.apose.Y {
				return
			}
			a.cursor.Y++
			e.cursor.X = e.getECursorFromACursor()
			a.screen.Show()
			return
		} else if key == "Home" || key == "Alt+u" {
			a.cursor.X = e.aposb.X
			e.cursor.X = 0
			a.cursor.Y = e.aposb.Y
		} else if key == "End" || key == "Alt+o" {
			e.cursor.X = len(b)
		} else if key == "Enter" {
			return
		} else if key == "Ctrl+V" {
			clip := Clip.ReadFrom("local", "clip")
			e.value = e.value + clip
			e.TextAreaKeyEvent("End", x, y)
			return
		} else if key == "Ctrl+R" || key == "Ctrl+J" {
			e.value = ""
			a.cursor.X = e.aposb.X
			e.cursor.X = 0
			a.cursor.Y = e.aposb.Y
		}
		e.Draw()
		e.setACursorFromECursor()
		a.screen.Show()
		return
	}
	e.value = string(a.insertCharAt(b, r, e.cursor.X))
	e.cursor.X++
	e.DrawTextArea()
	e.setACursorFromECursor()
	a.screen.Show()
}

// TextBoxKeyEvent handle key event
func (e *AppElement) TextBoxKeyEvent(key string, x, y int) {
	a := e.microapp
	f := e.frame
	maxlength := e.height
	r := []rune(key)
	b := []rune(e.value)
	if len(r) > 1 {
		// Process Control Keys
		if key == "Backspace2" {
			if e.cursor.X <= 0 {
				return
			}
			e.cursor.X--
			if a.cursor.X-1 >= e.aposb.X {
				a.cursor.X--
			}
			b = a.removeCharAt(b, e.cursor.X)
			e.value = string(b)
		} else if key == "Delete" || key == "Ctrl+U" {
			b = a.removeCharAt(b, e.cursor.X)
			e.value = string(b)
		} else if key == "Left" || key == "Alt+j" {
			if e.cursor.X-1 < 0 {
				return
			}
			if a.cursor.X-1 >= e.aposb.X {
				a.cursor.X--
			}
			e.cursor.X--
		} else if key == "Right" || key == "Alt+l" {
			if e.cursor.X < maxlength-1 && e.cursor.X < len(b) {
				e.cursor.X++
				if a.cursor.X+1 <= e.apose.X {
					a.cursor.X++
				}
			}
		} else if key == "Home" || key == "Alt+u" {
			a.cursor.X = e.aposb.X
			e.cursor.X = 0
			e.offset = 0
		} else if key == "End" || key == "Alt+o" {
			if len(b) == maxlength {
				e.cursor.X = len(b) - 1
				a.cursor.X = e.aposb.X + len(b) - 1
			} else {
				e.cursor.X = len(b)
				a.cursor.X = e.aposb.X + len(b)
			}
			if a.cursor.X > e.apose.X {
				a.cursor.X = e.apose.X
			}
		} else if key == "Enter" {
			return
		} else if key == "Tab" {
			f.SetFocusNextInputElement(e.name)
			return
		} else if key == "Backtab" {
			f.SetFocusPreviousInputElement(e.name)
			return
		} else if key == "Ctrl+V" {
			clip := Clip.ReadFrom("local", "cip")
			e.InsertStringAt(e.cursor.X, clip)
		} else if key == "Ctrl+R" || key == "Ctrl+J" {
			e.value = ""
			a.cursor.X = e.aposb.X
			e.cursor.X = 0
			e.offset = 0
		}
		e.DrawTextBox()
		a.screen.ShowCursor(a.cursor.X+f.left, a.cursor.Y+f.top)
		a.screen.Show()
		return
	}
	// Process Rune
	if len(b) >= maxlength {
		return
	}
	e.value = string(a.insertCharAt(b, r, e.cursor.X))
	if e.cursor.X < maxlength-1 && e.cursor.X <= len(b) {
		e.cursor.X++
		if a.cursor.X+1 <= e.apose.X {
			a.cursor.X++
		}
	}
	e.DrawTextBox()
	a.screen.ShowCursor(a.cursor.X+f.left, a.cursor.Y+f.top)
	a.screen.Show()
}

// ProcessElementKey dispatch an event to the corresponding element
func (e *AppElement) ProcessElementKey(key string, x, y int) {
	a := e.microapp
	if a.activeElement == "" || a.activeElement != e.name {
		// This should not happen, for safety
		a.screen.HideCursor()
		return
	}
	if e.callback != nil {
		if !e.callback(e.name, e.value, key, "PRE", x, y) {
			return
		}
	} else if e.luacallback != "" {
		v, err := Call(e.luacallback, e.name, e.value, key, "PRE", x, y)
		if err == nil && v.String() == "false" {
			return
		}
	}
	if e.form == "textbox" {
		e.TextBoxKeyEvent(key, x, y)
	} else if e.form == "textarea" {
		e.TextAreaKeyEvent(key, x, y)
	} else if e.form == "select" {
		e.SelectKeyEvent(key, x, y)
	}
	if e.callback != nil {
		e.callback(e.name, e.value, key, "POST", x, y)
	} else if e.luacallback != "" {
		Call(e.luacallback, e.name, e.value, key, "POST", x, y)
	}
}

// Check if event occurs on top of an element hotspot
// If it does, dispatch the appropriate method

// CheckElementsActions find kind of event and pass action to the proper routine
func (a *MicroApp) CheckElementsActions(event string, x, y int) bool {
	if x < 0 || y < 0 {
		return false
	}
	for _, f := range a.frames {
		if f.visible {
			for _, e := range f.elements {
				if e.index == 0 {
					// Skip boxes
					continue
				}
				// Check if location is inside the element hotspot
				if x >= e.aposb.X && x <= e.apose.X && y >= e.aposb.Y && y <= e.apose.Y {
					if a.lockActive && a.activeElement != e.name {
						continue
					}
					if a.activeFrame != e.frame.name {
						a.activeFrame = e.frame.name
					}
					if strings.Contains(event, "mouse") {
						if strings.Contains(event, "click") {
							a.cursor = Loc{x, y}
							e.ProcessElementClick(event, x, y)
						} else if strings.Contains(event, "button") {
							e.ProcessElementMouseDown(event, x, y)
						} else if strings.Contains(event, "wheel") {
							e.ProcessElementMouseWheel(event, x, y)
						} else {
							if a.mouseOver != "" && a.mouseOver != e.name {
								ex := f.elements[a.mouseOver]
								ex.ProcessElementMouseMove("mouseout", x, y)
								e.ProcessElementMouseMove("mousein", x, y)
							} else if a.mouseOver == "" {
								e.ProcessElementMouseMove("mousein", x, y)
							} else {
								e.ProcessElementMouseMove(event, x, y)
							}
							a.mouseOver = e.name
						}
					} else {
						e.ProcessElementKey(event, x, y)
						a.mouseOver = ""
					}
					if a.debug {
						RedrawAll(false)
						a.DrawAll()
					}
					return true
				}
			}
			if a.mouseOver != "" {
				ex := f.elements[a.mouseOver]
				ex.ProcessElementMouseMove("mouseout", x, y)
				a.mouseOver = ""
			}
		}
	}
	if strings.Contains(event, "click") {
		if a.lockActive {
			e := a.frames[a.activeFrame].elements[a.activeElement]
			if e.aposb.Y+e.height > a.frames[a.activeFrame].maxheight {
				RedrawAll(false)
			}
			a.lockActive = false
			e.checked = false
			e.height = 1
			e.apose = Loc{e.apose.X, e.aposb.Y}
		}
		a.screen.HideCursor()
		if a.activeElement != "" {
			a.activeElement = ""
			a.DrawAll()
		} else {
			a.screen.Show()
		}
	}
	return false
}

// GetActiveElement return the object that is active at the time
func (a *MicroApp) GetActiveElement(name string) *AppElement {
	if name == "" {
		return nil
	}
	for _, f := range a.frames {
		for k, e := range f.elements {
			if k == name {
				return e
			}
		}
	}
	return nil
}

// ------------------------------------------------
// String methods
// ------------------------------------------------

func (a *MicroApp) removeCharAt(b []rune, i int) []rune {
	if len(b) == 0 || i >= len(b) {
		return b
	}
	copy(b[i:], b[i+1:])
	return b[:len(b)-1]
}

func (a *MicroApp) insertCharAt(b, r []rune, i int) []rune {
	var b1 []rune
	if i > len(b) {
		b = append(b, r[0])
		return b
	}
	b1 = append(b1, b[i:]...)
	b = append(b[:i], r[0])
	b = append(b, b1...)
	return b
}

// ------------------------------------------------
// Eventhandler
// ------------------------------------------------

// HandleEvents handle terminal events
func (a *MicroApp) HandleEvents(event tcell.Event) {
	char := ""
	switch ev := event.(type) {
	case *tcell.EventPaste:
		e := a.GetActiveElement(a.activeElement)
		if e != nil && e.form == "textbox" {
			e.InsertStringAt(e.cursor.X, ev.Text())
			if e.callback != nil {
				e.callback(e.name, e.value, "Paste", "POST", e.cursor.X, e.cursor.Y)
			} else if e.luacallback != "" {
				Call(e.luacallback, e.name, e.value, "Paste", "POST", e.cursor.X, e.cursor.Y)
			}
		}
	case *tcell.EventResize:
		a.Resize()
	case *tcell.EventKey:
		if ev.Key() == tcell.KeyEscape {
			if a.Finish == nil {
				MicroAppStop()
			} else {
				a.Finish("")
			}
			return
		}
		if ev.Key() == 256 && ev.Modifiers() == 0 {
			char = string(ev.Rune())
		} else {
			if ev.Modifiers() == 4 {
				char = "Alt+" + string(ev.Rune())
			} else if ev.Modifiers() == 3 {
				char = "Ctrl+" + string(ev.Rune())
			} else {
				char = ev.Name()
			}
		}
		if a.WindowKeyEvent != nil {
			a.WindowKeyEvent(char, a.cursor.X, a.cursor.Y)
		} else {
			a.CheckElementsActions(char, a.cursor.X, a.cursor.Y)
		}
	case *tcell.EventMouse:
		f := a.frames[a.activeFrame]
		exit := false
		xa, ya := ev.Position()
		if len(a.frames) > 1 {
			// Check if location inside a different frame
			for fn, fx := range a.frames {
				if fx.visible && fx.name != a.activeFrame && xa >= fx.left && xa <= fx.right && ya >= fx.top && ya <= fx.bottom {
					// Change active frame interaction
					f = fx
					a.activeFrame = fn
					break
				}
			}
		}
		x := xa - f.left
		y := ya - f.top
		button := ev.Buttons()
		action := ""
		if button == tcell.ButtonNone {
			if a.mousedown {
				// This is a possible exit
				if xa < f.left || xa > f.right || ya < f.top || ya > f.bottom {
					exit = true
				}
				Dt := time.Since(a.lastclick) / time.Millisecond
				// If button released within 2 pixel consider as clic and not a dragstop
				if Abs(a.lastloc.X-x) < 3 && Abs(a.lastloc.Y-y) < 3 {
					if Dt < 300 {
						action = "mouse-doubleclick" + a.lastbutton
					} else {
						action = "mouse-click" + a.lastbutton
						a.mousedown = false
					}
				} else {
					action = "mousemove-dragstop" + a.lastbutton
				}
				a.lastclick = time.Now()
				a.lastbutton = ""
			} else {
				action = "mousemove"
			}
		} else if ev.HasMotion() {
			action = "mousemove-drag" + a.lastbutton
		} else if !ev.HasMotion() {
			a.mousedown = true

			if strings.Count(ev.EscSeq(), "[") > 1 {
				return
			}
			if button == tcell.Button1 {
				action = "mouse-button1"
				a.lastbutton = "1"
			} else if button == tcell.Button2 {
				action = "mouse-button2"
				a.lastbutton = "2"
			} else if button == tcell.Button3 {
				action = "mouse-button3"
				a.lastbutton = "3"
			} else if button&tcell.WheelUp != 0 {
				action = "mouse-wheelUp"
				a.mousedown = false
			} else if button&tcell.WheelDown != 0 {
				action = "mouse-wheelDown"
				a.mousedown = false
			} else if button&tcell.WheelLeft != 0 {
				action = "mouse-wheelLeft"
				a.mousedown = false
			} else if button&tcell.WheelRight != 0 {
				action = "mouse-wheelRight"
				a.mousedown = false
			}
			a.lastloc = Loc{x, y}
		} else {
			action = "mousemove"
		}
		if a.WindowMouseEvent != nil {
			a.WindowMouseEvent(action, x, y)
		}
		if !a.CheckElementsActions(action, x, y) && exit {
			if a.Finish == nil {
				MicroAppStop()
			} else {
				a.Finish("")
			}
			return
		}
	default:
	}
}

// ------------------------------------------------
// General App Methods
// ------------------------------------------------

// New Create new microapp element
func (a *MicroApp) New(name string) {
	a.defStyle = tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	a.frames = make(map[string]*Frame)
	a.styles = make(map[string]AppStyle)
	a.screen = screen
	a.name = name
	a.activeElement = ""
	a.activeFrame = ""
	a.WindowMouseEvent = nil
	a.WindowFinish = nil
	a.WindowKeyEvent = nil
	a.lockActive = false
	a.Finish = nil
	a.debug = false
}

// Start start the application
func (a *MicroApp) Start() {
	a.DrawAll()
	MouseOnOff(true)
	a.screen.HideCursor()
	a.screen.Show()
}

// Reset delete all elements
func (a *MicroApp) Reset() {
	for fname, f := range a.frames {
		for k := range f.elements {
			delete(f.elements, k)
		}
		delete(a.frames, fname)
	}
	for k := range a.styles {
		delete(a.styles, k)
	}
	a.WindowMouseEvent = nil
	a.WindowKeyEvent = nil
	a.WindowFinish = nil
	a.Finish = nil
	a.lockActive = false
	a.lastloc = Loc{-1, -1}
	a.activeElement = ""
	a.mouseOver = ""
	a.lastbutton = ""
	a.activeFrame = ""
	a.mousedown = false
}

// GetValues collect values of all elements in the forms
func (a *MicroApp) GetValues() map[string]string {
	var values = make(map[string]string)

	for _, f := range a.frames {
		for _, e := range f.elements {
			if e.form != "box" && e.form != "button" && e.form != "label" {
				if e.form == "checkbox" {
					if e.checked {
						if values[e.gname] == "" {
							values[e.gname] = e.value
						} else {
							values[e.gname] = values[e.gname] + "|" + e.value
						}
					}
				} else if e.form == "radio" {
					if e.checked {
						values[e.gname] = e.value
					}
				} else {
					values[e.name] = e.value
				}
			}
		}
	}
	return values
}

// GetValuesAsString needed for Lua, send the values as a single string
func (a *MicroApp) GetValuesAsString() string {
	var str string

	vals := a.GetValues()
	for k, v := range vals {
		str = str + k + "=" + v + "\n"
	}
	return str
}
