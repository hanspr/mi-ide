// microapp
package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/hanspr/highlight"
	"github.com/hanspr/ioencoder"
	"github.com/hanspr/tcell"
)

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
	LastPath        string
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
		f := m.myapp.AddFrame("menu", 1, 0, 1, 1, "fixed")
		m.AddSubmenu(name, "Micro-ide")
		f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", "Micro-ide"), "", 0, row, m.ShowSubmenuItems, "", "")
		m.AddSubMenuItem("microide", Language.Translate("Global Settings"), m.GlobalConfigDialog)
		m.AddSubMenuItem("microide", Language.Translate("KeyBindings"), m.KeyBindingsDialog)
		m.AddSubMenuItem("microide", Language.Translate("Plugin Manager"), m.PluginManagerDialog)
		m.AddSubMenuItem("microide", Language.Translate("Cloud Services"), m.MiCloudServices)
		row++
		for _, name := range keys {
			if name != "microide" {
				label := fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", m.submenu[name])
				f.AddWindowMenuLabel(name, label, "", 0, row, m.ShowSubmenuItems, "", "")
			}
			row++
		}
		f.AddWindowMenuBottom("menubottom", fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", " "), 0, row, nil, "", "")
		row++
		f.ChangeFrame(1, 0, m.maxwidth, row-1, "fixed")
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
		if y == m.myapp.frames["menu"].elements[name].aposb.Y {
			return true
		}
		e := m.myapp.frames["menu"].elements[name]
		e.style = e.style.Bold(false).Foreground(tcell.ColorWhite)
		m.myapp.frames["menu"].elements[name] = e
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
	e := m.myapp.frames["menu"].elements[name]
	e.style = e.style.Bold(true).Foreground(tcell.ColorYellow)
	m.myapp.frames["menu"].elements[name] = e
	width := m.submenuWidth[name]
	f := m.myapp.frames["menu"]
	if y > 1 {
		f.AddWindowMenuTop("smenubottom", fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", " "), m.maxwidth+2, y, nil, "", "")
	}
	// Show new submenu
	items := m.submenuElements[name]
	for i, s := range items {
		name := "submenu" + strconv.Itoa(i)
		if i == 0 {
			f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "r", m.maxwidth+2, y, m.MenuItemClick, "", "")
		} else if i == 1 {
			f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "cl", m.maxwidth+2, y, m.MenuItemClick, "", "")
		} else {
			f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "", m.maxwidth+2, y, m.MenuItemClick, "", "")
		}
		f.SetIndex(name, 3)
		f.SetgName(name, "submenu")
		f.SetiKey(name, i)
		y++
	}
	f.AddWindowMenuBottom("smenubottom", fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", " "), m.maxwidth+2, y, nil, "", "")
	f.SetgName("smenubottom", "submenu")
	m.myapp.DrawAll()
	// Release memory from unused menus?
	//for k, e := range f.elements {
	//if e.gname == "submenu" {
	//if f.elements[k].visible == false {
	//delete(f.elements, k)
	//}
	//}
	//}
	return true
}

func (m *microMenu) closeSubmenus() {
	if m.activemenu == "" {
		return
	}
	// Remove all submenu elements from last active menu
	f := m.myapp.frames["menu"]
	for k, e := range f.elements {
		if e.gname == "submenu" {
			f.elements[k].visible = false
		}
	}
	m.myapp.ResetFrames()
	m.myapp.DrawAll()
	m.activemenu = ""
}

func (m *microMenu) MenuItemClick(name, value, event, when string, x, y int) bool {
	f := m.myapp.frames["menu"]
	_, err := f.elements[name]
	if err == false {
		return false
	}
	if event == "mouseout" {
		e := f.elements[name]
		e.style = e.style.Bold(false).Foreground(tcell.ColorWhite)
		f.elements[name] = e
		e.Draw()
		return false
	} else if event == "mousein" {
		e := f.elements[name]
		e.style = e.style.Bold(true).Foreground(tcell.ColorYellow)
		f.elements[name] = e
		e.Draw()
		return false
	}
	if event != "mouse-click1" {
		return false
	}
	m.myapp.ClearScreen()
	ex := m.submenuElements[m.activemenu][f.GetiKey(name)]
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
		f := m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		f.AddWindowBox("enc", Language.Translate("Global Settings"), 0, 0, width, height, true, nil, "", "")
		keys := make([]string, 0, len(globalSettings))
		for k := range globalSettings {
			if strings.Contains(k, "mi-") {
				continue
			}
			keys = append(keys, k)
		}
		sort.Strings(keys)
		row := 2
		col := 2
		for _, k := range keys {
			if k == "fileformat" {
				f.AddWindowSelect(k, k+" ", globalSettings[k].(string), "unix|dos", col, row, 0, 1, nil, "", "")
			} else if k == "colorcolumn" {
				f.AddWindowTextBox(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "string", col, row, 4, 3, m.ValidateInteger, "", "")
			} else if k == "indentchar" {
				char := "s"
				if globalSettings[k].(string) != " " {
					char = "t"
				}
				f.AddWindowSelect(k, k+" ", char, "t]Tab|s]Space", col, row, 0, 1, nil, "", "")
			} else if k == "scrollmargin" {
				f.AddWindowSelect(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "0|1|2|3|4|5|6|7|8|9|10", col, row, 3, 1, nil, "", "")
				f.SetIndex(k, 3)
			} else if k == "tabsize" {
				f.AddWindowSelect(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "2|3|4|5|6|7|8|9|10", col, row, 3, 1, nil, "", "")
			} else if k == "lang" {
				Langs := ""
				langs := GeTFileListFromPath(configDir+"/langs", "lang")
				for _, l := range langs {
					if Langs == "" {
						Langs = l
					} else {
						Langs = Langs + "|" + l
					}
				}
				f.AddWindowSelect(k, k+" ", globalSettings[k].(string), Langs, col, row, 0, 1, nil, "", "")
			} else if k == "colorscheme" {
				Colors := ""
				colors := GeTFileListFromPath(configDir+"/colorschemes", "micro")
				for _, c := range colors {
					if Colors == "" {
						Colors = c
					} else {
						Colors = Colors + "|" + c
					}
				}
				f.AddWindowSelect(k, k+" ", globalSettings["colorscheme"].(string), Colors, col, row, 0, 1, nil, "", "")
				f.SetIndex(k, 3)
			} else if k == "cursorshape" {
				f.AddWindowSelect(k, k+" ", globalSettings["cursorshape"].(string), "disabled|block|ibeam|underline", col, row, 0, 1, nil, "", "")
			} else {
				kind := reflect.TypeOf(globalSettings[k]).Kind()
				if kind == reflect.Bool {
					f.AddWindowCheckBox(k, k, "true", col, row, globalSettings[k].(bool), nil, "", "")
				} else if kind == reflect.String {
					f.AddWindowTextBox(k, k+" ", globalSettings[k].(string), "string", col, row, 10, 20, nil, "", "")
				} else if kind == reflect.Float64 {
					f.AddWindowTextBox(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "integer", col, row, 5, 10, nil, "", "")
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
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", width-Count(lbl)-20, height-1, m.ButtonFinish, "", "")
		lbl = Language.Translate("Save")
		f.AddWindowButton("save", " "+lbl+" ", "ok", width-Count(lbl)-5, height-1, m.SaveSettings, "", "")
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
	values := m.myapp.GetValues()
	if values["indentchar"] == "t" {
		values["indentchar"] = "\t"
	} else {
		values["indentchar"] = " "
	}
	if values["cursorcolor"] == "" {
		values["cursorcolor"] = "disabled"
	}
	for k, _ := range globalSettings {
		if strings.Contains(k, "mi-") {
			continue
		}
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
		if v != values[k] {
			save = true
			SetOption(k, values[k])
			if k == "lang" {
				Language.ChangeLangName(values[k], configDir+"/langs/"+values[k]+".lang")
			}
		}
	}
	if save {
		err := WriteSettings(configDir + "/settings.json")
		if err != nil {
			messenger.Error(Language.Translate("Error writing settings.json file"), ": ", err.Error())
		}
	}
	CurView().SetCursorEscapeString()
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
	var f *Frame
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
			if strings.Contains(k, "Mouse") || strings.Contains(k, "Scroll") || strings.Contains(k, "Outdent") || strings.Contains(k, "Indent") || strings.Contains(k, "Backs") || strings.Contains(k, "Delete") || strings.Contains(k, "Insert") || strings.Index(k, "End") == 0 || strings.Contains(k, "Escape") || strings.Contains(k, "Cloud") {
				continue
			}
			switch k {
			case "CursorDown", "CursorUp", "CursorLeft", "CursorRight", "FindNext", "FindPrevious", "SelectLeft", "SelectRight", "SelectDown", "SelectUp", "CursorPageDown", "CursorPageUp":
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
			case "Left", "Up", "Down", "Right", "Backspace", "Delete", "CtrlH", "Esc", "Enter", "Backspace2", "ShiftLeft", "ShiftRight", "ShiftUp", "ShiftDown", "PageDown", "PageUp":
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
		f = m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		f.AddWindowBox("enc", Language.Translate("KeyBindings"), 0, 0, width, height, true, nil, "", "")
		row := 2
		col := 2
		for _, k := range keys {
			_, ok := bindings[k]
			if ok == false {
				bindings[k] = ""
			}
			str := strings.Split(k, "!")
			if len(str) == 1 {
				f.AddWindowTextBox(k, fmt.Sprintf("%-23s", str[0]), bindings[k], "string", col, row, 17, 40, m.SetBinding, "", "")
			} else {
				f.AddWindowTextBox(k, fmt.Sprintf("{grn}%-23s{/grn}", str[0]), bindings[k], "string", col, row, 17, 40, m.SetBinding, "", "")
			}
			f.SetgName(k, f.elements[k].value)
			row++
			if row > height-2 {
				row = 2
				col += 42
			}
		}
		row++
		f.AddWindowTextBox("?test", fmt.Sprintf("{yw}%-23s{/yw}", Language.Translate("Key check")), "", "string", col, row, 17, 40, m.SetBinding, "", "")
		row++
		f.AddWindowLabel("?msg", "", col, row, nil, "", "")
		lbl := Language.Translate("Cancel")
		f.AddWindowButton("?cancel", " "+lbl+" ", "cancel", width-Count(lbl)-20, 32, m.ButtonFinish, "", "")
		lbl = Language.Translate("Save")
		f.AddWindowButton("?save", " "+lbl+" ", "ok", width-Count(lbl)-5, 32, m.SaveKeyBindings, "", "")
	}
	m.myapp.Start()
	f.SetFocus("?test", "B")
	apprunning = m.myapp
}

func (m *microMenu) SetBinding(name, value, event, when string, x, y int) bool {
	f := m.myapp.frames["f"]
	if event == "mouse-click1" {
		if name == "?test" {
			f.SetValue(name, "")
		}
		f.SetFocus(name, "B")
		return true
	}
	if strings.Contains(event, "mouse") || strings.Contains(event, "Alt+Ctrl") || when == "POST" || Count(event) < 2 {
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
		} else if strings.Contains(event, "Backspace") {
			event = strings.ReplaceAll(event, "+", "-")
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
	switch event {
	case "Left", "Right", "Down", "Up", "Esc", "Enter", "Tab", "Backspace2", "Backspace", "Delete", "PgDn", "PgUp", "Shift+Left", "Shift+Right", "Shift+Up", "Shift+Down":
		f.SetLabel("?msg", event+" {red}micro-ide{/red}")
		if event == "Delete" || event == "Backspace" || event == "Backspace2" {
			f.SetValue(name, "")
		}
		f.SetFocus(name, "B")
		return false
	case "Alt-0", "Alt-1", "Alt-2", "Alt-3", "Alt-4", "Alt-5", "Alt-6", "Alt-7", "Alt-8", "Alt-9", "F9", "F10", "F11", "F12":
		f.SetLabel("?msg", event+" {red}reserved plugins{/red}")
		if name == "?test" {
			f.SetValue(name, "")
		}
		return false
	}
	free := true
	for _, e := range f.elements {
		if e.value == event {
			free = false
			if e.name != "?test" {
				f.SetLabel("?msg", "{red}"+Language.Translate("Taken")+" : {/red}"+e.label)
			}
			if name != "?test" {
				f.SetValue(e.name, "")
			}
			break
		}
	}
	_, ok := findKey(event)
	if ok == false {
		f.SetLabel("?msg", "{red}"+Language.Translate("Error")+" : {/red}"+event)
		if name == "?test" {
			f.SetValue(name, "")
		}
		return false
	}
	f.SetValue(name, event)
	if free {
		f.SetLabel("?msg", "{grn}"+Language.Translate("Free")+"{/grn}")
	}
	return false
}

func (m *microMenu) SaveKeyBindings(name, value, event, when string, x, y int) bool {
	f := m.myapp.frames["f"]
	if event != "mouse-click1" {
		return true
	}
	if when == "POST" {
		return true
	}
	write := false
	save := make(map[string]string)
	actionToKey := make(map[string]string)
	values := m.myapp.GetValues()
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
		if k != "" && actionToKey[ak[0]] != k {
			save[k] = ak[0]
			write = true
			BindKey(k, a)
		}
		e := f.elements[a]
		if e.gname != "" && e.gname != k {
			if defaults[e.gname] == "" {
				BindKey(e.gname, "UnbindKey")
			} else if defaults[e.gname] != ak[0] {
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
	if m.myapp == nil || m.myapp.name != "mi-pluginmanager" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-pluginmanager")
		} else {
			m.myapp.name = "mi-pluginmanager"
		}
		width := 110
		height := 37
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		m.myapp.AddStyle("gold", "#ffd700,#262626")
		m.myapp.AddStyle("title", "underline #0087d7,#262626")
		m.myapp.AddStyle("button", "bold #ffffff,#4e4e4e")
		f := m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		f.AddWindowBox("enc", Language.Translate("Plugin Manager"), 0, 0, width, height, true, nil, "", "")
		lbl0 := Language.Translate("Install")
		f.AddWindowButton("install", lbl0, "cancel", width-Count(lbl0)-3, height-1, m.InstallPlugin, "", "")
		lbl1 := Language.Translate("Remove")
		f.AddWindowButton("remove", lbl1, "cancel", width-Count(lbl0)-Count(lbl1)-8, height-1, m.RemovePlugin, "", "")
		f.SetVisible("install", false)
		f.SetVisible("remove", false)
		lbl0 = Language.Translate("Languages")
		f.AddWindowButton("langs", lbl0, "", 1, 2, m.ChangeSource, "button", "")
		lbl1 = Language.Translate("Coding Plugins")
		offset := 1 + Count(lbl0) + 3
		f.AddWindowButton("codeplugins", lbl1, "", offset, 2, m.ChangeSource, "button", "")
		lbl0 = Language.Translate("Application Plugins")
		offset += Count(lbl1) + 3
		f.AddWindowButton("apps", lbl0, "", offset, 2, m.ChangeSource, "button", "")
		lbl0 = Language.Translate("Exit")
		f.AddWindowButton("exit", lbl0, "ok", width-Count(lbl0)-3, 2, m.ButtonFinish, "", "")
		f.AddWindowLabel("title", "", 1, 4, nil, "title", "")
		f.AddWindowSelect("list", "list", "x", "x", 1, 5, 0, 1, nil, "", "")
		f.AddWindowLabel("msg", "", 1, 3, nil, "gold", "")
		f.SetVisible("list", false)
		f.SetVisible("title", false)
	}
	//m.myapp.debug = true
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) ChangeSource(name, value, event, when string, x, y int) bool {
	if when == "PRE" {
		return true
	}
	if event != "mouse-click1" {
		return true
	}
	const MAX = 30
	f := m.myapp.frames["f"]
	f.elements["langs"].style = StringToStyle("bold #ffffff,#4e4e4e")
	f.elements["codeplugins"].style = StringToStyle("bold #ffffff,#4e4e4e")
	f.elements["apps"].style = StringToStyle("bold #ffffff,#4e4e4e")
	sel := ""
	val := ""
	str := ""
	height := 0
	f.DeleteElement("list")
	if name == "langs" {
		var keys []string
		f.elements["langs"].style = StringToStyle("bold #ffffff,#878700")
		f.SetLabel("msg", Language.Translate("Downloading list of")+" "+Language.Translate("Languages")+", "+Language.Translate("please wait")+"...")
		langs := GetAvailableLanguages()
		for k, _ := range langs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		list := ""
		for _, l := range keys {
			url := langs[l]
			tokens := strings.Split(url, "/")
			fileName := tokens[len(tokens)-1]
			val = "langs?" + l + "?" + url
			if PathExists(configDir + "/langs/" + fileName) {
				str = fmt.Sprintf("%-50s✓", l)
			} else {
				str = fmt.Sprintf("%-50s ", l)
			}
			if list == "" {
				list = val + "]" + str
			} else {
				list = list + "|" + val + "]" + str
			}
		}
		if len(langs) > MAX {
			height = MAX
		} else {
			height = len(langs)
		}
		f.SetLabel("title", Language.Translate("Available"))
		f.AddWindowSelect("list", "", "", list, 1, 5, 0, height, nil, "", "")
		f.SetVisible("title", true)
		f.SetVisible("remove", false)
	} else {
		var plugins PluginPackages
		list := ""
		if name == "codeplugins" {
			f.elements["codeplugins"].style = StringToStyle("bold #ffffff,#878700")
			f.SetLabel("msg", Language.Translate("Downloading list of")+" "+Language.Translate("Coding Plugins")+", "+Language.Translate("please wait")+"...")
			plugins = SearchPlugin([]string{"language"})
		} else {
			f.elements["apps"].style = StringToStyle("bold #ffffff,#878700")
			f.SetLabel("msg", Language.Translate("Downloading list of")+" "+Language.Translate("Application Plugins")+", "+Language.Translate("please wait")+"...")
			plugins = SearchPlugin([]string{"application"})
		}
		for _, p := range plugins {
			if name == "codeplugins" && p.Tags[0] != "language" {
				continue
			} else if name == "apps" && p.Tags[0] != "application" {
				continue
			} else if name == "codeplugin" && p.Filetype != p.Name {
				continue
			}
			if Count(p.Description) > 42 {
				desc := []rune(p.Description)
				p.Description = string(desc[0:40])
			}
			cver, _ := semver.Make("0.0.0")
			for _, pv := range p.Versions {
				if pv.Version.GT(cver) {
					cver = pv.Version
				}
			}
			val = name + "?" + p.Author + "?" + strings.ToLower(p.Name) + "?" + cver.String()
			_, loaded := loadedPlugins[strings.ToLower(p.Name)]
			if loaded == false {
				str = fmt.Sprintf("%-20s%-11s %-11s%-18s%-44s  ", p.Name, " ", cver.String(), p.Author, p.Description)
			} else {
				version := GetPluginOption(strings.ToLower(p.Name), "version")
				if version == nil {
					str = fmt.Sprintf("%-20s%-11s %-11s%-18s%-44s  !", p.Name, Language.Translate("restart"), cver.String(), p.Author, p.Description)
				} else {
					sversion := version.(string)
					iver, _ := semver.Make(sversion)
					if iver.LT(cver) {
						str = fmt.Sprintf("%-20s%-11s %-11s%-18s%-44s  *", p.Name, sversion+"*", cver.String(), p.Author, p.Description)
					} else {
						str = fmt.Sprintf("%-20s%-11s %-11s%-18s%-44s  ✓", p.Name, sversion, cver.String(), p.Author, p.Description)
					}
				}
			}
			if list == "" {
				list = val + "]" + str
				sel = val
			} else {
				list = list + "|" + val + "]" + str
			}
		}
		if list == "" {
			f.SetLabel("msg", Language.Translate("No plugins found"))
			return false
		}
		if len(plugins) > 1 {
			sel = ""
		}
		if len(plugins) > MAX {
			height = MAX
		} else {
			height = len(plugins)
		}
		f.SetLabel("title", fmt.Sprintf("%-20s%-11s %-11s%-18s%-48s", Language.Translate("Name"), Language.Translate("Installed"), Language.Translate("Actual"), Language.Translate("Author"), Language.Translate("Description")))
		f.AddWindowSelect("list", "", sel, list, 1, 5, 0, height, nil, "", "")
		f.SetVisible("title", true)
		f.SetVisible("remove", true)
	}
	f.elements["langs"].Draw()
	f.elements["codeplugins"].Draw()
	f.elements["apps"].Draw()
	f.elements["list"].Draw()
	f.SetLabel("msg", "")
	f.SetVisible("install", true)
	return true
}

func (m *microMenu) InstallPlugin(name, value, event, when string, x, y int) bool {
	var plugin []string
	if when == "PRE" {
		return true
	}
	if event != "mouse-click1" {
		return true
	}
	f := m.myapp.frames["f"]
	values := m.myapp.GetValues()
	f.SetVisible("install", false)
	f.SetVisible("remove", false)
	for a, b := range values {
		if a == "list" {
			plugin = strings.Split(b, "?")
			break
		}
	}
	if plugin == nil || len(plugin) < 2 {
		f.SetLabel("msg", Language.Translate("No selection made"))
		f.SetVisible("install", true)
		f.SetVisible("remove", true)
		return true
	}
	if plugin[0] == "langs" {
		// Install Language plugin[2]
		f.SetLabel("msg", Language.Translate("Installing")+" "+Language.Translate("Language")+" "+plugin[1]+", "+Language.Translate("please wait")+"...")
		msg := InstallLanguage(plugin[2])
		f.SetLabel("msg", msg)
	} else {
		// Locate plugin and install
		ptype := "language"
		if plugin[0] == "apps" {
			ptype = "application"
		}
		plugins := SearchPlugin([]string{ptype, plugin[2]})
		if len(plugins) == 0 {
			f.SetLabel("msg", "Plugin not found")
			return true
		}
		cver, _ := semver.Make(plugin[3])
		for _, p := range plugins {
			if p.Author == plugin[1] && p.Tags[0] == ptype && p.Name == plugin[2] {
				for _, v := range p.Versions {
					if v.Version.EQ(cver) {
						err := v.DownloadAndInstall()
						if err != nil {
							f.SetLabel("msg", "Error: "+err.Error())
							f.SetVisible("install", true)
							f.SetVisible("remove", true)
							return true
						}
						goto Exit
					}
				}
			}
		}
	Exit:
	}
	loadedPlugins[plugin[2]] = "disabled"
	m.ChangeSource(plugin[0], "", "mouse-click1", "POST", 1, 1)
	f.SetVisible("install", false)
	f.SetVisible("remove", false)
	f.SetLabel("msg", Language.Translate("Plugin installed")+": "+plugin[2]+"-"+plugin[3])
	go func() {
		time.Sleep(1 * time.Second)
		f.SetVisible("install", true)
		f.SetVisible("remove", true)
	}()
	go func() {
		time.Sleep(4 * time.Second)
		f.SetLabel("msg", "")
	}()

	return true
}

func (m *microMenu) RemovePlugin(name, value, event, when string, x, y int) bool {
	var plugin []string
	if when == "PRE" {
		return true
	}
	if event != "mouse-click1" {
		return true
	}
	f := m.myapp.frames["f"]
	values := m.myapp.GetValues()
	f.SetVisible("install", false)
	f.SetVisible("remove", false)
	for a, b := range values {
		if a == "list" {
			plugin = strings.Split(b, "?")
			break
		}
	}
	if plugin == nil || len(plugin) < 2 {
		f.SetLabel("msg", "No selection made")
		f.SetVisible("install", true)
		f.SetVisible("remove", true)
		return true
	}
	UninstallPlugin(plugin[2])
	m.ChangeSource(plugin[0], "", "mouse-click1", "POST", 1, 1)
	f.SetVisible("install", false)
	f.SetVisible("remove", false)
	f.SetLabel("msg", Language.Translate("Plugin uninstalled")+": "+plugin[2]+"-"+plugin[3])
	go func() {
		time.Sleep(1 * time.Second)
		f.SetVisible("install", true)
		f.SetVisible("remove", true)
	}()
	go func() {
		time.Sleep(4 * time.Second)
		f.SetLabel("msg", "")
	}()
	return true
}

// ---------------------------------------
// Application Search & Replace
// ---------------------------------------

// Search (Find) Dialog

func (m *microMenu) Search(callback func(map[string]string)) {
	var f *Frame
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
		height := 8
		f = m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		m.myapp.AddStyle("f", "black,yellow")
		f.AddWindowBox("enc", Language.Translate("Search"), 0, 0, width, height, true, nil, "", "")
		lbl := Language.Translate("Search regex:")
		f.AddWindowTextBox("find", lbl+" ", "", "string", 2, 2, 61-Count(lbl), 50, m.SubmitSearchOnEnter, "", "")
		f.AddWindowCheckBox("i", "i", "i", 65, 2, false, m.SubmitSearchOnEnter, "", "")
		f.AddWindowLabel("found", "", 2, 4, nil, "", "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 46-Count(lbl), 6, m.ButtonFinish, "", "")
		lbl = Language.Translate("Search")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 64-Count(lbl), 6, m.StartSearch, "", "")
		m.myapp.Finish = m.AbortSearch
	} else {
		f = m.myapp.frames["f"]
	}
	m.myapp.WindowFinish = callback
	m.searchMatch = false
	m.myapp.Start()
	if CurView().Cursor.HasSelection() {
		sel := CurView().Cursor.CurSelection
		if sel[0].Y == sel[1].Y {
			f.SetValue("find", CurView().Cursor.GetSelection())
		}
	}
	f.SetFocus("find", "E")
	apprunning = m.myapp
	m.SubmitSearchOnEnter("find", f.GetValue("find"), "Paste", "POST", 0, 0)
}

// Search & Replace Dialog

func (m *microMenu) SearchReplace(callback func(map[string]string)) {
	var f *Frame
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
		height := 12
		f = m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		m.myapp.AddStyle("f", "bold black,yellow")
		f.AddWindowBox("enc", Language.Translate("Search / Replace"), 0, 0, width, height, true, nil, "", "")
		lbl := Language.Translate("Search regex:")
		f.AddWindowTextBox("search", lbl+" ", "", "string", 2, 2, 63-Count(lbl), 50, m.SubmitSearchOnEnter, "", "")
		lbl = Language.Translate("Replace regex:")
		f.AddWindowTextBox("replace", lbl+" ", "", "string", 2, 4, 54-Count(lbl), 40, m.SubmitSearchOnEnter, "", "")
		lbl = Language.Translate("ignore case")
		offset := Count(lbl) + 6
		f.AddWindowCheckBox("i", lbl, "i", 2, 6, false, m.SubmitSearchOnEnter, "", "")
		lbl = Language.Translate("all")
		f.AddWindowCheckBox("a", lbl, "a", offset, 6, false, m.SubmitSearchOnEnter, "", "")
		offset += Count(lbl) + 4
		f.AddWindowCheckBox("s", `s (\n)`, "s", offset, 6, false, m.SubmitSearchOnEnter, "", "")
		offset += Count(`s (\n)`) + 4
		lbl = Language.Translate("No regexp")
		f.AddWindowCheckBox("l", lbl, "l", offset, 6, false, m.SubmitSearchOnEnter, "", "")
		f.AddWindowLabel("found", "", 2, 8, nil, "", "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 46-Count(lbl), 10, m.ButtonFinish, "", "")
		lbl = Language.Translate("Search")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 64-Count(lbl), 10, m.StartSearch, "", "")
		m.myapp.Finish = m.AbortSearch
	} else {
		f = m.myapp.frames["f"]
	}
	m.myapp.WindowFinish = callback
	m.searchMatch = false
	m.myapp.Start()
	if CurView().Cursor.HasSelection() {
		sel := CurView().Cursor.CurSelection
		if sel[0].Y == sel[1].Y {
			f.SetValue("search", CurView().Cursor.GetSelection())
		}
	}
	f.SetFocus("search", "E")
	apprunning = m.myapp
	m.SubmitSearchOnEnter("search", f.GetValue("search"), "", "POST", 0, 0)
}

func (m *microMenu) AbortSearch(s string) {
	var resp = make(map[string]string)
	f := m.myapp.frames["f"]
	resp["search"] = ""
	f.SetValue("search", "")
	f.SetLabel("found", "")
	m.myapp.WindowFinish(resp)
	m.Finish("Search Abort")
}

func (m *microMenu) StartSearch(name, value, event, when string, x, y int) bool {
	f := m.myapp.frames["f"]
	if event != "mouse-click1" {
		return true
	}
	if when == "POST" {
		return true
	}
	if m.searchMatch == false {
		f.SetValue("search", "")
	}
	m.myapp.WindowFinish(m.myapp.GetValues())
	m.Finish("Done")
	return true
}

func (m *microMenu) SubmitSearchOnEnter(name, value, event, when string, x, y int) bool {
	f := m.myapp.frames["f"]
	if strings.Contains(event, "mouse") && event != "mouse-click1" {
		return true
	}
	if event == "Enter" && when == "PRE" && (name == "find" || name == "replace") {
		if m.searchMatch == false {
			f.SetValue("search", "")
			f.SetValue("find", "")
		}
		m.myapp.WindowFinish(m.myapp.GetValues())
		m.Finish("Done")
		return false
	}
	if name == "replace" {
		return true
	}
	if when == "PRE" {
		return true
	}
	// Need to get search values
	if name != "find" && name != "search" {
		value = f.GetValue("find")
		if value == "" {
			value = f.GetValue("search")
		}
	}
	if f.GetChecked("i0") || f.GetChecked("s0") {
		if f.GetChecked("i0") {
			value = "(?i)" + value
		}
		if f.GetChecked("s0") {
			value = "(?s)" + value
		}
		event = ""
	} else if event == "mouse-click1" {
		event = ""
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
	f.SetLabel("found", found)
	return true
}

// ---------------------------------------
// Application Save As ...
// ---------------------------------------

func (m *microMenu) SaveAs(b *Buffer, usePlugin bool, callback func(map[string]string)) {
	var f *Frame
	if m.myapp == nil || m.myapp.name != "mi-saveas" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-saveas")
		} else {
			m.myapp.name = "mi-saveas"
		}
		enc := ioencoder.New()
		ENCODINGS := enc.GetAvailableEncodingsAsString("|")
		ENCODINGS = "UTF8|" + ENCODINGS
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 80
		height := 17
		f = m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		f.AddWindowBox("enc", Language.Translate("Save As ..."), 0, 0, width, height, true, nil, "", "")
		lbl := Language.Translate("File name :")
		f.AddWindowTextBox("filename", lbl+" ", "", "string", 2, 2, 76-Count(lbl), 200, m.SaveFile, "", "")
		lbl = Language.Translate("Encoding:")
		f.AddWindowSelect("encoding", lbl+" ", "", ENCODINGS, 2, 4, 0, 12, m.SaveAsEncodingEvent, "", "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 56-Count(lbl), height-2, m.SaveAsButtonFinish, "", "")
		lbl = Language.Translate("Save")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 75-Count(lbl), height-2, m.SaveFile, "", "")
	} else {
		f = m.myapp.frames["f"]
	}
	m.myapp.WindowFinish = callback
	m.usePlugin = usePlugin
	encoder := b.encoder
	encoder = strings.ReplaceAll(encoder, "-", "")
	encoder = strings.ReplaceAll(encoder, "_", "")
	f.SetValue("encoding", encoder)
	f.SetValue("filename", b.Path)
	m.myapp.Finish = m.SaveAsFinish
	m.myapp.Start()
	f.SetFocus("filename", "E")
	apprunning = m.myapp
}

func (m *microMenu) SaveAsEncodingEvent(name, value, event, when string, x, y int) bool {
	f := m.myapp.frames["f"]
	if when == "POST" {
		return true
	}
	if event == "mouse-click1" {
		f.SetValue("encode", "")
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
	values := m.myapp.GetValues()
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
// Set Buffer FileType
// ---------------------------------------

func (m *microMenu) SelFileType(x int) {
	// For this case we need to always rebuild, because the hotspots move depending on the window and view where is located
	var ftypes []string
	m.myapp = nil
	m.myapp = new(MicroApp)
	m.myapp.New("mi-ftype")
	m.myapp.Reset()
	m.myapp.defStyle = StringToStyle("#ffffff,#3a3a3a")
	m.myapp.AddStyle("normal", "#ffffff,#3a3a3a")
	_, h := screen.Size()
	for _, rtf := range ListRuntimeFiles(RTSyntax) {
		data, err := rtf.Data()
		if err == nil {
			file, err := highlight.ParseFile(data)
			if err == nil {
				if len(file.FileType) > 12 {
					ftypes = append(ftypes, file.FileType[0:13])
				} else {
					ftypes = append(ftypes, file.FileType)
				}
			}
		}
	}
	height := h - 4
	if len(ftypes) < height {
		height = len(ftypes)
	}
	col := 0
	row := 0
	f := m.myapp.AddFrame("f", h-height-2, x-8, 1, height, "close")
	pft := ""
	value := ""
	sort.Strings(ftypes)
	for _, ft := range ftypes {
		if pft == ft {
			continue
		}
		pft = ft
		if len(ft) < 14 {
			value = fmt.Sprintf("%-14s", ft)
		} else {
			value = ft
		}
		f.AddWindowLabel(ft, value, col, row, m.SetFtype, "normal", "")
		row++
		if row > height-1 {
			row = 0
			col += 14
		}
	}
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) SetFtype(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	e := m.myapp.frames[m.myapp.activeFrame].elements[name]
	if event == "mouseout" {
		e.style = m.myapp.defStyle
		e.Draw()
		return true
	}
	if event == "mousein" {
		e.style = e.style.Foreground(tcell.ColorBlack).Background(tcell.Color220)
		e.Draw()
		return true
	}
	if event != "mouse-click1" {
		return true
	}
	SetLocalOption("filetype", name, CurView())
	m.Finish("FileType")
	return true
}

// ---------------------------------------
// Set Buffer Tab Space indent
// ---------------------------------------

func (m *microMenu) SelTabSpace(x, y int) {
	// For this case we need to always rebuild, because the hotspots move depending on the window and view where is located
	m.myapp = nil
	m.myapp = new(MicroApp)
	m.myapp.New("mi-tabspc")
	m.myapp.Reset()
	m.myapp.defStyle = StringToStyle("#ffffff,#3a3a3a")
	m.myapp.AddStyle("tab", "#afd7ff,#3a3a3a")
	m.myapp.AddStyle("spc", "#ffffff,#3a3a3a")
	height := 14
	row := 0
	f := m.myapp.AddFrame("f", y-height, x-3, 1, height, "close")
	for i := 2; i < 9; i++ {
		ft := " Tab " + strconv.Itoa(i) + " "
		f.AddWindowLabel(ft, ft, 0, row, m.SetTabSpace, "tab", "")
		f.SetValue(ft, strconv.Itoa(i))
		row++
	}
	for i := 2; i < 9; i++ {
		ft := " Spc " + strconv.Itoa(i) + " "
		f.AddWindowLabel(ft, ft, 0, row, m.SetTabSpace, "spc", "")
		f.SetValue(ft, strconv.Itoa(i))
		row++
	}
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) SetTabSpace(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	e := m.myapp.frames[m.myapp.activeFrame].elements[name]
	if event == "mouseout" {
		if strings.Contains(name, "Spc") {
			e.style = m.myapp.styles["spc"].style
		} else {
			e.style = m.myapp.styles["tab"].style
		}
		e.Draw()
		return true
	}
	if event == "mousein" {
		e.style = e.style.Foreground(tcell.ColorBlack).Background(tcell.Color220)
		e.Draw()
		return true
	}
	if event != "mouse-click1" {
		return true
	}
	iChar := "\t"
	if strings.Contains(name, "Spc") {
		iChar = " "
	}
	b := CurView().Buf
	n, _ := strconv.ParseFloat(value, 64)
	b.ChangeIndentation(b.Settings["indentchar"].(string), iChar, int(b.Settings["tabsize"].(float64)), int(n))
	SetLocalOption("indentchar", iChar, CurView())
	SetLocalOption("tabsize", value, CurView())
	m.Finish("TabSpaceType")
	return true
}

// ---------------------------------------
// Local Configuracion : language or file
// ---------------------------------------

const mmBufferSettings = "autoclose,autoindent,eofnewline,fileformat,indentchar,keepautoindent,matchbrace,rmtrailingws,smartindent,smartpaste,softwrap,tabindents,tabmovement,tabstospaces,tabsize"

func (m *microMenu) SelLocalSettings(b *Buffer) {
	var f *Frame
	if m.myapp == nil || m.myapp.name != "mi-localsettings-"+b.name {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-localsettings-" + b.name)
		} else {
			m.myapp.name = "mi-localsettings" + b.name
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 80
		height := 16
		f = m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		f.AddWindowBox("enc", Language.Translate("Buffer Settings"), 0, 0, width, height, true, nil, "", "")
		f.AddWindowRadio("savefor", Language.Translate("Save as this file settings only"), "file", 2, height-4, false, nil, "", "")
		f.AddWindowRadio("savefor", fmt.Sprintf(Language.Translate("Save as default settings for all (%s) files"), b.FileType()), "lang", 2, height-3, false, nil, "", "")
		f.AddWindowRadio("savefor", Language.Translate("Change settings for this session only"), "none", 2, height-2, true, nil, "", "")
		lbl := Language.Translate("Save")
		f.AddWindowButton("set", lbl, "ok", width-Count(lbl)-3, height-1, m.SetLocalSettings, "", "")
		w := Count(Language.Translate("Cancel"))
		f.AddWindowButton("cancel", Language.Translate("Cancel"), "cancel", width-Count(lbl)-w-8, height-1, m.ButtonFinish, "", "")
		keys := make([]string, 0, len(b.Settings))
		for k := range b.Settings {
			if strings.Contains(mmBufferSettings, k) {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		row := 2
		col := 2
		for _, k := range keys {
			if k == "fileformat" {
				f.AddWindowSelect(k, k+" ", b.Settings[k].(string), "unix|dos", col, row, 0, 1, nil, "", "")
			} else if k == "indentchar" {
				char := "s"
				if b.Settings[k].(string) != " " {
					char = "t"
				}
				f.AddWindowSelect(k, k+" ", char, "t]Tab|s]Space", col, row, 0, 1, nil, "", "")
			} else if k == "tabsize" {
				f.AddWindowSelect(k, k+" ", fmt.Sprintf("%g", b.Settings[k].(float64)), "2|3|4|5|6|7|8|9|10", col, row, 3, 1, nil, "", "")
			} else {
				kind := reflect.TypeOf(b.Settings[k]).Kind()
				if kind == reflect.Bool {
					f.AddWindowCheckBox(k, k, "true", col, row, b.Settings[k].(bool), nil, "", "")
				} else if kind == reflect.String {
					f.AddWindowTextBox(k, k+" ", b.Settings[k].(string), "string", col, row, 10, 20, nil, "", "")
				} else if kind == reflect.Float64 {
					f.AddWindowTextBox(k, k+" ", fmt.Sprintf("%g", b.Settings[k].(float64)), "integer", col, row, 5, 10, nil, "", "")
				} else {
					continue
				}
			}
			row += 2
			if row > height-6 {
				row = 2
				col += 20
			}
		}
	} else {
		f = m.myapp.frames["f"]
	}
	f.SetValue("encode", "")
	m.myapp.Start()
	apprunning = m.myapp
}

func (m *microMenu) SetLocalSettings(name, value, event, when string, x, y int) bool {
	if when == "POST" {
		return true
	}
	if event != "mouse-click1" {
		return true
	}
	values := m.myapp.GetValues()
	if values["indentchar"] == "t" {
		values["indentchar"] = "\t"
	} else {
		values["indentchar"] = " "
	}
	// Check if we need to indent buffer
	b := CurView().Buf
	n, _ := strconv.ParseFloat(values["tabsize"], 64)
	if b.Settings["indentchar"].(string) != values["indentchar"] || int(b.Settings["tabsize"].(float64)) != int(n) {
		b.ChangeIndentation(b.Settings["indentchar"].(string), values["indentchar"], int(b.Settings["tabsize"].(float64)), int(n))
	}
	// Get all keys we need, checkboxes return no values if not selected
	keys := strings.Split(mmBufferSettings, ",")
	for _, k := range keys {
		if _, ok := values[k]; !ok {
			values[k] = "false"
		}
	}
	// Assing options selected to current view
	for k, v := range values {
		if k != "savefor" {
			SetLocalOption(k, v, CurView())
		}
	}
	// Set destination for this settings, default language
	fname := configDir + "/settings/" + b.FileType() + ".json"
	if values["savefor"] == "file" {
		// Add current filetype selected too
		values["encoder"] = b.buf.encoder
		values["filetype"] = b.FileType()
		// Change destintation to this file only
		absFilename := ReplaceHome(b.AbsPath)
		fname = configDir + "/buffers/" + strings.ReplaceAll(absFilename+".settings", "/", "")
	}
	if values["savefor"] != "none" {
		// Remove non buffer settings values
		delete(values, "savefor")
		// Update JSON file
		UpdateFileJSON(fname, values)
	}
	m.Finish("settings")
	return true
}

// ---------------------------------------
// Set Buffer Encoding
// ---------------------------------------

func (m *microMenu) SelEncoding(encoder string, callback func(map[string]string)) {
	var f *Frame
	if m.myapp == nil || m.myapp.name != "mi-selencoding" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-selencoding")
		} else {
			m.myapp.name = "mi-selencoding"
		}
		enc := ioencoder.New()
		ENCODINGS := enc.GetAvailableEncodingsAsString("|")
		ENCODINGS = "UTF8|" + ENCODINGS
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 60
		height := 15
		f = m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		f.AddWindowBox("enc", Language.Translate("Select Encoding"), 0, 0, width, height, true, nil, "", "")
		lbl := Language.Translate("Encoding:")
		f.AddWindowSelect("encoding", lbl+" ", "", ENCODINGS, 2, 2, 0, 12, nil, "", "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 43, height-4, m.ButtonFinish, "", "")
		lbl = Language.Translate("Set encoding")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 43, height-2, m.SetEncoding, "", "")
	} else {
		f = m.myapp.frames["f"]
	}
	m.myapp.WindowFinish = callback
	encoder = strings.ReplaceAll(encoder, "-", "")
	encoder = strings.ReplaceAll(encoder, "_", "")
	f.SetValue("encoding", encoder)
	f.SetValue("encode", "")
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
	values := m.myapp.GetValues()
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
// Directory view
// ---------------------------------------

const MINWIDTH = 26

func (m *microMenu) DirTreeView() {
	m.myapp = nil
	m.myapp = new(MicroApp)
	m.myapp.New("mi-dirview")
	dir, width := m.getDir()
	_, height := screen.Size()
	height -= 3
	if width < MINWIDTH {
		width = MINWIDTH
	}
	m.myapp.Reset()
	m.myapp.defStyle = StringToStyle("#ffffff,#1c1c1c")
	m.myapp.AddStyle("d", "#A6E22E,#1c1c1c")
	f := m.myapp.AddFrame("f", 1, 0, width+2, height, "fixed")
	f.AddWindowBox("dbox", "", 0, 0, width+2, height, true, nil, "", "")
	f.AddWindowSelect("dirview", "", "../", dir, 1, 1, width, height-1, m.TreeViewEvent, "", "")
	m.myapp.Start()
	apprunning = m.myapp
	m.myapp.CheckElementsActions("mouse-click1", 1, 1)
}

func (m *microMenu) TreeViewEvent(name, value, event, when string, x, y int) bool {
	if event == "mouse-click1" {
		return true
	} else if when == "POST" {
		return false
	}
	reset := false
	f := m.myapp.frames["f"]
	if event == "mouse-doubleclick1" || event == "Enter" {
		// Open file in new Tab
		if strings.Contains(value, "/") {
			backdir := ""
			if strings.Contains(value, "../") {
				x := strings.LastIndex(m.LastPath, "/")
				fname := string([]rune(m.LastPath)[:x])
				x = strings.LastIndex(fname, "/") + 1
				backdir = string([]rune(m.LastPath)[x:])
				fname = string([]rune(m.LastPath)[:x])
				m.LastPath = fname
			} else {
				m.LastPath = m.LastPath + value
			}
			dir, width := m.getDir()
			if width < MINWIDTH {
				width = MINWIDTH
			}
			height := m.myapp.frames["f"].oheight - 1
			if f.elements["dbox"].width > width {
				reset = true
			}
			f.elements["dbox"].width = width + 2
			f.AddWindowSelect("dirview", "", "", dir, 1, 1, width, height-1, m.TreeViewEvent, "", "")
			m.myapp.CheckElementsActions("mouse-click1", 1, 1)
			if backdir != "" {
				for k, v := range f.elements["dirview"].opts {
					if v.value == backdir {
						f.elements["dirview"].offset = k
						f.elements["dirview"].value = f.elements["dirview"].opts[k].value
						break
					}
				}
			}
			if reset {
				m.myapp.ResetFrames()
			} else {
				m.myapp.DrawAll()
			}
			return false
		} else {
			for i, t := range tabs {
				for _, v := range t.Views {
					if strings.Contains(v.Buf.Path, value) {
						curTab = i
						messenger.Information(value, " : Focused")
						m.Finish("focuse same file")
						return false
					}
				}
			}
			NewTab([]string{m.LastPath + value})
			CurView().Buf.name = value
			m.myapp.activeElement = name
			messenger.Information(value, " "+Language.Translate("opened in new Tab"))
			m.myapp.ResetFrames()
			return false
		}
	} else if event == "mouse-click3" || event == "mouse-click2" || event == "Right" {
		if event == "Right" && strings.Contains(value, "/") {
			m.myapp.activeElement = name
			return true
		}
		if CurView().Buf.Modified() {
			messenger.Error(Language.Translate("You need to save view first"))
			m.Finish("file dirty")
			return false
		}
		messenger.Information(value, " "+Language.Translate("opened in View"))
		CurView().Open(m.LastPath + value)
		CurView().Buf.name = value
		m.myapp.activeElement = name
		m.myapp.ResetFrames()
	}
	m.myapp.activeElement = name
	return true
}

func (m *microMenu) getDir() (string, int) {
	var dir string
	var s string

	if m.LastPath == "" {
		m.LastPath = "/"
	}
	width := 0
	dir = "../]{d}🖿   ../"
	files, _ := ioutil.ReadDir(m.LastPath)
	for _, f := range files {
		if f.IsDir() {
			s = "/]{d}🖿  " + f.Name() + "/"
		} else {
			s = "] 🗎 " + f.Name()
		}
		if Count(s) > width {
			width = Count(s)
		}
		dir = dir + "|" + f.Name() + s
	}
	return dir, width
}

// ---------------------------------------
// General Routines
// ---------------------------------------

func (m *microMenu) Finish(s string) {
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

// ---------------------------------------
// Micro Ide Cloud Services
// ---------------------------------------

func (m *microMenu) MiCloudServices() {
	var f *Frame
	if m.myapp == nil || m.myapp.name != "mi-cloud" {
		if m.myapp == nil {
			m.myapp = new(MicroApp)
			m.myapp.New("mi-cloud")
		} else {
			m.myapp.name = "mi-cloud"
		}
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 70
		height := 12
		row := 2
		f = m.myapp.AddFrame("f", -1, -1, width, height, "relative")
		m.myapp.AddStyle("f", "black,yellow")
		f.AddWindowBox("enc", Language.Translate("Micro Ide Cloud Services"), 0, 0, width, height, true, nil, "", "")
		f.AddWindowTextBox("ukey", fmt.Sprintf("%25s ", Language.Translate("Key")), globalSettings["mi-key"].(string), "string", 2, row, 40, 40, nil, "", "")
		row += 2
		f.AddWindowTextBox("pass", fmt.Sprintf("%25s ", Language.Translate("Password")), globalSettings["mi-pass"].(string), "string", 2, row, 20, 20, nil, "", "")
		row += 2
		if globalSettings["mi-pass"].(string) != "" {
			f.AddWindowTextBox("newpass", fmt.Sprintf("%25s ", Language.Translate("New Password")), "", "string", 2, row, 20, 20, nil, "", "")
			row += 2
		}
		f.AddWindowTextBox("phrs", fmt.Sprintf("%25s ", Language.Translate("Passphrase")), globalSettings["mi-phrase"].(string), "string", 2, row, 20, 20, nil, "", "")
		lbl := Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", width-Count(lbl)-20, height-1, m.ButtonFinish, "", "")
		lbl = Language.Translate("Save")
		f.AddWindowButton("save", " "+lbl+" ", "ok", width-Count(lbl)-5, height-1, m.SaveCloudSettings, "", "")
	} else {
		f = m.myapp.frames["f"]
	}
	m.myapp.Start()
	apprunning = m.myapp
	f.SetFocus("ukey", "begin")
}

func (m *microMenu) SaveCloudSettings(name, value, event, when string, x, y int) bool {
	if when == "PRE" {
		return true
	}
	if event != "mouse-click1" {
		return true
	}
	reset := false
	setup := false
	passchg := false
	savesettings := false
	values := m.myapp.GetValues()
	if values["ukey"] != "" && values["ukey"] != globalSettings["mi-key"].(string) {
		SetOption("mi-key", values["ukey"])
		savesettings = true
	}
	if values["pass"] != "" && globalSettings["mi-pass"].(string) == "" {
		SetOption("mi-pass", values["pass"])
		setup = true
		savesettings = true
	} else if values["pass"] != "" && values["newpass"] != "" && values["newpass"] != globalSettings["mi-pass"].(string) {
		SetOption("mi-pass", values["newpass"])
		savesettings = true
		passchg = true
	}
	if values["phrs"] != "" && values["phrs"] != globalSettings["mi-phrase"].(string) {
		// Delete cloud stored data (will be useless)
		reset = true
		SetOption("mi-phrase", values["phrs"])
		savesettings = true
	}
	if savesettings {
		Errors := ""
		err := Clip.SetCloudPath(cloudPath, values["ukey"], values["pass"], values["phrs"])
		if err == nil {
			if setup {
				msg := Clip.SetUpCloudService()
				if msg != "" {
					Errors = msg
				}
			}
			if passchg {
				msg := Clip.ChangeCloudPassword(values["newpass"])
				if msg != "" {
					Errors = Errors + "; " + msg
				}
			}
			if reset {
				msg := Clip.ResetCloudService()
				if msg != "" {
					Errors = Errors + "; " + msg
				}
			}
		} else {
			Errors = Errors + "; " + err.Error()
		}
		if Errors != "" {
			messenger.Error(Errors)
		} else {
			err = WriteSettings(configDir + "/settings.json")
			if err != nil {
				messenger.Error(Language.Translate("Error writing settings.json file"), ": ", err.Error())
			}
			messenger.Success("Cloud service setup OK")
		}
	}
	m.Finish("MiServices")
	return true
}
