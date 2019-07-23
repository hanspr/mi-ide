// app
package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hanspr/microidelibs/clipboard"
	"github.com/hanspr/tcell"
)

type AppElement struct {
	name       string // element name
	label      string // element label if it applyes
	form       string // types: box, textbox, textarea, label, checkbox, radio, button
	value      string
	value_type string                                              // string, number
	pos        Loc                                                 // Top left corner where to position element
	aposb      Loc                                                 // hotspot top,left
	apose      Loc                                                 // hotspot bottom,right
	cursor     Loc                                                 // cursor location inside text elements
	width      int                                                 // text element width, text area width
	height     int                                                 // Textbox: maxlength, text area have no predefined maxlength
	index      int                                                 // order in which to draw
	callback   func(string, string, string, string, int, int) bool // (element.name, element.value, event, x, y)
	style      tcell.Style                                         // color style for this element
	checked    bool                                                // Checkbox, Radio checked. Select is open
	gname      string                                              // Original name, required for radio buttons
	offset     int                                                 // Select = indexSelected, Text = offset from the age id box smaller than maxlength
	visible    bool                                                //Set element vibilility attribute
	iKey       int
	microapp   *MicroApp
}

type AppStyle struct {
	name  string
	style tcell.Style
}

type Canvas struct {
	top      int
	left     int
	right    int
	bottom   int
	position string
	otop     int
	oleft    int
	owidth   int
	oheight  int
}

type MicroApp struct {
	name             string
	screen           tcell.Screen
	defStyle         tcell.Style
	elements         map[string]AppElement
	styles           map[string]AppStyle
	Finish           func(string)
	WindowMouseEvent func(string, int, int) // event, x, y
	WindowKeyEvent   func(string, int, int) // key, x, y
	WindowFinish     func(map[string]string)
	cursor           Loc // Relative position inside canvas
	activeElement    string
	lockActive       bool
	mousedown        bool
	lastloc          Loc
	lastbutton       string
	lastclick        time.Time
	eint             int
	maxheigth        int
	canvas           Canvas
	maxindex         int
	mouseOver        string
}

// ------------------------------------------------
// App Build Methods
// ------------------------------------------------

func (a *MicroApp) SetCanvas(top, left, width, height int, position string) {
	w, h := a.screen.Size()
	a.canvas.otop = top
	a.canvas.oleft = left
	a.canvas.owidth = width
	a.canvas.oheight = height
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
	a.canvas.top = top
	a.canvas.left = left
	a.canvas.right = left + width
	a.canvas.bottom = top + height
	a.canvas.position = position
}

func (a *MicroApp) AddStyle(name, style string) {
	var s AppStyle

	s.name = name
	s.style = StringToStyle(style)
	a.styles[name] = s
}

func (a *MicroApp) AddWindowElement(name, label, form, value, value_type string, x, y, w, h int, chk bool, callback func(string, string, string, string, int, int) bool, style string) {
	var e AppElement

	e.name = name
	e.label = label
	e.form = form
	e.value = value
	e.value_type = value_type
	e.pos = Loc{x, y}
	e.cursor = Loc{0, 0}
	e.width = w
	e.height = h
	e.index = 1
	e.callback = callback
	e.checked = chk
	e.offset = 0
	e.microapp = a
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
		if y+h > a.maxheigth {
			a.maxheigth = y + h
		}
	} else if form == "radio" || form == "checkbox" {
		n := 0
		for _, e := range a.elements {
			if e.gname == name {
				n++
			}
		}
		e.gname = name
		e.name = name + strconv.Itoa(n)
	}
	// Calculate begin and end coordenates for the hotspot for element type
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
		e.index++
		h--
		e.offset = -1
		e.aposb = Loc{x + lblwidth, y}
		e.apose = Loc{x + lblwidth + w, y + h}
		opts := strings.Split(e.value_type, "|")
		e.cursor.X = len(opts)
		for i := 0; i < len(opts); i++ {
			opt := strings.Split(opts[i], ":")
			if opt[0] == e.value {
				e.offset = i
				break
			}
		}
	} else if form == "textarea" {
		e.aposb = Loc{x + 1, y + 1}
		e.apose = Loc{x + w - 1, y + h - 1}
	}
	a.elements[e.name] = e
}

func (a *MicroApp) AddWindowBox(name, title string, x, y, width, height int, border bool, callback func(string, string, string, string, int, int) bool, style string) {
	if width < 1 || height < 1 {
		return
	}
	a.AddWindowElement(name, title, "box", "", "", x, y, width, height, border, callback, style)
}

func (a *MicroApp) AddWindowLabel(name, label string, x, y int, callback func(string, string, string, string, int, int) bool, style string) {
	a.AddWindowElement(name, label, "label", "", "", x, y, 0, 0, false, callback, style)
}

func (a *MicroApp) AddWindowMenuLabel(name, label, kind string, x, y int, callback func(string, string, string, string, int, int) bool, style string) {
	if kind == "r" {
		a.AddWindowElement(name, " "+label+string(tcell.RuneVLine), "label", "", "", x, y, 0, 0, false, callback, style)
	} else if kind == "l" {
		a.AddWindowElement(name, string(tcell.RuneVLine)+label+" ", "label", "", "", x, y, 0, 0, false, callback, style)
	} else if kind == "cl" {
		a.AddWindowElement(name, string('┐')+label+string(tcell.RuneVLine), "label", "", "", x, y, 0, 0, false, callback, style)
	} else {
		a.AddWindowElement(name, string(tcell.RuneVLine)+label+string(tcell.RuneVLine), "label", "", "", x, y, 0, 0, false, callback, style)
	}
}

func (a *MicroApp) AddWindowMenuTop(name, label string, x, y int, callback func(string, string, string, string, int, int) bool, style string) {
	label = strings.ReplaceAll(label, " ", string(tcell.RuneHLine))
	a.AddWindowElement(name, string(tcell.RuneLLCorner)+label+string(tcell.RuneURCorner), "label", "", "", x, y, 0, 0, false, callback, style)
}

func (a *MicroApp) AddWindowMenuBottom(name, label string, x, y int, callback func(string, string, string, string, int, int) bool, style string) {
	label = strings.ReplaceAll(label, " ", string(tcell.RuneHLine))
	a.AddWindowElement(name, string(tcell.RuneLLCorner)+label+string(tcell.RuneLRCorner), "label", "", "", x, y, 0, 0, false, callback, style)
}

func (a *MicroApp) AddWindowTextBox(name, label, value, value_type string, x, y, width, maxlength int, callback func(string, string, string, string, int, int) bool, style string) {
	if width < 1 {
		return
	} else if width <= 3 && maxlength > 3 {
		maxlength = width
	}
	a.AddWindowElement(name, label, "textbox", value, value_type, x, y, width, maxlength, false, callback, style)
}

func (a *MicroApp) AddWindowCheckBox(name, label, value string, x, y int, checked bool, callback func(string, string, string, string, int, int) bool, style string) {
	a.AddWindowElement(name, label, "checkbox", value, "", x, y, 0, 0, checked, callback, style)
}

func (a *MicroApp) AddWindowRadio(name, label, value string, x, y int, checked bool, callback func(string, string, string, string, int, int) bool, style string) {
	a.AddWindowElement(name, label, "radio", value, "", x, y, 0, 0, checked, callback, style)
}

func (a *MicroApp) AddWindowTextArea(name, label, value string, x, y, columns, rows int, readonly bool, callback func(string, string, string, string, int, int) bool, style string) {
	if columns < 5 || rows < 2 {
		return
	}
	a.AddWindowElement(name, label, "textarea", value, "", x, y, columns+2, rows+2, readonly, callback, style)
}

func (a *MicroApp) AddWindowSelect(name, label, value string, options string, x, y, width, height int, callback func(string, string, string, string, int, int) bool, style string) bool {
	if options == "" {
		return false
	}
	if strings.Contains(options, "|") == false {
		return false
	}
	if width == 0 {
		opts := strings.Split(options, "|")
		for _, opt := range opts {
			if strings.Contains(opt, ":") == false {
				return false
			}
			v := strings.Split(opt, ":")
			if v[1] == "" {
				v[1] = v[0]
			}
			if Count(v[1]) > width {
				width = Count(v[1])
			}
		}
	}
	a.AddWindowElement(name, label, "select", value, options, x, y, width, height, false, callback, style)
	return true
}

func (a *MicroApp) AddWindowButton(name, label, button_type string, x, y int, callback func(string, string, string, string, int, int) bool, style string) {
	a.AddWindowElement(name, label, "button", "", button_type, x, y, 0, 0, false, callback, style)
}

// ------------------------------------------------
// Element Methods
// ------------------------------------------------

func (a *MicroApp) SetIndex(k string, v int) {
	if v > 5 {
		v = 5
	} else if v < 0 {
		v = 0
	}
	if v > a.maxindex {
		a.maxindex = v
	}
	e := a.elements[k]
	e.index = v
	a.elements[k] = e
}

func (a *MicroApp) GetVisible(k string) bool {
	return a.elements[k].visible
}

func (a *MicroApp) SetVisible(k string, v bool) {
	e := a.elements[k]
	if e.visible == v {
		return
	}
	e.visible = v
	a.elements[e.name] = e
	if v == true {
		e.Draw()
	} else {
		a.ResetCanvas()
		if a.activeElement == "" {
			a.cursor = Loc{0, 0}
			a.screen.HideCursor()
		} else {
			a.screen.ShowCursor(a.cursor.X+a.canvas.left, a.cursor.Y+a.canvas.top)
		}
	}
	a.screen.Show()
}

func (a *MicroApp) GetgName(k string) string {
	return a.elements[k].gname
}

func (a *MicroApp) SetgName(k, v string) {
	e := a.elements[k]
	e.gname = v
	a.elements[k] = e
}

func (a *MicroApp) GetiKey(k string) int {
	return a.elements[k].iKey
}

func (a *MicroApp) SetiKey(k string, v int) {
	e := a.elements[k]
	e.iKey = v
	a.elements[k] = e
}

func (a *MicroApp) GetValue(k string) string {
	return a.elements[k].value
}

func (a *MicroApp) SetValue(k, v string) {
	e := a.elements[k]
	e.value = v
	a.elements[k] = e
	e.Draw()
	a.screen.Show()
}

func (a *MicroApp) GetChecked(k string) bool {
	return a.elements[k].checked
}

func (a *MicroApp) SetCheked(k string, v bool) {
	e := a.elements[k]
	e.checked = v
	a.elements[e.name] = e
	e.Draw()
	a.screen.Show()
}

func (a *MicroApp) GetPos(k string) Loc {
	return a.elements[k].pos
}

func (a *MicroApp) SetPos(k string, v Loc) {
	e := a.elements[k]
	e.pos = v
	a.elements[e.name] = e
	e.Draw()
	a.screen.Show()
}

func (a *MicroApp) GetLabel(k string) string {
	return a.elements[k].label
}

func (a *MicroApp) SetLabel(k, v string) {
	e := a.elements[k]
	e.label = v
	a.elements[e.name] = e
	a.DrawAll()
	a.screen.Show()
}

func (a *MicroApp) SetFocus(k, where string) {
	if a.elements[k].index == 0 || (a.elements[k].form != "textbox" && a.elements[k].form != "textarea") {
		return
	}
	a.activeElement = k
	e := a.elements[k]
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

func (a *MicroApp) SetFocusPreviousInputElement(k string) {
	var next AppElement
	var last AppElement

	last.pos.Y = -1
	next.pos.Y = -1
	next.pos.X = -1
	me := a.elements[k]
	for _, e := range a.elements {
		if e.index == 0 || e.name == me.name || (e.form != "textbox" && e.form != "textarea") {
			continue
		}
		if e.pos.Y == me.pos.Y && e.pos.X < me.pos.X {
			next = e
			break
		}
		if e.pos.Y < me.pos.Y && e.pos.Y >= next.pos.Y && e.pos.X > next.pos.X {
			next = e
			continue
		}
		if e.pos.Y > last.pos.Y {
			last = e
			continue
		}
		if e.pos.Y == last.pos.Y && e.pos.X > last.pos.X {
			last = e
			continue
		}
	}
	if next.name != "" {
		a.SetFocus(next.name, "E")
	} else if last.name != "" {
		a.SetFocus(last.name, "E")
	}
}

func (a *MicroApp) SetFocusNextInputElement(k string) {
	var next AppElement
	var first AppElement

	first.pos.Y = 99999
	next.pos.Y = 99999
	next.pos.X = 99999
	me := a.elements[k]
	for _, e := range a.elements {
		if e.index == 0 || e.name == me.name || (e.form != "textbox" && e.form != "textarea") {
			continue
		}
		if e.pos.Y == me.pos.Y && e.pos.X > me.pos.X {
			next = e
			break
		}
		if e.pos.Y > me.pos.Y && e.pos.Y <= next.pos.Y && e.pos.X < next.pos.X {
			next = e
			continue
		}
		if e.pos.Y < first.pos.Y {
			first = e
			continue
		}
		if e.pos.Y == first.pos.Y && e.pos.X < first.pos.X {
			first = e
			continue
		}
	}
	if next.name != "" {
		a.SetFocus(next.name, "E")
	} else if first.name != "" {
		a.SetFocus(first.name, "E")
	}
}

func (a *MicroApp) DeleteElement(k string) {
	delete(a.elements, k)
}

// ------------------------------------------------
// Drawing Methods
// ------------------------------------------------

// Element Drawing

func (e *AppElement) Draw() {
	if e.visible == false {
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

func (e *AppElement) Hide() {
}

func (e *AppElement) DrawBox() {
	a := e.microapp
	Hborder := tcell.RuneHLine
	Vborder := tcell.RuneVLine
	ULC := tcell.RuneULCorner
	URC := tcell.RuneURCorner
	LLC := tcell.RuneLLCorner
	LRC := tcell.RuneLRCorner
	if e.checked == false {
		Hborder = ' '
		Vborder = ' '
		ULC = ' '
		URC = ' '
		LLC = ' '
		LRC = ' '
	}
	x1 := e.pos.X + a.canvas.left
	y1 := e.pos.Y + a.canvas.top
	x2 := e.pos.X + a.canvas.left + e.width
	y2 := e.pos.Y + a.canvas.top + e.height
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
		a.PrintStyle(e.label, e.pos.X+1, e.pos.Y, &e.style)
	}
	a.screen.Show()
}

func (e *AppElement) DrawLabel() {
	e.microapp.PrintStyle(e.label, e.pos.X, e.pos.Y, &e.style)
	e.microapp.screen.Show()
}

func (e *AppElement) DrawTextBox() {
	var r rune

	a := e.microapp
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
		a.elements[e.name] = *e
	}
	a.PrintStyle(e.label, e.pos.X, e.pos.Y, &e.style)
	style := e.style.Underline(true)
	for W := 0; W < e.width; W++ {
		if W+e.offset < len(val) {
			r = val[W+e.offset]
		} else {
			r = ' '
		}
		a.screen.SetContent(e.aposb.X+W+a.canvas.left, e.pos.Y+a.canvas.top, r, nil, style)
	}
	a.screen.Show()
}

func (e *AppElement) DrawRadio() {
	var radio string

	radio = "◎ "
	if e.checked == true {
		radio = "◉ "
	}
	e.microapp.PrintStyle(radio+e.label, e.pos.X, e.pos.Y, &e.style)
	e.microapp.screen.Show()
}

func (e *AppElement) DrawCheckBox() {
	check := "☐ "

	if e.checked == true {
		check = "☒ "
	}
	e.microapp.PrintStyle(check+e.label, e.pos.X, e.pos.Y, &e.style)
	e.microapp.screen.Show()
}

func (e *AppElement) DrawSelect() {
	var style tcell.Style

	a := e.microapp
	start := 0
	Y := -1
	f := "%-" + strconv.Itoa(e.width) + "s"
	a.PrintStyle(e.label, e.pos.X, e.pos.Y, &e.style)
	opts := strings.Split(e.value_type, "|")
	if e.height > 1 && e.height < len(opts) && e.offset >= e.height {
		// Overflow, find the starting point
		start = e.offset - e.height + 1
	}
	a.eint = start
	if a.activeElement == e.name {
		e.style = a.defStyle.Foreground(tcell.ColorBlack).Background(tcell.ColorKhaki)
	}
	for i := start; i < len(opts); i++ {
		chr := ""
		opt := strings.SplitN(opts[i], ":", 2)
		if opt[1] == "" {
			opt[1] = opt[0]
		}
		if i == e.offset {
			style = e.style.Reverse(true)
		} else if e.height == 1 {
			continue
		} else {
			style = e.style.Reverse(false)
		}
		if e.height == 1 {
			Y = 0
		} else {
			Y++
			if start > 0 && i == start {
				chr = "▲"
			} else if Y+1 == e.height && i+1 < len(opts) {
				chr = "▼"
			} else {
				chr = " "
			}
		}
		if Y >= e.height {
			break
		}
		label := []rune(fmt.Sprintf(f, opt[1]) + chr)
		for N := 0; N < len(label); N++ {
			a.screen.SetContent(e.aposb.X+N+a.canvas.left, e.aposb.Y+Y+a.canvas.top, label[N], nil, style)
		}
		if e.height == 1 {
			return
		}
	}
	a.screen.Show()
}

func (e *AppElement) DrawButton() {
	a := e.microapp
	style := e.style
	label := []rune(" " + e.label + " ")
	if e.value_type == "cancel" {
		style = e.style.Background(tcell.ColorDarkRed).Foreground(tcell.ColorWhite).Bold(true)
	} else if e.value_type == "ok" {
		style = e.style.Background(tcell.ColorDarkGreen).Foreground(tcell.ColorWhite).Bold(true)
	} else {
		style = e.style.Background(tcell.ColorDarkSlateGray).Foreground(tcell.ColorWhite).Bold(true)
	}
	for x := 0; x < len(label); x++ {
		e.microapp.screen.SetContent(e.pos.X+x+a.canvas.left, e.pos.Y+a.canvas.top, label[x], nil, style)
	}
}

func (e *AppElement) DrawTextArea() {
	a := e.microapp
	y := e.aposb.Y
	e.DrawBox()
	str := e.value
	str = WordWrap(str, e.width-1)
	lines := strings.Split(str, "\\N")
	for _, line := range lines {
		a.Print(line, e.aposb.X, y, nil)
		y++
		if y > e.apose.Y {
			break
		}
	}
	a.screen.Show()
}

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

func (e *AppElement) getECursorFromACursor() int {
	a := e.microapp
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
			a.screen.ShowCursor(a.cursor.X+a.canvas.left, a.cursor.Y+a.canvas.top)
			return newx
		}
		ac = ac + X
	}
	a.cursor.Y = e.aposb.Y + Y
	a.cursor.X = e.aposb.X + X
	a.screen.ShowCursor(a.cursor.X+a.canvas.left, a.cursor.Y+a.canvas.top)
	return ac
}

func (e *AppElement) setACursorFromECursor() {
	a := e.microapp
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
	a.screen.ShowCursor(a.cursor.X+a.canvas.left, a.cursor.Y+a.canvas.top)
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

func (a *MicroApp) DrawAll() {
	// Draw all elements in index order from 0 to current max index (normally 2)
	for i := 0; i <= a.maxindex; i++ {
		for _, e := range a.elements {
			if e.index == i {
				e.Draw()
			}
		}
	}
	if a.activeElement == "" {
		a.screen.HideCursor()
	}
	a.screen.Show()
}

func (a *MicroApp) Resize() {
	for _, t := range tabs {
		t.Resize()
	}
	if a.canvas.position == "relative" {
		w, h := a.screen.Size()
		if a.canvas.otop+a.canvas.oleft+a.canvas.owidth+a.canvas.oheight == 0 {
			a.canvas.right = w
			a.canvas.bottom = h
		} else {
			if a.canvas.oleft < 0 {
				a.canvas.left = w/2 - a.canvas.owidth/2
			}
			if a.canvas.otop < 0 {
				a.canvas.top = h/2 - a.canvas.oheight/2
			}
			a.canvas.right = a.canvas.left + a.canvas.owidth
			a.canvas.bottom = a.canvas.top + a.canvas.oheight
		}
	}
	RedrawAll(false)
	a.DrawAll()
	if a.activeElement == "" {
		a.cursor = Loc{0, 0}
		a.screen.HideCursor()
	} else {
		a.screen.ShowCursor(a.cursor.X+a.canvas.left, a.cursor.Y+a.canvas.top)
	}
	a.screen.Show()
}

// Clear Routines

// Clear the entire screen from any app drawings
func (a *MicroApp) ClearScreen(style *tcell.Style) {
	RedrawAll(false)
	a.screen.Show()
}

// Redraw Canvas
func (a *MicroApp) ResetCanvas() {
	RedrawAll(false)
	a.DrawAll()
}

// Print Routines

// Debug works with absolute coordenates
func (a *MicroApp) Debug(msg string, x, y int) {
	a.PrintAbsolute(msg, x, y, nil)
	a.screen.Show()
}

// Print works with relative coordinates
func (a *MicroApp) Print(msg string, x, y int, style *tcell.Style) {
	x += a.canvas.left
	y += a.canvas.top
	a.PrintAbsolute(msg, x, y, style)
}

// PrintStyle works with relative coordinates
func (a *MicroApp) PrintStyle(msg string, x, y int, estyle *tcell.Style) {
	var style *tcell.Style
	var defstyle *tcell.Style

	if estyle == nil {
		estyle = &a.defStyle
	} else {
		defstyle = estyle
	}
	re := regexp.MustCompile(`{(\w+)}`)
	parts := regexp.MustCompile(`{/\w+}`).Split(msg, -1)
	for _, part := range parts {
		if part == "" {
			continue
		}
		name := re.FindStringSubmatch(part)
		if len(name) == 0 {
			style = estyle
		} else if val, ok := a.styles[name[1]]; ok {
			style = &val.style
		} else {
			style = estyle
		}
		txt := re.Split(part, -1)
		if txt[0] != "" {
			a.Print(txt[0], x, y, defstyle)
			x = x + Count(txt[0])
		}
		if len(txt) > 1 {
			a.Print(txt[1], x, y, style)
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

func (e *AppElement) TextAreaClickEvent(event string, x, y int) {
	a := e.microapp
	a.activeElement = e.name
	a.cursor.X = x
	a.cursor.Y = y
	e.cursor.X = e.getECursorFromACursor()
	a.elements[e.name] = *e
	a.screen.Show()
}

func (e *AppElement) TextBoxClickEvent(event string, x, y int) {
	a := e.microapp
	a.activeElement = e.name
	w := Count(e.value)
	if x > e.aposb.X+w {
		x = e.aposb.X + w
		a.cursor.X = x
	}
	e.cursor.X = e.offset + a.cursor.X - e.aposb.X
	a.screen.ShowCursor(x+a.canvas.left, y+a.canvas.top)
	a.elements[e.name] = *e
	a.screen.Show()
}

func (e *AppElement) SelectClickEvent(event string, x, y int) {
	a := e.microapp

	a.activeElement = e.name
	// SELECT BEGIN
	if e.height == 1 && e.checked == false {
		// Set height, and new hotspot, save
		opts := strings.Split(e.value_type, "|")
		e.height = len(opts)
		e.apose = Loc{e.apose.X, e.apose.Y + e.height - 1}
		e.checked = true
		// Lock events to this element until closed
		a.lockActive = true
		a.activeElement = e.name
		a.elements[e.name] = *e
		a.DrawAll()
		return
	} else if e.checked == true {
		opts := strings.Split(e.value_type, "|")
		for i := 0; i < e.cursor.X; i++ {
			if e.aposb.Y+i == y {
				if i < e.cursor.X {
					opt := strings.SplitN(opts[i], ":", 2)
					e.value = opt[0]
					e.offset = i
				}
				break
			}
		}

		// Reset to height=1, hotspot, savew
		if e.aposb.Y+e.height > a.maxheigth {
			a.screen.Clear()
			RedrawAll(false)
		}
		e.height = 1
		e.apose = Loc{e.apose.X, e.aposb.Y}
		// release event locking
		a.lockActive = false
		a.activeElement = ""
		e.checked = false
		a.elements[e.name] = *e
		// Needs to be all to erase list
		a.DrawAll()
		return
	}
	// SELECT END

	// LIST
	opts := strings.Split(e.value_type, "|")
	offset := a.eint
	//a.Debug(fmt.Sprintf("CLICKB S:%d", offset), 100, 4)
	for i := 0; i < e.cursor.X; i++ {
		if e.aposb.Y+i == y {
			if i+offset < e.cursor.X {
				opt := strings.SplitN(opts[i+offset], ":", 2)
				e.value = opt[0]
				e.offset = i + offset
				//} else {
				//a.Debug(fmt.Sprintf("ERROR! i+offset<len(opts): %d+%d < %d ecX %d", i, offset, len(opts), e.cursor.X), 100, 7)
			}
			break
		}
	}
	//a.Debug(fmt.Sprintf("CLICKA selectIndex:%d value:%s", e.offset, e.value), 100, 5)
	a.elements[e.name] = *e
	e.Draw()
}

func (e *AppElement) RadioCheckboxClickEvent(event string, x, y int) {
	a := e.microapp
	if e.form == "radio" && e.checked == true {
		return
	}
	if e.checked == true {
		e.checked = false
	} else {
		e.checked = true
	}
	if e.form == "radio" {
		e.DrawRadio()
		for _, r := range a.elements {
			if r.form == "radio" && r.gname == e.gname && r.name != e.name && e.checked == true {
				r.checked = false
				a.elements[r.name] = r
				r.DrawRadio()
			}
		}
	} else {
		e.DrawCheckBox()
	}
	a.elements[e.name] = *e
	a.screen.Show()
}

func (e *AppElement) ProcessElementClick(event string, x, y int) {
	a := e.microapp
	if a.lockActive == true {
		if a.activeElement != e.name {
			return
		}
		e.SelectClickEvent(event, x, y)
		return
	}
	name := e.name
	if a.activeElement != "" {
		a.activeElement = ""
		a.screen.HideCursor()
		a.DrawAll()
	}
	check := false
	if e.form == "checkbox" || e.form == "radio" {
		name = e.gname
		check = true
	}
	if e.callback != nil {
		if e.callback(name, e.value, event, "PRE", x, y) == false {
			return
		}
	}
	if event == "click1" {
		if check == true {
			e.RadioCheckboxClickEvent(event, x, y)
		} else if e.form == "textbox" {
			e.TextBoxClickEvent(event, x, y)
		} else if e.form == "textarea" {
			e.TextAreaClickEvent(event, x, y)
		} else if e.form == "select" {
			e.SelectClickEvent(event, x, y)
		}
	}
	if e.callback != nil {
		if check == true {
			value := "false"
			if e.checked == true {
				value = "true"
			}
			e.callback(name, value, event, "POST", x, y)
		} else {
			e.callback(name, e.value, event, "POST", x, y)
		}
	}
}

func (e *AppElement) ProcessElementMouseMove(event string, x, y int) {
	if e.callback != nil {
		if e.callback(e.name, e.value, event, "", x, y) == false {
			return
		}
	}
	// TODO highlitght elements on mouse(over/out)
}

func (e *AppElement) ProcessElementMouseDown(event string, x, y int) {
}

func (e *AppElement) ProcessElementMouseWheel(event string, x, y int) {
}

// ------------------------------------------------
// Key Events
// ------------------------------------------------

func (e *AppElement) SelectKeyEvent(key string, x, y int) {
	a := e.microapp
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
	} else if key == "Home" {
		e.offset = 0
	} else if key == "End" {
		e.offset = e.cursor.X - 1
	} else if key == "Enter" {
		a.activeElement = ""
		if a.lockActive == true || e.checked == true {
			e.height = 1
			e.apose = Loc{e.apose.X, e.aposb.Y}
			a.lockActive = false
			e.checked = false
		}
	}
	opts := strings.Split(e.value_type, "|")
	opt := strings.SplitN(opts[e.offset], ":", 2)
	e.value = opt[0]
	a.elements[e.name] = *e
	//a.Debug(fmt.Sprintf("KEY: v=%s , offset=%d", e.value, e.offset), 100, 4)
	a.DrawAll()
	a.screen.Show()
}

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
		} else if key == "Delete" {
			b = a.removeCharAt(b, e.cursor.X)
			e.value = string(b)
		} else if key == "Left" {
			if e.cursor.X-1 < 0 {
				return
			}
			e.cursor.X--
			e.setACursorFromECursor()
		} else if key == "Right" {
			if e.cursor.X+1 > len(b) || (x+1 > e.apose.X && y == e.apose.Y) {
				return
			}
			e.cursor.X++
			e.setACursorFromECursor()
		} else if key == "Up" {
			if a.cursor.Y-1 < e.aposb.Y {
				return
			}
			a.cursor.Y--
			e.cursor.X = e.getECursorFromACursor()
			a.elements[e.name] = *e
			a.screen.Show()
			return
		} else if key == "Down" {
			if a.cursor.Y+1 > e.apose.Y {
				return
			}
			a.cursor.Y++
			e.cursor.X = e.getECursorFromACursor()
			a.elements[e.name] = *e
			a.screen.Show()
			return
		} else if key == "Home" {
			a.cursor.X = e.aposb.X
			e.cursor.X = 0
			a.cursor.Y = e.aposb.Y
		} else if key == "End" {
			e.cursor.X = len(b)
		} else if key == "Enter" {
			return
		} else if key == "Ctrl+V" {
			clip, _ := clipboard.ReadAll("clipboard")
			e.value = e.value + clip
			a.elements[e.name] = *e
			e.TextAreaKeyEvent("End", x, y)
			return
		}
		a.elements[e.name] = *e
		e.Draw()
		e.setACursorFromECursor()
		a.screen.Show()
		return
	}
	e.value = string(a.insertCharAt(b, r, e.cursor.X))
	e.cursor.X++
	e.DrawTextArea()
	e.setACursorFromECursor()
	a.elements[e.name] = *e
	a.screen.Show()
	return
}

func (e *AppElement) TextBoxKeyEvent(key string, x, y int) {
	a := e.microapp
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
		} else if key == "Delete" {
			b = a.removeCharAt(b, e.cursor.X)
			e.value = string(b)
		} else if key == "Left" {
			if e.cursor.X-1 < 0 {
				return
			}
			if a.cursor.X-1 >= e.aposb.X {
				a.cursor.X--
			}
			e.cursor.X--
		} else if key == "Right" {
			if e.cursor.X < maxlength-1 && e.cursor.X < len(b) {
				e.cursor.X++
				if a.cursor.X+1 <= e.apose.X {
					a.cursor.X++
				}
			}
		} else if key == "Home" {
			a.cursor.X = e.aposb.X
			e.cursor.X = 0
			e.offset = 0
		} else if key == "End" {
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
			a.SetFocusNextInputElement(e.name)
			return
		} else if key == "Backtab" {
			a.SetFocusPreviousInputElement(e.name)
			return
		} else if key == "Ctrl+V" {
			clip, _ := clipboard.ReadAll("clipboard")
			e.value = e.value + clip
			a.elements[e.name] = *e
			e.TextBoxKeyEvent("End", x, y)
			return
		}
		a.elements[e.name] = *e
		e.DrawTextBox()
		a.screen.ShowCursor(a.cursor.X+a.canvas.left, a.cursor.Y+a.canvas.top)
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
	a.elements[e.name] = *e
	e.DrawTextBox()
	a.screen.ShowCursor(a.cursor.X+a.canvas.left, a.cursor.Y+a.canvas.top)
	a.screen.Show()
}

func (e *AppElement) ProcessElementKey(key string, x, y int) {
	a := e.microapp
	if a.activeElement == "" || a.activeElement != e.name {
		// This should not happen, for safety
		a.screen.HideCursor()
		return
	}
	if e.callback != nil {
		if e.callback(e.name, e.value, key, "PRE", x, y) == false {
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
	}
}

// Check if event occures on top of an element hotspot
// If it does, dispacth the appropiate method

func (a *MicroApp) CheckElementsActions(event string, x, y int) bool {
	//a.Debug(fmt.Sprintf("CheckElementActions %s", time.Now()), 90, 2)
	if x < 0 || y < 0 {
		return false
	}
	for _, e := range a.elements {
		if e.index == 0 {
			// Skip boxes
			continue
		}
		// Check if location is inside the element hotspot
		if x >= e.aposb.X && x <= e.apose.X && y >= e.aposb.Y && y <= e.apose.Y {
			//a.Debug(fmt.Sprintf("Hotspot ok %s , %s", e.name, event), 90, 3)
			if strings.Contains(event, "mouse") {
				if a.mouseOver != "" && a.mouseOver != e.name {
					ex := a.elements[a.mouseOver]
					ex.ProcessElementMouseMove("mouseout", x, y)
					e.ProcessElementMouseMove("mousein", x, y)
				} else if a.mouseOver == "" {
					e.ProcessElementMouseMove("mousein", x, y)
				} else {
					e.ProcessElementMouseMove(event, x, y)
				}
				a.mouseOver = e.name
			} else if strings.Contains(event, "click") {
				a.cursor = Loc{x, y}
				e.ProcessElementClick(event, x, y)
			} else if strings.Contains(event, "button") {
				e.ProcessElementMouseDown(event, x, y)
			} else if strings.Contains(event, "wheel") {
				e.ProcessElementMouseWheel(event, x, y)
			} else {
				e.ProcessElementKey(event, x, y)
				a.mouseOver = ""
			}
			return true
		}
	}
	if a.mouseOver != "" {
		ex := a.elements[a.mouseOver]
		ex.ProcessElementMouseMove("mouseout", x, y)
		a.mouseOver = ""
	}
	//a.Debug(fmt.Sprintf("CheckElementActions END %s", time.Now()), 90, 4)
	if strings.Contains(event, "click") && a.lockActive == false {
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

func (a *MicroApp) HandleEvents(event tcell.Event) {
	char := ""
	//a.Debug(fmt.Sprintf("LOOP %s", time.Now()), 90, 10)
	switch ev := event.(type) {
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
		if ev.Key() == 256 {
			char = string(ev.Rune())
		} else {
			char = ev.Name()
		}
		if a.WindowKeyEvent != nil {
			a.WindowKeyEvent(char, a.cursor.X, a.cursor.Y)
		} else {
			a.CheckElementsActions(char, a.cursor.X, a.cursor.Y)
		}
	case *tcell.EventMouse:
		//a.Debug(fmt.Sprintf("LOOP %s", time.Now()), 90, 10)
		exit := false
		xa, ya := ev.Position()
		x := xa - a.canvas.left
		y := ya - a.canvas.top
		button := ev.Buttons()
		action := ""
		if button == tcell.ButtonNone {
			//a.Debug(fmt.Sprintf("BUTTON None %s", time.Now()), 90, 11)
			if a.mousedown == true {
				if xa < a.canvas.left || xa > a.canvas.right || ya < a.canvas.top || ya > a.canvas.bottom {
					exit = true
				}
				//a.Debug(fmt.Sprintf("MouseUp? %s", time.Now()), 90, 12)
				Dt := time.Since(a.lastclick) / time.Millisecond
				// If button released within 2 pixel consider as clic and not a dragstop
				if Abs(a.lastloc.X-x) < 3 && Abs(a.lastloc.Y-y) < 3 {
					if Dt < 300 {
						//a.Debug(fmt.Sprintf("doubleclick? (%d) %s", Dt, time.Now()), 90, 13)
						action = "doubleclick" + a.lastbutton
					} else {
						//a.Debug(fmt.Sprintf("click? %s", time.Now()), 90, 14)
						action = "click" + a.lastbutton
						a.mousedown = false
					}
				} else {
					//a.Debug(fmt.Sprintf("mouse-drag? %s", time.Now()), 90, 15)
					action = "mousemove-dragstop" + a.lastbutton
				}
				//a.Debug(fmt.Sprintf("Limpia %s", time.Now()), 90, 16)
				a.lastclick = time.Now()
				a.lastbutton = ""
			} else {
				action = "mousemove"
			}
		} else if ev.HasMotion() == true {
			action = "mousemove-drag" + a.lastbutton
		} else if ev.HasMotion() == false {
			//a.Debug(fmt.Sprintf("MouseDown? %s", time.Now()), 90, 30)
			a.mousedown = true
			if button == tcell.Button1 {
				action = "button1"
				a.lastbutton = "1"
			} else if button == tcell.Button2 {
				action = "button2"
				a.lastbutton = "2"
			} else if button == tcell.Button3 {
				action = "button3"
				a.lastbutton = "3"
			} else if button&tcell.WheelUp != 0 {
				action = "wheelUp"
				a.mousedown = false
			} else if button&tcell.WheelDown != 0 {
				action = "wheelDown"
				a.mousedown = false
			} else if button&tcell.WheelLeft != 0 {
				action = "wheelLeft"
				a.mousedown = false
			} else if button&tcell.WheelRight != 0 {
				action = "wheelRight"
				a.mousedown = false
			}
			a.lastloc = Loc{x, y}
			//a.Debug(fmt.Sprintf("Action %s : %s", action, time.Now()), 90, 31)
		} else {
			action = "mousemove"
		}
		if a.WindowMouseEvent != nil {
			//a.Debug(fmt.Sprintf("MouseEvent??? %s", time.Now()), 90, 32)
			a.WindowMouseEvent(action, x, y)
		}
		//a.Debug(fmt.Sprintf("CheckActions %s", time.Now()), 90, 33)
		if a.CheckElementsActions(action, x, y) == false && exit {
			if a.Finish == nil {
				MicroAppStop()
			} else {
				a.Finish("")
			}
			return
		}
	default:
		//a.Debug(fmt.Sprintf("DEFAULT? %s", time.Now()), 90, 40)
	}
}

// ------------------------------------------------
// General App Methods
// ------------------------------------------------

func (a *MicroApp) New(name string) {
	a.defStyle = tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	a.elements = make(map[string]AppElement)
	a.styles = make(map[string]AppStyle)
	a.screen = screen
	a.name = name
	a.WindowMouseEvent = nil
	a.WindowFinish = nil
	a.WindowKeyEvent = nil
	a.lockActive = false
	a.maxindex = 2
	a.Finish = nil
}

func (a *MicroApp) Start() {
	a.DrawAll()
	a.screen.HideCursor()
	a.screen.Show()
}

func (a *MicroApp) Reset() {
	for k, _ := range a.elements {
		delete(a.elements, k)
	}
	for k, _ := range a.styles {
		delete(a.styles, k)
	}
	a.WindowMouseEvent = nil
	a.WindowKeyEvent = nil
	a.WindowFinish = nil
	a.Finish = nil
	a.lockActive = false
	a.activeElement = ""
	a.mousedown = false
	a.lastloc = Loc{0, 0}
	a.lastbutton = ""
	a.eint = 0
}

func (a *MicroApp) getValues() map[string]string {
	var values = make(map[string]string)

	for _, e := range a.elements {
		if e.form != "box" && e.form != "button" && e.form != "label" {
			if e.form == "checkbox" {
				if e.checked == true {
					if values[e.gname] == "" {
						values[e.gname] = e.value
					} else {
						values[e.gname] = values[e.gname] + "|" + e.value
					}
				}
			} else if e.form == "radio" {
				if e.checked == true {
					values[e.gname] = e.value
				}
			} else {
				values[e.name] = e.value
			}
		}
	}
	return values
}
