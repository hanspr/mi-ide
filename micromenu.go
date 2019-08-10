// microapp
package main

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/blang/semver"
	"github.com/hanspr/microidelibs/highlight"
	"github.com/hanspr/tcell"
)

const ENCODINGS = "UTF-8|ISO-8859-1|ISO-8859-2|ISO-8859-15|WINDOWS-1250|WINDOWS-1251|WINDOWS-1252|WINDOWS-1256|SHIFT-JIS|GB2312|EUC-KR|EUC-JP|GBK|BIG-5|ASCII"

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
		f := m.myapp.AddFrame("menu", 1, 0, 1, 1, "fixed")
		m.AddSubmenu(name, "Micro-ide")
		f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", "Micro-ide"), "", 0, row, m.ShowSubmenuItems, "")
		m.AddSubMenuItem("microide", Language.Translate("Global Settings"), m.GlobalConfigDialog)
		m.AddSubMenuItem("microide", Language.Translate("KeyBindings"), m.KeyBindingsDialog)
		m.AddSubMenuItem("microide", Language.Translate("Plugin Manager"), m.PluginManagerDialog)
		row++
		for _, name := range keys {
			if name != "microide" {
				label := fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", m.submenu[name])
				f.AddWindowMenuLabel(name, label, "", 0, row, m.ShowSubmenuItems, "")
			}
			row++
		}
		f.AddWindowMenuBottom("menubottom", fmt.Sprintf("%-"+strconv.Itoa(m.maxwidth+1)+"s", " "), 0, row, nil, "")
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
		f.AddWindowMenuTop("smenubottom", fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", " "), m.maxwidth+2, y, nil, "")
	}
	// Show new submenu
	items := m.submenuElements[name]
	for i, s := range items {
		name := "submenu" + strconv.Itoa(i)
		if i == 0 {
			f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "r", m.maxwidth+2, y, m.MenuItemClick, "")
		} else if i == 1 {
			f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "cl", m.maxwidth+2, y, m.MenuItemClick, "")
		} else {
			f.AddWindowMenuLabel(name, fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", s.label), "", m.maxwidth+2, y, m.MenuItemClick, "")
		}
		f.SetIndex(name, 3)
		f.SetgName(name, "submenu")
		f.SetiKey(name, i)
		y++
	}
	f.AddWindowMenuBottom("smenubottom", fmt.Sprintf("%-"+strconv.Itoa(width+1)+"s", " "), m.maxwidth+2, y, nil, "")
	f.SetgName("smenubottom", "submenu")
	m.myapp.DrawAll()
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
			delete(f.elements, k)
		}
	}
	m.myapp.ResetFrames()
	m.myapp.DrawAll()
	m.activemenu = ""
}

func (m *microMenu) MenuItemClick(name, value, event, when string, x, y int) bool {
	f := m.myapp.frames["menu"]
	if f.elements == nil {
		return false
	}
	_, err := f.elements[name]
	if err == false {
		TermMessage("Error en menu," + name + " no existe")
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
		f.AddWindowBox("enc", Language.Translate("Global Settings"), 0, 0, width, height, true, nil, "")
		keys := make([]string, 0, len(globalSettings))
		for k := range globalSettings {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		row := 2
		col := 2
		for _, k := range keys {
			if k == "fileformat" {
				f.AddWindowSelect(k, k+" ", globalSettings[k].(string), "unix|dos", col, row, 0, 1, nil, "")
			} else if k == "colorcolumn" {
				f.AddWindowTextBox(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "string", col, row, 4, 3, m.ValidateInteger, "")
			} else if k == "indentchar" {
				char := "s"
				if globalSettings[k].(string) != " " {
					char = "t"
				}
				f.AddWindowSelect(k, k+" ", char, "t]Tab|s]Space", col, row, 0, 1, nil, "")
			} else if k == "scrollmargin" {
				f.AddWindowSelect(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "0|1|2|3|4|5|6|7|8|9|10", col, row, 3, 1, nil, "")
				f.SetIndex(k, 3)
			} else if k == "tabsize" {
				f.AddWindowSelect(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "1|2|3|4|5|6|7|8|9|10", col, row, 3, 1, nil, "")
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
				f.AddWindowSelect(k, k+" ", globalSettings[k].(string), Langs, col, row, 0, 1, nil, "")
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
				f.AddWindowSelect(k, k+" ", globalSettings["colorscheme"].(string), Colors, col, row, 0, 1, nil, "")
				f.SetIndex(k, 3)
			} else {
				kind := reflect.TypeOf(globalSettings[k]).Kind()
				if kind == reflect.Bool {
					f.AddWindowCheckBox(k, k, "true", col, row, globalSettings[k].(bool), nil, "")
				} else if kind == reflect.String {
					f.AddWindowTextBox(k, k+" ", globalSettings[k].(string), "string", col, row, 10, 20, nil, "")
				} else if kind == reflect.Float64 {
					f.AddWindowTextBox(k, k+" ", fmt.Sprintf("%g", globalSettings[k].(float64)), "integer", col, row, 5, 10, nil, "")
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
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", width-Count(lbl)-20, height-1, m.ButtonFinish, "")
		lbl = Language.Translate("Save")
		f.AddWindowButton("save", " "+lbl+" ", "ok", width-Count(lbl)-5, height-1, m.SaveSettings, "")
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
			case "CursorDown", "CursorUp", "CursorLeft", "CursorRight", "FindNext", "FindPrevious", "SelectLeft", "SelectRight", "SelectDown", "SelectUp":
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
			case "Left", "Up", "Down", "Right", "Backspace", "Delete", "CtrlH", "Esc", "Enter", "Backspace2", "ShiftLeft", "ShiftRight", "ShiftUp", "ShiftDown":
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
		f.AddWindowBox("enc", Language.Translate("KeyBindings"), 0, 0, width, height, true, nil, "")
		row := 2
		col := 2
		for _, k := range keys {
			_, ok := bindings[k]
			if ok == false {
				bindings[k] = ""
			}
			str := strings.Split(k, "!")
			if len(str) == 1 {
				f.AddWindowTextBox(k, fmt.Sprintf("%-23s", str[0]), bindings[k], "string", col, row, 17, 40, m.SetBinding, "")
			} else {
				f.AddWindowTextBox(k, fmt.Sprintf("{grn}%-23s{/grn}", str[0]), bindings[k], "string", col, row, 17, 40, m.SetBinding, "")
			}
			f.SetgName(k, f.elements[k].value)
			row++
			if row > height-2 {
				row = 2
				col += 42
			}
		}
		row++
		f.AddWindowTextBox("?test", fmt.Sprintf("{yw}%-23s{/yw}", Language.Translate("Key check")), "", "string", col, row, 17, 40, m.SetBinding, "")
		row++
		f.AddWindowLabel("?msg", "", col, row, nil, "")
		lbl := Language.Translate("Cancel")
		f.AddWindowButton("?cancel", " "+lbl+" ", "cancel", width-Count(lbl)-20, 32, m.ButtonFinish, "")
		lbl = Language.Translate("Save")
		f.AddWindowButton("?save", " "+lbl+" ", "ok", width-Count(lbl)-5, 32, m.SaveKeyBindings, "")
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
		f.AddWindowBox("enc", Language.Translate("Plugin Manager"), 0, 0, width, height, true, nil, "")
		lbl0 := Language.Translate("Install")
		f.AddWindowButton("install", lbl0, "cancel", width-Count(lbl0)-3, height-1, m.InstallPlugin, "")
		f.SetVisible("install", false)
		lbl0 = Language.Translate("Languages")
		f.AddWindowButton("langs", lbl0, "", 1, 2, m.ChangeSource, "button")
		lbl1 := Language.Translate("Coding Plugins")
		offset := 1 + Count(lbl0) + 3
		f.AddWindowButton("codeplugins", lbl1, "", offset, 2, m.ChangeSource, "button")
		lbl0 = Language.Translate("Application Plugins")
		offset += Count(lbl1) + 3
		f.AddWindowButton("apps", lbl0, "", offset, 2, m.ChangeSource, "button")
		lbl0 = Language.Translate("Exit")
		f.AddWindowButton("exit", lbl0, "ok", width-Count(lbl0)-3, 2, m.ButtonFinish, "")
		f.AddWindowLabel("title", "", 1, 4, nil, "title")
		f.AddWindowSelect("list", "list", "x", "x", 1, 5, 0, 1, nil, "")
		f.AddWindowLabel("msg", "", 1, 3, nil, "gold")
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
		f.AddWindowSelect("list", "", "", list, 1, 5, 0, height, nil, "")
		f.SetVisible("title", true)
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
			val = "codeplugin?" + p.Author + "?" + p.Name
			version := GetPluginOption(strings.ToLower(p.Name), "version")
			if version == nil {
				str = fmt.Sprintf("%-20s%-10s %-10s%-20s%-44s  ", p.Name, " ", cver.String(), p.Author, p.Description)
			} else {
				iver, _ := semver.Make(version.(string))
				if iver.LT(cver) {
					str = fmt.Sprintf("%-20s%-10s %-10s%-20s%-44s  *", p.Name, version.(string)+"*", cver.String(), p.Author, p.Description)
				} else {
					str = fmt.Sprintf("%-20s%-10s %-10s%-20s%-44s  ✓", p.Name, version.(string), cver.String(), p.Author, p.Description)
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
		f.SetLabel("title", fmt.Sprintf("%-20s%-10s %-10s%-20s%-48s", Language.Translate("Name"), Language.Translate("Installed"), Language.Translate("Actual"), Language.Translate("Author"), Language.Translate("Description")))
		f.AddWindowSelect("list", "", sel, list, 1, 5, 0, height, nil, "")
		f.SetVisible("title", true)
	}
	f.elements["langs"].Draw()
	f.elements["codeplugins"].Draw()
	f.elements["apps"].Draw()
	f.elements["list"].Draw()
	f.SetVisible("install", true)
	f.SetLabel("msg", "")
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
	values := m.myapp.getValues()
	for a, b := range values {
		if a == "list" {
			plugin = strings.Split(b, "?")
			break
		}
	}
	if plugin[0] == "langs" {
		// Install Language plugin[2]
		f.SetLabel("msg", Language.Translate("Installing")+" "+Language.Translate("Language")+" "+plugin[1]+", "+Language.Translate("please wait")+"...")
		msg := InstallLanguage(plugin[2])
		f.SetLabel("msg", msg)
	} else {
		// Locate plugin and install
		f.SetLabel("msg", Language.Translate("Installing")+" "+plugin[1]+"/"+plugin[2]+" "+Language.Translate("please wait")+"...")
	}
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
		heigth := 8
		f = m.myapp.AddFrame("f", -1, -1, width, heigth, "relative")
		m.myapp.AddStyle("f", "black,yellow")
		f.AddWindowBox("enc", Language.Translate("Search"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Search regex:")
		f.AddWindowTextBox("find", lbl+" ", "", "string", 2, 2, 61-Count(lbl), 50, m.SubmitSearchOnEnter, "")
		f.AddWindowCheckBox("i", "i", "i", 65, 2, false, m.SubmitSearchOnEnter, "")
		f.AddWindowLabel("found", "", 2, 4, nil, "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 46-Count(lbl), 6, m.ButtonFinish, "")
		lbl = Language.Translate("Search")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 64-Count(lbl), 6, m.StartSearch, "")
		m.myapp.Finish = m.AbortSearch
		m.myapp.WindowFinish = callback
	} else {
		f = m.myapp.frames["f"]
	}
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
		heigth := 12
		f = m.myapp.AddFrame("f", -1, -1, width, heigth, "relative")
		m.myapp.AddStyle("f", "bold black,yellow")
		f.AddWindowBox("enc", Language.Translate("Search / Replace"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Search regex:")
		f.AddWindowTextBox("search", lbl+" ", "", "string", 2, 2, 63-Count(lbl), 50, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("Replace regex:")
		f.AddWindowTextBox("replace", lbl+" ", "", "string", 2, 4, 54-Count(lbl), 40, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("ignore case")
		offset := Count(lbl) + 6
		f.AddWindowCheckBox("i", lbl, "i", 2, 6, false, m.SubmitSearchOnEnter, "")
		lbl = Language.Translate("all")
		f.AddWindowCheckBox("a", lbl, "a", offset, 6, false, m.SubmitSearchOnEnter, "")
		offset += Count(lbl) + 4
		f.AddWindowCheckBox("s", `s (\n)`, "s", offset, 6, false, m.SubmitSearchOnEnter, "")
		offset += Count(`s (\n)`) + 4
		lbl = Language.Translate("No regexp")
		f.AddWindowCheckBox("l", lbl, "l", offset, 6, false, m.SubmitSearchOnEnter, "")
		f.AddWindowLabel("found", "", 2, 8, nil, "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 46-Count(lbl), 10, m.ButtonFinish, "")
		lbl = Language.Translate("Search")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 64-Count(lbl), 10, m.StartSearch, "")
		m.myapp.Finish = m.AbortSearch
		m.myapp.WindowFinish = callback
	} else {
		f = m.myapp.frames["f"]
	}
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
	m.myapp.WindowFinish(m.myapp.getValues())
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
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 80
		heigth := 8
		f = m.myapp.AddFrame("f", -1, -1, width, heigth, "relative")
		f.AddWindowBox("enc", Language.Translate("Save As ..."), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("File name :")
		f.AddWindowTextBox("filename", lbl+" ", "", "string", 2, 2, 76-Count(lbl), 200, m.SaveFile, "")
		lbl = Language.Translate("Encoding:")
		f.AddWindowSelect("encoding", lbl+" ", b.encoder, ENCODINGS+"|"+b.encoder, 2, 4, 0, 1, m.SaveAsEncodingEvent, "")
		lbl = Language.Translate("Use this encoding:")
		f.AddWindowTextBox("encode", lbl+" ", "", "string", 55-Count(lbl), 4, 15, 15, nil, "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 56-Count(lbl), 6, m.SaveAsButtonFinish, "")
		lbl = Language.Translate("Save")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 75-Count(lbl), 6, m.SaveFile, "")
		m.myapp.WindowFinish = callback
	} else {
		f = m.myapp.frames["f"]
	}
	m.usePlugin = usePlugin
	f.SetValue("filename", b.Path)
	if strings.Contains(ENCODINGS, b.encoder) {
		f.SetValue("encoding", b.encoder)
		f.SetValue("encode", "")
	} else {
		f.SetValue("encoding", "UTF-8")
		f.SetValue("encode", b.encoder)
	}
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
		f.AddWindowLabel(ft, value, col, row, m.SetFtype, "normal")
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
		m.myapp.Reset()
		m.myapp.defStyle = StringToStyle("#ffffff,#262626")
		width := 60
		heigth := 8
		f = m.myapp.AddFrame("f", -1, -1, width, heigth, "relative")
		f.AddWindowBox("enc", Language.Translate("Select Encoding"), 0, 0, width, heigth, true, nil, "")
		lbl := Language.Translate("Encoding:")
		f.AddWindowSelect("encoding", lbl+" ", encoder, ENCODINGS, 2, 2, 0, 1, nil, "")
		lbl = Language.Translate("Use this encoding:")
		f.AddWindowTextBox("encode", lbl+" ", "", "string", 2, 4, 15, 15, nil, "")
		lbl = Language.Translate("Cancel")
		f.AddWindowButton("cancel", " "+lbl+" ", "cancel", 33-Count(lbl), 6, m.ButtonFinish, "")
		lbl = Language.Translate("Set encoding")
		f.AddWindowButton("set", " "+lbl+" ", "ok", 43, 6, m.SetEncoding, "")
		m.myapp.WindowFinish = callback
	} else {
		f = m.myapp.frames["f"]
	}
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
