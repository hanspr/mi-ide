// microapp
package main

import (
	"strings"

	"github.com/hanspr/tcell"
)

const ENCODINGS = "UTF-8:|ISO-8859-1:|ISO-8859-2:|ISO-8859-15:|WINDOWS-1250:|WINDOWS-1251:|WINDOWS-1252:|WINDOWS-1256:|SHIFT-JIS:|GB2312:|EUC-KR:|EUC-JP:|GBK:|BIG-5:|ASCII:"

type microMenu struct {
	myapp       *MicroApp
	usePlugin   bool
	searchMatch bool
}

// Applicacion Micro Menu

func (m *microMenu) Menu() {
	if m.myapp == nil || m.myapp.name != "micromenu" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("micromenu")
		} else {
			m.myapp.name = "micromenu"
		}
		m.myapp.Reset()
		style := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
		m.myapp.defStyle = style
		m.myapp.AddWindowBox("b1", "", 0, 1, 20, 3, nil, "")
		m.myapp.AddWindowLabel("m1", "micro-ide >", 1, 2, nil, "")
	}
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) Finish(s string) {
	messenger.AddLog(s)
	apprunning = nil
	FixTabsIconArea()
}

func (m *microMenu) ButtonFinish(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	m.Finish("Abort")
	return true
}

// END Menu

// Application Search / Replace

func (m *microMenu) Search(callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "search" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("search")
		} else {
			m.myapp.name = "search"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		m.myapp.AddStyle("def", "ColorWhite,ColorDarkGrey")
		w, h := screen.Size()
		width := 70
		heigth := 8
		left := w/2 - width/2
		top := h/2 - heigth/2
		right := width
		bottom := heigth
		m.myapp.AddStyle("f", "black,yellow")
		m.myapp.AddWindowBox("enc", Language.Translate("Search"), left, top, right, bottom, nil, "")
		lbl := Language.Translate("Search regex:")
		m.myapp.AddWindowTextBox("search", lbl+" ", "", "string", left+2, top+2, 61-Count(lbl), 50, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("i", "i", "i", left+65, top+2, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowLabel("found", "", left+2, top+4, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", left+46-Count(lbl), top+6, m.ButtonFinish, "")
		lbl = Language.Translate("Search")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", left+64-Count(lbl), top+6, m.StartSearch, "")
		m.myapp.Finish = m.AbortSearch
		m.myapp.WindowFinish = callback
	}
	m.searchMatch = false
	m.myapp.Start()
	if CurView().Cursor.HasSelection() {
		sel := CurView().Cursor.CurSelection
		messenger.AddLog(sel)
		if sel[0].Y == sel[1].Y {
			m.myapp.SetValue("search", CurView().Cursor.GetSelection())
		}
	}
	m.myapp.SetFocus("search", "E")
	apprunning = m.myapp
	m.SubmitSearchOnEnter("search", m.myapp.GetValue("search"), "", "POST", 0, 0)
}

func (m *microMenu) SearchReplace(callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "searchreplace" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("searchreplace")
		} else {
			m.myapp.name = "searchreplace"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		m.myapp.AddStyle("def", "ColorWhite,ColorDarkGrey")
		w, h := screen.Size()
		width := 70
		heigth := 12
		left := w/2 - width/2
		top := h/2 - heigth/2
		right := width
		bottom := heigth
		m.myapp.AddStyle("f", "bold black,yellow")
		m.myapp.AddWindowBox("enc", Language.Translate("Search / Replace"), left, top, right, bottom, nil, "")
		lbl := Language.Translate("Search regex:")
		m.myapp.AddWindowTextBox("search", lbl+" ", "", "string", left+2, top+2, 63-Count(lbl), 50, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("Replace regex:")
		m.myapp.AddWindowTextBox("replace", lbl+" ", "", "string", left+2, top+4, 54-Count(lbl), 40, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("ignore case")
		m.myapp.AddWindowCheckBox("i", lbl, "i", left+2, top+6, false, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("all")
		m.myapp.AddWindowCheckBox("a", lbl, "a", left+17, top+6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("s", `s (\n)`, "s", left+32, top+6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("l", "No regex", "l", left+47, top+6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowLabel("found", "", left+2, top+8, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", left+46-Count(lbl), top+10, m.ButtonFinish, "")
		lbl = Language.Translate("Search")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", left+64-Count(lbl), top+10, m.StartSearch, "")
		m.myapp.Finish = m.AbortSearch
		m.myapp.WindowFinish = callback
	}
	m.searchMatch = false
	m.myapp.Start()
	if CurView().Cursor.HasSelection() {
		sel := CurView().Cursor.CurSelection
		messenger.AddLog(sel)
		if sel[0].Y == sel[1].Y {
			m.myapp.SetValue("search", CurView().Cursor.GetSelection())
		}
	}
	m.myapp.SetFocus("search", "E")
	apprunning = m.myapp
	m.SubmitSearchOnEnter("search", m.myapp.GetValue("search"), "", "POST", 0, 0)
}

func (m *microMenu) AbortSearch(s string) {
	var resp = make(map[string]string)
	resp["search"] = ""
	m.myapp.SetValue("search", "")
	m.myapp.SetLabel("found", "")
	m.myapp.WindowFinish(resp)
	m.Finish("Search Abort")
}

func (m *microMenu) StartSearch(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	if m.searchMatch == false {
		m.myapp.SetValue("search", "")
	}
	m.myapp.WindowFinish(m.myapp.getValues())
	m.Finish("Done")
	return true
}

func (m *microMenu) SubmitSearchOnEnter(name, value, event, when string, x, y int) bool {
	if event == "Enter" && when == "PRE" {
		if m.searchMatch == false {
			m.myapp.SetValue("search", "")
		}
		m.myapp.WindowFinish(m.myapp.getValues())
		m.Finish("Done")
		return false
	}
	if name == "replace" {
		return true
	}
	if when == "PRE" {
		return true
	}
	if event == "click1" {
		value2 := m.myapp.GetValue("search")
		if name == "i" && value == "true" {
			value2 = "(?i)" + value2
		}
		if name == "s" && value == "true" {
			value2 = "(?s)" + value2
		}
		event = ""
		value = value2
	} else if strings.Contains(event, "click") {
		return true
	}
	if len([]rune(event)) > 1 {
		if event == "Backspace2" || event == "Delete" || event == "Ctrl+V" {
			event = ""
		} else {
			return true
		}
	}
	found := DialogSearch(value)
	if len(found) > 68 {
		found = found[:67]
	}
	if len(found) > 0 {
		m.searchMatch = true
	} else {
		m.searchMatch = false
	}
	m.myapp.SetLabel("found", found)
	return true
}

// END Application Search / Replace

// Application Save As ...

func (m *microMenu) SaveAs(b *Buffer, usePlugin bool, callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "saveas" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("saveas")
		} else {
			m.myapp.name = "saveas"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		m.myapp.AddStyle("def", "ColorWhite,ColorDarkGrey")
		w, h := screen.Size()
		width := 80
		//heigth := 22
		heigth := 8
		left := w/2 - width/2
		top := h/2 - heigth/2
		right := width
		bottom := heigth
		m.myapp.AddWindowBox("enc", Language.Translate("Save As ..."), left, top, right, bottom, nil, "")
		lbl := Language.Translate("File name :")
		m.myapp.AddWindowTextBox("filename", lbl+" ", "", "string", left+2, top+2, 76-Count(lbl), 200, nil, "")
		lbl = Language.Translate("Encoding:")
		m.myapp.AddWindowSelect("encoding", lbl+" ", b.encoder, ENCODINGS+"|"+b.encoder+":"+b.encoder, left+2, top+4, 0, 1, m.SaveAsEncodingEvent, "")
		lbl = Language.Translate("Use this encoding:")
		m.myapp.AddWindowTextBox("encode", lbl+" ", "", "string", left+55-Count(lbl), top+4, 15, 15, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", left+56-Count(lbl), top+6, m.SaveAsButtonFinish, "")
		lbl = Language.Translate("Save")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", left+75-Count(lbl), top+6, m.SaveFile, "")
		m.myapp.WindowFinish = callback
	}
	m.usePlugin = usePlugin
	m.myapp.SetValue("filename", b.Path)
	m.myapp.SetFocus("filename", "B")
	if strings.Contains(ENCODINGS, b.encoder) {
		m.myapp.SetValue("encoding", b.encoder)
		m.myapp.SetValue("encode", "")
	} else {
		m.myapp.SetValue("encoding", "UTF-8")
		m.myapp.SetValue("encode", b.encoder)
	}
	m.myapp.Finish = m.SaveAsFinish
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) SaveAsEncodingEvent(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	if event == "click1" {
		m.myapp.SetValue("encode", "")
	}
	return true
}

func (m *microMenu) SaveAsFinish(x string) {
	var resp = make(map[string]string)
	resp["filename"] = ""
	resp["encoding"] = ""
	m.myapp.WindowFinish(resp)
	m.Finish("Abort")
}

func (m *microMenu) SaveAsButtonFinish(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	m.SaveAsFinish("")
	return true
}

func (m *microMenu) SaveFile(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	var resp = make(map[string]string)
	values := m.myapp.getValues()
	if values["encode"] != "" {
		resp["encoding"] = strings.ToUpper(values["encode"])
	} else {
		resp["encoding"] = values["encoding"]
	}
	resp["filename"] = values["filename"]
	if m.usePlugin {
		resp["usePlugin"] = "1"
	} else {
		resp["usePlugin"] = "0"
	}
	m.myapp.WindowFinish(resp)
	m.Finish("Save As")
	return true
}

// END Save As ...

// Application to Set a Buffer Encoding

func (m *microMenu) SelEncoding(callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "microselencoding" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("microselencoding")
		} else {
			m.myapp.name = "microselencoding"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		m.myapp.AddStyle("def", "ColorWhite,ColorDarkGrey")
		w, h := screen.Size()
		width := 60
		heigth := 8
		left := w/2 - width/2
		top := h/2 - heigth/2
		right := width
		bottom := heigth
		m.myapp.AddWindowBox("enc", Language.Translate("Select Encoding"), left, top, right, bottom, nil, "")
		lbl := Language.Translate("Encoding:")
		m.myapp.AddWindowSelect("encoding", lbl+" ", "UTF-8", ENCODINGS, left+2, top+2, 0, 1, nil, "")
		lbl = Language.Translate("Use this encoding:")
		m.myapp.AddWindowTextBox("encode", lbl+" ", "", "string", left+2, top+4, 15, 15, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", left+33-Count(lbl), top+6, m.ButtonFinish, "")
		lbl = Language.Translate("Set encoding")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", left+43, top+6, m.SetEncoding, "")
		m.myapp.WindowFinish = callback
	}
	m.myapp.SetValue("encode", "")
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) SetEncoding(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	var resp = make(map[string]string)
	values := m.myapp.getValues()
	if values["encode"] != "" {
		resp["encoding"] = strings.ToUpper(values["encode"])
	} else {
		resp["encoding"] = values["encoding"]
	}
	m.myapp.WindowFinish(resp)
	m.Finish("Encoding")
	return true
}

// END Encoding
