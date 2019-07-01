// microapp
package main

import (
	"strings"

	"github.com/hanspr/tcell"
)

const ENCODINGS = "UTF-8:|ISO-8859-1:|ISO-8859-2:|ISO-8859-15:|WINDOWS-1250:|WINDOWS-1251:|WINDOWS-1252:|WINDOWS-1256:|SHIFT-JIS:|GB2312:|EUC-KR:|EUC-JP:|GBK:|BIG-5:|ASCII:"

type microMenu struct {
	myapp     *MicroApp
	usePlugin bool
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
		m.myapp.AddWindowBox("b1", "Menu", 0, 1, 10, 4, nil, "")
		m.myapp.AddWindowLabel("m1", "micro-ide >", 4, 4, nil, "")
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
		width := 50
		heigth := 8
		left := w/2 - width/2
		top := h/2 - heigth/2
		right := width
		bottom := heigth
		m.myapp.AddStyle("f", "ColorWhite,ColorDarkBlue")
		m.myapp.AddWindowBox("enc", "Search", left, top, right, bottom, nil, "")
		m.myapp.AddWindowTextBox("search", "Search regex: ", "", "string", left+2, top+2, 27, 30, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowLabel("lblf", "Found:", left+2, top+4, nil, "")
		m.myapp.AddWindowLabel("found", "", left+9, top+4, nil, "f")
		m.myapp.AddWindowCheckBox("i", "i", "i", left+45, top+2, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowButton("cancel", " Cancel ", "cancel", left+23, top+6, m.ButtonFinish, "")
		m.myapp.AddWindowButton("set", " Search ", "ok", left+38, top+6, m.StartSearch, "")
		m.myapp.Finish = m.AbortSearch
		m.myapp.WindowFinish = callback
	}
	m.myapp.SetFocus("search", "E")
	m.myapp.Start()
	apprunning = m.myapp
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
		width := 50
		heigth := 12
		left := w/2 - width/2
		top := h/2 - heigth/2
		right := width
		bottom := heigth
		m.myapp.AddStyle("f", "ColorWhite,ColorDarkBlue")
		m.myapp.AddWindowBox("enc", "Search / Replace", left, top, right, bottom, nil, "")
		m.myapp.AddWindowTextBox("search", "Search regex:   ", "", "string", left+2, top+2, 30, 30, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowTextBox("replace", "Replace string: ", "", "string", left+2, top+4, 30, 30, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("i", "i", "i", left+2, top+6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("a", "all", "a", left+9, top+6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("s", `s (\n)`, "s", left+16, top+6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("l", "No regex", "l", left+26, top+6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowLabel("lblf", "Found:", left+2, top+8, nil, "")
		m.myapp.AddWindowLabel("found", "", left+9, top+8, nil, "f")
		m.myapp.AddWindowButton("cancel", " Cancel ", "cancel", left+23, top+10, m.ButtonFinish, "")
		m.myapp.AddWindowButton("set", " Search ", "ok", left+38, top+10, m.StartSearch, "")
		m.myapp.Finish = m.AbortSearch
		m.myapp.WindowFinish = callback
	}
	m.myapp.SetFocus("search", "E")
	m.myapp.Start()
	apprunning = m.myapp
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
	m.myapp.WindowFinish(m.myapp.getValues())
	m.Finish("Done")
	return true
}

func (m *microMenu) SubmitSearchOnEnter(name, value, event, when string, x, y int) bool {
	if event == "Enter" && when == "PRE" {
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
	if len(event) > 1 {
		if event == "Backspace2" || event == "Delete" {
			event = ""
		} else {
			return true
		}
	}
	found := DialogSearch(value)
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
		m.myapp.AddWindowBox("enc", "Save As ...", left, top, right, bottom, nil, "")
		m.myapp.AddWindowTextBox("filename", "File name : ", "", "string", left+2, top+2, 65, 200, nil, "")
		m.myapp.AddWindowSelect("encoding", "Encoding  : ", b.encoder, ENCODINGS+"|"+b.encoder+":"+b.encoder, left+2, top+4, 0, 1, m.SaveAsEncodingEvent, "")
		m.myapp.AddWindowTextBox("encode", "Use this encoding: ", "", "string", left+32, top+4, 15, 15, nil, "")
		m.myapp.AddWindowButton("cancel", " Cancel ", "cancel", left+56, top+6, m.SaveAsButtonFinish, "")
		m.myapp.AddWindowButton("set", "  Save  ", "ok", left+69, top+6, m.SaveFile, "")
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
		heigth := 6
		left := w/2 - width/2
		top := h/2 - heigth/2
		right := width
		bottom := heigth
		m.myapp.AddWindowBox("enc", "Select Encoding", left, top, right, bottom, nil, "")
		m.myapp.AddWindowSelect("encoding", "Encoding ", "UTF-8", ENCODINGS, left+2, top+2, 0, 1, nil, "")
		m.myapp.AddWindowTextBox("encode", "Use this encoding: ", "", "string", left+26, top+2, 14, 15, nil, "")
		m.myapp.AddWindowButton("cancel", "  Cancel  ", "cancel", left+26, top+4, m.ButtonFinish, "")
		m.myapp.AddWindowButton("set", "Set encoding", "ok", left+45, top+4, m.SetEncoding, "")
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

// Borrar al final de las pruebas

func (m *microMenu) Test() {
	//Create MicroApp objetc
	if m.myapp == nil || m.myapp.name != "test" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("test")
		} else {
			m.myapp.name = "test"
		}
		m.myapp.Reset()
		style := tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorWhite)
		m.myapp.defStyle = style
		m.myapp.AddStyle("def", "ColorWhite,ColorDarkBlue")
		m.myapp.AddStyle("bred", "bold #ff0000,ColorDarkBlue")
		m.myapp.AddStyle("bdef", "bold ColorWhite,ColorDarkBlue")
		m.myapp.AddStyle("bidef", "bold ColorDarkBlue,ColorWhite")
		m.myapp.AddWindowBox("b1", "Pruebas Generales", 2, 2, 80, 22, nil, "")
		m.myapp.AddWindowLabel("lbl1", "Presiona {bred}Esc{/bred} para salir áéíóú", 4, 4, nil, "")
		m.myapp.AddWindowRadio("gusta", "Sí gusta", "1", 4, 5, true, nil, "")
		m.myapp.AddWindowRadio("gusta", "{bred}No{/bred} gusta", "0", 15, 5, false, nil, "")
		m.myapp.AddWindowCheckBox("chk1", "Cine", "cine", 4, 6, false, nil, "")
		m.myapp.AddWindowCheckBox("chk2", "Leer", "leer", 14, 6, true, nil, "")
		m.myapp.AddWindowCheckBox("chk3", "Ca{bdef}mi{/bdef}nar", "caminar", 24, 6, false, nil, "bred")
		m.myapp.AddWindowTextBox("nombre", "{bdef}Nombre {/bdef}", "", "string", 4, 7, 5, 10, nil, "")
		m.myapp.AddWindowSelect("title", "Grupo    ", "1", "0:Pink Floyd|1:Rush|2:Genesis|3:Yes", 25, 8, 0, 1, nil, "")
		m.myapp.AddWindowSelect("prof", "Profesión ", "1", "0:Doctor|1:Engineer|2:Pilot|3:Pope|4:Philosopher|5:Cook|6:Writer|7:Programmer|8:Others", 4, 9, 0, 4, nil, "")
		m.myapp.AddWindowTextArea("area", "Área Texto", "", 4, 14, 30, 5, false, nil, "")
		m.myapp.AddWindowButton("btn1", "Save", "ok", 60, 15, m.GetDatos, "")
		m.myapp.AddWindowButton("btn2", "Cancel", "cancel", 60, 17, nil, "")
		m.myapp.AddWindowButton("btn3", "Exit", "", 60, 19, m.ButtonFinish, "")
		m.myapp.Finish = m.Finish
		//m.myapp.AddStyle("a1", "bold #ff0000,#ffffff")
		//m.myapp.AddStyle("a2", "underline #000000,#aaaaff")
	}
	m.myapp.Start()
	apprunning = m.myapp
	//m.myapp.PrintStyle("{a1}Color a1{/a1} {a2}Color a2{/a2}", 10, 25,nil)
}

func (m *microMenu) GetDatos(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	values := m.myapp.getValues()
	for k, v := range values {
		messenger.AddLog(k, "=", v)
	}
	messenger.Information("Revisa log")
	m.Finish("From save")
	return true
}

func (m *microMenu) CallBackMouseEvent(name, value, event, when string, x, y int) bool {
	return true
}

func (m *microMenu) CallBackKeyEvent(name, value, char, when string, x, y int) bool {
	return true
}
