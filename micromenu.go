// microapp
package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const ENCODINGS = "UTF-8:|ISO-8859-1:|ISO-8859-2:|ISO-8859-15:|WINDOWS-1250:|WINDOWS-1251:|WINDOWS-1252:|WINDOWS-1256:|SHIFT-JIS:|GB2312:|EUC-KR:|EUC-JP:|GBK:|BIG-5:|ASCII:"

type menuElements struct {
	label    string
	callback func()
}

type microMenu struct {
	myapp           *MicroApp
	usePlugin       bool
	searchMatch     bool
	submenu         map[string]string
	submenuElements map[string][]menuElements
	submenuWidth    map[string]int
	maxwidth        int
	activemenu      string
}

// Applicacion Micro Menu

func (m *microMenu) Menu() {
	if m.myapp == nil || m.myapp.name != "micromenu" {
		m.submenu = make(map[string]string)
		m.submenuElements = make(map[string][]menuElements)
		m.submenuWidth = make(map[string]int)
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("micromenu")
		} else {
			m.myapp.name = "micromenu"
		}
		m.myapp.Reset()
		style := StringToStyle("#ffffff,#222222")
		m.myapp.defStyle = style
		keys := make([]string, 0, len(m.submenu))
		m.maxwidth = 0
		for k := range m.submenu {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		name := "microide"
		row := 0
		m.AddSubmenu(name, "Micro-ide")
		m.myapp.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", "Micro-ide"), "", 0, row, m.ShowSubmenuItems, "")
		m.AddSubMenuItem("microide", Language.Translate("Global Configurations"), m.GlobalConfigDialog)
		m.AddSubMenuItem("microide", Language.Translate("KeyBindings"), m.KeyBindingsDialog)
		m.AddSubMenuItem("microide", Language.Translate("Plugin Manager"), m.PluginManagerDialog)
		row++
		for _, name := range keys {
			if name != "microide" {
				label := fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", m.submenu[name])
				m.myapp.AddWindowLabel(name, label, 0, row, m.ShowSubmenuItems, "")
			}
			row++
		}
		m.myapp.AddWindowMenuBottom("menubottom", fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", " "), 0, row, nil, "")
		row++
		m.myapp.SetCanvas(1, 0, m.maxwidth, row-1, "fixed")
		m.myapp.Finish = m.MenuFinish
	} else {
		m.closeSubmenus()
	}
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) AddSubmenu(name, label string) {
	m.submenu[name] = label
	if Count(label) > m.maxwidth {
		m.maxwidth = Count(label)
	}
}

func (m *microMenu) AddSubMenuItem(submenu, label string, callback func()) {
	es := m.submenuElements[submenu]
	e := new(menuElements)
	e.label = label
	e.callback = callback
	es = append(es, *e)
	m.submenuElements[submenu] = es
	if Count(label) > m.submenuWidth[submenu] {
		m.submenuWidth[submenu] = Count(label)
	}
}

func (m *microMenu) ShowSubmenuItems(name, value, event, when string, x, y int) bool {
	if event == "mouseout" {
		if y == m.myapp.elements[name].aposb.Y {
			return true
		}
		m.closeSubmenus()
		m.activemenu = ""
		return true
	}
	if m.activemenu != "" && m.activemenu == name {
		return true
	} else if m.activemenu != "" {
		m.closeSubmenus()
	}
	m.activemenu = name
	width := m.submenuWidth[name]
	if y > 1 {
		m.myapp.AddWindowMenuTop("smenubottom", fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", " "), m.maxwidth+2, y, nil, "")
	}
	// Show new submenu
	items := m.submenuElements[name]
	for i, s := range items {
		name := "submenu" + strconv.Itoa(i)
		if i == 0 {
			m.myapp.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "r", m.maxwidth+2, y, m.MenuItemClick, "")
		} else if i == 1 {
			m.myapp.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "cl", m.maxwidth+2, y, m.MenuItemClick, "")
		} else {
			m.myapp.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "", m.maxwidth+2, y, m.MenuItemClick, "")
		}
		m.myapp.SetIndex(name, 3)
		m.myapp.SetgName(name, "submenu")
		m.myapp.SetiKey(name, i)
		y++
	}
	m.myapp.AddWindowMenuBottom("smenubottom", fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", " "), m.maxwidth+2, y, nil, "")
	m.myapp.SetgName("smenubottom", "submenu")
	m.myapp.DrawAll()
	return true
}

func (m *microMenu) closeSubmenus() {
	if m.activemenu == "" {
		return
	}
	// TODO Remove all submenu elements from last active menu
	for k, e := range m.myapp.elements {
		if e.gname == "submenu" {
			delete(m.myapp.elements, k)
		}
	}
	m.myapp.ResetCanvas()
	m.myapp.DrawAll()
	m.activemenu = ""
}

func (m *microMenu) MenuItemClick(name, value, event, when string, x, y int) bool {
	if event != "mouse-click1" {
		return false
	}
	ex := m.submenuElements[m.activemenu][m.myapp.GetiKey(name)]
	ex.callback()
	return false
}

func (m *microMenu) GlobalConfigDialog() {
	m.Finish("")
}

func (m *microMenu) KeyBindingsDialog() {
	m.Finish("")
}

func (m *microMenu) PluginManagerDialog() {
	m.Finish("")
}

func (m *microMenu) MenuFinish(s string) {
	m.closeSubmenus()
	m.Finish("")
}

// END Menu

// Application Search & Replace

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
		width := 70
		heigth := 8
		m.myapp.SetCanvas(-1, -1, width, heigth, "relative")
		m.myapp.AddStyle("f", "black,yellow")
		m.myapp.AddWindowBox("enc", Language.Translate("Search"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Search regex:")
		m.myapp.AddWindowTextBox("search", lbl+" ", "", "string", 2, 2, 61-Count(lbl), 50, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowCheckBox("i", "i", "i", 65, 2, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowLabel("found", "", 2, 4, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", 46-Count(lbl), 6, m.ButtonFinish, "")
		lbl = Language.Translate("Search")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", 64-Count(lbl), 6, m.StartSearch, "")
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
		width := 70
		heigth := 12
		m.myapp.SetCanvas(-1, -1, width, heigth, "relative")
		m.myapp.AddStyle("f", "bold black,yellow")
		m.myapp.AddWindowBox("enc", Language.Translate("Search / Replace"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Search regex:")
		m.myapp.AddWindowTextBox("search", lbl+" ", "", "string", 2, 2, 63-Count(lbl), 50, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("Replace regex:")
		m.myapp.AddWindowTextBox("replace", lbl+" ", "", "string", 2, 4, 54-Count(lbl), 40, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("ignore case")
		offset := Count(lbl) + 6
		m.myapp.AddWindowCheckBox("i", lbl, "i", 2, 6, false, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("all")
		m.myapp.AddWindowCheckBox("a", lbl, "a", offset, 6, false, m.SubmitSearchOnEnter, "")
		offset += Count(lbl) + 4
		m.myapp.AddWindowCheckBox("s", `s (\n)`, "s", offset, 6, false, m.SubmitSearchOnEnter, "")
		offset += Count(`s (\n)`) + 4
		lbl = Language.Translate("No regexp")
		m.myapp.AddWindowCheckBox("l", lbl, "l", offset, 6, false, m.SubmitSearchOnEnter, "")
		m.myapp.AddWindowLabel("found", "", 2, 8, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", 46-Count(lbl), 10, m.ButtonFinish, "")
		lbl = Language.Translate("Search")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", 64-Count(lbl), 10, m.StartSearch, "")
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
	if event != "mouse-click1" {
		return true
	}
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
	if strings.Contains(event, "mouse") {
		return true
	}
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
	if event == "mouse-click1" {
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
		width := 80
		heigth := 8
		m.myapp.SetCanvas(-1, -1, width, heigth, "relative")
		m.myapp.AddWindowBox("enc", Language.Translate("Save As ..."), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("File name :")
		m.myapp.AddWindowTextBox("filename", lbl+" ", "", "string", 2, 2, 76-Count(lbl), 200, m.SaveFile, "")
		lbl = Language.Translate("Encoding:")
		m.myapp.AddWindowSelect("encoding", lbl+" ", b.encoder, ENCODINGS+"|"+b.encoder+":"+b.encoder, 2, 4, 0, 1, m.SaveAsEncodingEvent, "")
		lbl = Language.Translate("Use this encoding:")
		m.myapp.AddWindowTextBox("encode", lbl+" ", "", "string", 55-Count(lbl), 4, 15, 15, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", 56-Count(lbl), 6, m.SaveAsButtonFinish, "")
		lbl = Language.Translate("Save")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", 75-Count(lbl), 6, m.SaveFile, "")
		m.myapp.WindowFinish = callback
	}
	m.usePlugin = usePlugin
	m.myapp.SetValue("filename", b.Path)
	if strings.Contains(ENCODINGS, b.encoder) {
		m.myapp.SetValue("encoding", b.encoder)
		m.myapp.SetValue("encode", "")
	} else {
		m.myapp.SetValue("encoding", "UTF-8")
		m.myapp.SetValue("encode", b.encoder)
	}
	m.myapp.Finish = m.SaveAsFinish
	m.myapp.Start()
	m.myapp.SetFocus("filename", "E")
	apprunning = m.myapp
}

func (m *microMenu) SaveAsEncodingEvent(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	if event == "mouse-click1" {
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
	if event != "mouse-click1" {
		return true
	}
	m.SaveAsFinish("")
	return true
}

func (m *microMenu) SaveFile(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	if (name == "filename" && event != "Enter") || (name == "set" && event != "mouse-click1") {
		return true
	}
	//	if name == "filename" && event == "Enter" {
	//	} else if name="ok" && event == "mouse-click1" {
	//	} else {
	//		return true
	//	}
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
		width := 60
		heigth := 8
		m.myapp.SetCanvas(-1, -1, width, heigth, "relative")
		m.myapp.AddWindowBox("enc", Language.Translate("Select Encoding"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Encoding:")
		m.myapp.AddWindowSelect("encoding", lbl+" ", "UTF-8", ENCODINGS, 2, 2, 0, 1, nil, "")
		lbl = Language.Translate("Use this encoding:")
		m.myapp.AddWindowTextBox("encode", lbl+" ", "", "string", 2, 4, 15, 15, nil, "")
		lbl = Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", 33-Count(lbl), 6, m.ButtonFinish, "")
		lbl = Language.Translate("Set encoding")
		m.myapp.AddWindowButton("set", " "+lbl+" ", "ok", 43, 6, m.SetEncoding, "")
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

// General Default Routines

func (m *microMenu) Finish(s string) {
	messenger.AddLog(s)
	apprunning = nil
	MicroToolBar.FixTabsIconArea()
}

func (m *microMenu) ButtonFinish(name, value, event, when string, x, y int) bool {
	if event != "mouse-click1" {
		return true
	}
	if when == "POST" {
		return true
	}
	m.Finish("Abort")
	return true
}
