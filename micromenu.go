// microapp
package main

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/hanspr/tcell"
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

// ---------------------------------------
// Micro-ide Menu
// ---------------------------------------

func (m *microMenu) Menu() {
	if m.myapp == nil || m.myapp.name != "mi-menu" {
		m.activemenu = ""
		m.submenu = make(map[string]string)
		m.submenuElements = make(map[string][]menuElements)
		m.submenuWidth = make(map[string]int)
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-menu")
		} else {
			m.myapp.name = "mi-menu"
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
		m.AddSubMenuItem("microide", Language.Translate("Global Settings"), m.GlobalConfigDialog)
		m.AddSubMenuItem("microide", Language.Translate("KeyBindings"), m.KeyBindingsDialog)
		m.AddSubMenuItem("microide", Language.Translate("Plugin Manager"), m.PluginManagerDialog)
		row++
		for _, name := range keys {
			if name != "microide" {
				label := fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", m.submenu[name])
				m.myapp.AddWindowMenuLabel(name, label, "", 0, row, m.ShowSubmenuItems, "")
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
		e := m.myapp.elements[name]
		e.style = e.style.Bold(false).Foreground(tcell.ColorWhite)
		m.myapp.elements[name] = e
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
	e := m.myapp.elements[name]
	e.style = e.style.Bold(true).Foreground(tcell.ColorYellow)
	m.myapp.elements[name] = e
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
	if event == "mouseout" {
		e := m.myapp.elements[name]
		e.style = e.style.Bold(false).Foreground(tcell.ColorWhite)
		m.myapp.elements[name] = e
		e.Draw()
		return false
	} else if event == "mousein" {
		e := m.myapp.elements[name]
		e.style = e.style.Bold(true).Foreground(tcell.ColorYellow)
		m.myapp.elements[name] = e
		e.Draw()
		return false
	}
	if event != "mouse-click1" {
		return false
	}
	m.myapp.ClearScreen()
	ex := m.submenuElements[m.activemenu][m.myapp.GetiKey(name)]
	ex.callback()
	return false
}

func (m *microMenu) MenuFinish(s string) {
	m.closeSubmenus()
	m.Finish("")
}

// ---------------------------------------
// Edit Global Settings
// ---------------------------------------

func (m *microMenu) GlobalConfigDialog() {
	if m.myapp == nil || m.myapp.name != "mi-globalconfig" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-globalconfig")
		} else {
			m.myapp.name = "mi-globalconfig"
		}
		width := 100
		height := 25
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		m.myapp.SetCanvas(-1, -1, width, height, "relative")
		m.myapp.AddWindowBox("enc", Language.Translate("Global Settings"), 0, 0, width, height, true, nil, "")
		keys := make([]string, 0, len(globalSettings))
		for k := range globalSettings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		row := 2
		col := 2
		for _, k := range keys {
			if k == "fileformat" {
				m.myapp.AddWindowSelect(k, k+" ", globalSettings[k].(string), "unix:|dos:", col, row, 0, 1, nil, "")
			} else if k == "colorcolumn" {
				m.myapp.AddWindowTextBox(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "string", col, row, 4, 3, m.ValidateInteger, "")
			} else if k == "indentchar" {
				char := "s"
				if globalSettings[k].(string) != " " {
					char = "t"
				}
				m.myapp.AddWindowSelect(k, k+" ", char, "t:Tab|s:Space", col, row, 0, 1, nil, "")
			} else if k == "scrollmargin" {
				m.myapp.AddWindowSelect(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "0:|1:|2:|3:|4:|5:|6:|7:|8:|9:|10:", col, row, 3, 1, nil, "")
				m.myapp.SetIndex(k, 3)
			} else if k == "tabsize" {
				m.myapp.AddWindowSelect(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "1:|2:|3:|4:|5:|6:|7:|8:|9:|10:", col, row, 3, 1, nil, "")
			} else if k == "lang" {
				Langs := ""
				langs := GeTFileListFromPath(configDir+"/langs", "lang")
				for _, l := range langs {
					if Langs == "" {
						Langs = Langs + l + ":"
					} else {
						Langs = Langs + "|" + l + ":"
					}
				}
				m.myapp.AddWindowSelect(k, k+" ", globalSettings[k].(string), Langs, col, row, 0, 1, nil, "")
			} else if k == "colorscheme" {
				Colors := ""
				colors := GeTFileListFromPath(configDir+"/colorschemes", "micro")
				for _, c := range colors {
					if Colors == "" {
						Colors = Colors + c + ":"
					} else {
						Colors = Colors + "|" + c + ":"
					}
				}
				m.myapp.AddWindowSelect(k, k+" ", globalSettings[k].(string), Colors, col, row, 0, 1, nil, "")
				m.myapp.SetIndex(k, 3)
			} else {
				kind := reflect.TypeOf(globalSettings[k]).Kind()
				if kind == reflect.Bool {
					m.myapp.AddWindowCheckBox(k, k, "true", col, row, globalSettings[k].(bool), nil, "")
				} else if kind == reflect.String {
					m.myapp.AddWindowTextBox(k, k+" ", globalSettings[k].(string), "string", col, row, 10, 20, nil, "")
				} else if kind == reflect.Float64 {
					m.myapp.AddWindowTextBox(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "integer", col, row, 5, 10, nil, "")
				} else {
					continue
				}
			}
			row += 2
			if row > height-2 {
				row = 2
				col += 26
			}
		}
		lbl := Language.Translate("Cancel")
		m.myapp.AddWindowButton("cancel", " "+lbl+" ", "cancel", width-Count(lbl)-20, height-1, m.ButtonFinish, "")
		lbl = Language.Translate("Save")
		m.myapp.AddWindowButton("save", " "+lbl+" ", "ok", width-Count(lbl)-5, height-1, m.SaveSettings, "")
	}
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) SaveSettings(name, value, event, when string, x, y int) bool {
	if event != "mouse-click1" {
		return true
	}
	if when == "POST" {
		return true
	}
	m.Finish("")
	save := false
	v := ""
	values := m.myapp.getValues()
	if values["indentchar"] == "t" {
		values["indentchar"] = "\t"
	} else {
		values["indentchar"] = " "
	}
	for k, _ := range globalSettings {
		kind := reflect.TypeOf(globalSettings[k]).Kind()
		if kind == reflect.Bool {
			v = strconv.FormatBool(globalSettings[k].(bool))
			if values[k] == "" {
				values[k] = "false"
			}
		} else if kind == reflect.String {
			v = globalSettings[k].(string)
		} else if kind == reflect.Float64 {
			v = fmt.Sprintf("%g", globalSettings[k].(float64))
		} else {
			continue
		}
		messenger.AddLog(k, ">>", v, "?", values[k])
		if v != values[k] {
			messenger.AddLog("  Save setting")
			save = true
			SetOption(k, values[k])
			if k == "lang" {
				Language.ChangeLangName(values[k], configDir+"/langs/"+values[k]+".lang")
			}
		}
	}
	if save {
		messenger.AddLog("SAVE")
		err := WriteSettings(configDir + "/settings.json")
		messenger.AddLog("Error? ", err)
		if err != nil {
			messenger.Error(Language.Translate("Error writing settings.json file"), ": ", err.Error())
		}
	}
	return true
}

func (m *microMenu) ValidateInteger(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	if strings.Contains(event, "mouse") {
		return true
	}
	if Count(event) > 1 {
		return true
	}
	if strings.Contains("0123456789", event) {
		return true
	}
	return false
}

// ---------------------------------------
// Assign Keybindings to actions
// ---------------------------------------

func (m *microMenu) KeyBindingsDialog() {
	if m.myapp == nil || m.myapp.name != "mi-keybindings" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-keybindings")
		} else {
			m.myapp.name = "mi-keybindings"
		}
		width := 0 // 125
		height := 33
		c := 0
		defaults := DefaultBindings()
		mkeys := make(map[string]int)
		keys := make([]string, 0, len(bindingActions))
		for k := range bindingActions {
			if strings.Contains(k, "Mouse") || strings.Contains(k, "Scroll") || strings.Contains(k, "Outdent") || strings.Contains(k, "Indent") || strings.Contains(k, "Backs") || strings.Contains(k, "Delete") || strings.Contains(k, "Insert") || strings.Index(k, "End") == 0 || strings.Contains(k, "Escape") {
				continue
			}
			switch k {
			case "CursorDown", "CursorUp", "CursorLeft", "CursorRight", "FindNext", "FindPrevious":
				continue
			}
			keys = append(keys, k)
			mkeys[k] = 1
		}
		bindings := make(map[string]string)
		for k, v := range bindingsStr {
			if strings.Contains(k, "Mouse") || strings.Contains(v, ",") {
				continue
			}
			switch k {
			case "Left", "Up", "Down", "Right", "Backspace", "Delete", "CtrlH", "Esc", "Enter":
				continue
			}
			_, ok := bindings[v]
			if ok {
				if defaults[k] == v {
					bindings[v+"!"+strconv.Itoa(c)] = bindings[v]
					bindings[v] = k
				} else {
					bindings[v+"!"+strconv.Itoa(c)] = k
				}
				keys = append(keys, v+"!"+strconv.Itoa(c))
				c++
			} else {
				bindings[v] = k
				_, ok := mkeys[v]
				if ok == false {
					keys = append(keys, v)
				}
			}
		}
		width = ((len(keys)+2)/(height-2)+1)*42 + 1
		sort.Strings(keys)
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		m.myapp.AddStyle("grn", "#00FF00,#262626")
		m.myapp.AddStyle("yw", "#ffff00,#262626")
		m.myapp.AddStyle("red", "#ff0000,#262626")
		m.myapp.SetCanvas(-1, -1, width, height, "relative")
		m.myapp.AddWindowBox("enc", Language.Translate("KeyBindings"), 0, 0, width, height, true, nil, "")
		row := 2
		col := 2
		for _, k := range keys {
			_, ok := bindings[k]
			if ok == false {
				bindings[k] = ""
			}
			str := strings.Split(k, "!")
			if len(str) == 1 {
				m.myapp.AddWindowTextBox(k, fmt.Sprintf("%-23s", str[0]), bindings[k], "string", col, row, 17, 40, m.SetBinding, "")
			} else {
				m.myapp.AddWindowTextBox(k, fmt.Sprintf("{grn}%-23s{/grn}", str[0]), bindings[k], "string", col, row, 17, 40, m.SetBinding, "")
			}
			m.myapp.SetgName(k, m.myapp.elements[k].value)
			row++
			if row > height-2 {
				row = 2
				col += 42
			}
		}
		row++
		m.myapp.AddWindowTextBox("?test", fmt.Sprintf("{yw}%-23s{/yw}", Language.Translate("Key check")), "", "string", col, row, 17, 40, m.SetBinding, "")
		row++
		m.myapp.AddWindowLabel("?msg", "", col, row, nil, "")
		lbl := Language.Translate("Cancel")
		m.myapp.AddWindowButton("?cancel", " "+lbl+" ", "cancel", width-Count(lbl)-20, 32, m.ButtonFinish, "")
		lbl = Language.Translate("Save")
		m.myapp.AddWindowButton("?save", " "+lbl+" ", "ok", width-Count(lbl)-5, 32, m.SaveKeyBindings, "")
	}
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) SetBinding(name, value, event, when string, x, y int) bool {
	if event == "mouse-click1" {
		if name == "?test" {
			m.myapp.SetValue(name, "")
		}
		m.myapp.SetFocus(name, "B")
		return true
	}
	if strings.Contains(event, "mouse") || strings.Contains(event, "Alt+Ctrl") || when == "POST" || Count(event) < 2 {
		return false
	}
	switch event {
	case "Left", "Right", "Down", "Up", "Esc", "Enter", "Tab", "Backspace2", "Backspace", "Delete", "F9", "F10", "F11", "F12", "PgDn", "PgUp":
		m.myapp.SetValue(name, "")
		m.myapp.SetFocus(name, "B")
		return false
	}
	if strings.Contains(event, "Alt") {
		if strings.Contains(event, "Rune") {
			event = strings.Replace(event, "Rune[", "", 1)
			event = strings.Replace(event, "]", "", 1)
			if strings.Contains(event, "++") {
				event = strings.Replace(event, "++", "-+", 1)
			} else {
				event = strings.ReplaceAll(event, "+", "-")
			}
		} else {
			event = strings.ReplaceAll(event, "+", "")
		}
		event = strings.Replace(event, "ShiftAlt", "AltShift", 1)
	} else if strings.Contains(event, "Ctrl") {
		event = strings.ReplaceAll(event, "+", "")
		if strings.Contains(event, "ShiftCtrl") {
			event = strings.Replace(event, "ShiftCtrl", "CtrlShift", 1)
		}
	}
	free := true
	for _, e := range m.myapp.elements {
		if e.value == event {
			free = false
			if e.name != "?test" {
				m.myapp.SetLabel("?msg", "{red}"+Language.Translate("Taken")+" : {/red}"+e.label)
			}
			if name != "?test" {
				m.myapp.SetValue(e.name, "")
			}
			break
		}
	}
	m.myapp.SetValue(name, event)
	if free {
		m.myapp.SetLabel("?msg", "{grn}"+Language.Translate("Free")+"{/grn}")
	}
	return false
}

func (m *microMenu) SaveKeyBindings(name, value, event, when string, x, y int) bool {
	if event != "mouse-click1" {
		return true
	}
	if when == "POST" {
		return true
	}
	write := false
	save := make(map[string]string)
	actionToKey := make(map[string]string)
	values := m.myapp.getValues()
	defaults := DefaultBindings()
	for k, a := range defaults {
		if strings.Contains(a, "Mouse") || strings.Contains(k, "Mouse") {
			continue
		}
		actionToKey[a] = k
	}
	// Compare each binding vs defaults
	for a, k := range values {
		// Check actions
		if strings.Contains(a, "?") || strings.Contains(a, ".") {
			// Skip buttons, test, plugins
			continue
		}
		ak := strings.Split(a, "!")
		//messenger.AddLog("Compare:", ak[0], " > ", actionToKey[ak[0]], "==", k)
		if k != "" && actionToKey[ak[0]] != k {
			//messenger.AddLog("Guardar=", k, ",", ak[0])
			save[k] = ak[0]
			write = true
			BindKey(k, a)
		}
		e := m.myapp.elements[a]
		if e.gname != "" && e.gname != k {
			//messenger.AddLog("??:", ak[0], " > ", actionToKey[ak[0]], "==", k, " :: ", e.gname, ":: ", defaults[e.gname])
			if defaults[e.gname] == "" {
				//messenger.AddLog("UnbindKey=", e.gname)
				BindKey(e.gname, "UnbindKey")
			} else if defaults[e.gname] != ak[0] {
				//messenger.AddLog("RebindToDefault=", e.gname, ":", defaults[e.gname])
				BindKey(e.gname, defaults[e.gname])
			}
		}
	}
	if write {
		WriteBindings(save)
	}
	m.Finish("")
	// Forse to reload always
	m.myapp = nil
	return false
}

// ---------------------------------------
// Plugin Manager Dialog
// ---------------------------------------

func (m *microMenu) PluginManagerDialog() {
	m.Finish("")
}

// ---------------------------------------
// Application Search & Replace
// ---------------------------------------

// Search (Find) Dialog

func (m *microMenu) Search(callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "mi-search" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-search")
		} else {
			m.myapp.name = "mi-search"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 70
		heigth := 8
		m.myapp.SetCanvas(-1, -1, width, heigth, "relative")
		m.myapp.AddStyle("f", "black,yellow")
		m.myapp.AddWindowBox("enc", Language.Translate("Search"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Search regex:")
		m.myapp.AddWindowTextBox("find", lbl+" ", "", "string", 2, 2, 61-Count(lbl), 50, m.SubmitSearchOnEnter, "")
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
			m.myapp.SetValue("find", CurView().Cursor.GetSelection())
		}
	}
	m.myapp.SetFocus("find", "E")
	apprunning = m.myapp
	m.SubmitSearchOnEnter("find", m.myapp.GetValue("find"), "", "POST", 0, 0)
}

// Search & Replace Dialog

func (m *microMenu) SearchReplace(callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "mi-searchreplace" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-searchreplace")
		} else {
			m.myapp.name = "mi-searchreplace"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
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
	if strings.Contains(event, "mouse") && event != "mouse-click1" {
		return true
	}
	if event == "Enter" && when == "PRE" && (name == "find" || name == "replace") {
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
	}
	if len([]rune(event)) > 1 {
		if event == "Backspace2" || event == "Delete" || event == "Ctrl+V" || event == "Paste" {
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

// ---------------------------------------
// Application Save As ...
// ---------------------------------------

func (m *microMenu) SaveAs(b *Buffer, usePlugin bool, callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "mi-saveas" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-saveas")
		} else {
			m.myapp.name = "mi-saveas"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
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

// ---------------------------------------
// Set Buffer Encoding
// ---------------------------------------

func (m *microMenu) SelEncoding(encoder string, callback func(map[string]string)) {
	if m.myapp == nil || m.myapp.name != "mi-selencoding" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-selencoding")
		} else {
			m.myapp.name = "mi-selencoding"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 60
		heigth := 8
		m.myapp.SetCanvas(-1, -1, width, heigth, "relative")
		m.myapp.AddWindowBox("enc", Language.Translate("Select Encoding"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Encoding:")
		m.myapp.AddWindowSelect("encoding", lbl+" ", encoder, ENCODINGS, 2, 2, 0, 1, nil, "")
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
	if event != "mouse-click1" {
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

// ---------------------------------------
// General Routines
// ---------------------------------------

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
