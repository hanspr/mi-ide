package main

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/flynn/json5"
	"github.com/hanspr/tcell"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

var loadedPlugins map[string]string

// Call calls the lua function 'function'
// If it does not exist nothing happens, if there is an error,
// the error is returned
func Call(function string, args ...any) (lua.LValue, error) {
	var luaFunc lua.LValue
	if strings.Contains(function, ".") {
		plugin := L.GetGlobal(strings.Split(function, ".")[0])
		if plugin.String() == "nil" {
			return nil, errors.New("function does not exist: " + function)
		}
		luaFunc = L.GetField(plugin, strings.Split(function, ".")[1])
	} else {
		luaFunc = L.GetGlobal(function)
	}

	if luaFunc.String() == "nil" {
		return nil, errors.New("function does not exist: " + function)
	}
	var luaArgs []lua.LValue
	for _, v := range args {
		luaArgs = append(luaArgs, luar.New(L, v))
	}
	err := L.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    1,
		Protect: true,
	}, luaArgs...)
	ret := L.Get(-1) // returned value
	if ret.String() != "nil" {
		L.Pop(1) // remove received value
	}
	return ret, err
}

// CallP function call to plugins
func CallP(luaFunc lua.LValue, args ...any) (lua.LValue, error) {
	var luaArgs []lua.LValue
	for _, v := range args {
		luaArgs = append(luaArgs, luar.New(L, v))
	}
	err := L.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    1,
		Protect: true,
	}, luaArgs...)
	ret := L.Get(-1) // returned value
	if ret.String() != "nil" {
		L.Pop(1) // remove received value
	}
	return ret, err
}

// LuaFunctionBinding is a function generator which takes the name of a lua function
// and creates a function that will call that lua function
// Specifically it creates a function that can be called as a binding because this is used
// to bind keys to lua functions
func LuaFunctionBinding(function string) func(*View, bool) bool {
	return func(v *View, _ bool) bool {
		_, err := Call(function, nil)
		if err != nil {
			TermMessage(err)
		}
		return false
	}
}

// LuaFunctionMouseBinding is a function generator which takes the name of a lua function
// and creates a function that will call that lua function
// Specifically it creates a function that can be called as a mouse binding because this is used
// to bind mouse actions to lua functions
func LuaFunctionMouseBinding(function string) func(*View, bool, *tcell.EventMouse) bool {
	return func(v *View, _ bool, e *tcell.EventMouse) bool {
		_, err := Call(function, e)
		if err != nil {
			TermMessage(err)
		}
		return false
	}
}

func unpack(old []string) []any {
	new := make([]any, len(old))
	for i, v := range old {
		new[i] = v
	}
	return new
}

// LuaFunctionCommand is the same as LuaFunctionBinding except it returns a normal function
// so that a command can be bound to a lua function
func LuaFunctionCommand(function string) func([]string) {
	return func(args []string) {
		_, err := Call(function, unpack(args)...)
		if err != nil {
			TermMessage(err)
		}
	}
}

// LuaFunctionComplete returns a function which can be used for autocomplete in plugins
func LuaFunctionComplete(function string) func(string) []string {
	return func(input string) (result []string) {

		res, err := Call(function, input)
		if err != nil {
			TermMessage(err)
		}
		if tbl, ok := res.(*lua.LTable); !ok {
			TermMessage(function, "should return a table of strings")
		} else {
			for i := 1; i <= tbl.Len(); i++ {
				val := tbl.RawGetInt(i)
				if v, ok := val.(lua.LString); !ok {
					TermMessage(function, "should return a table of strings")
				} else {
					result = append(result, string(v))
				}
			}
		}
		return result
	}
}

// LuaFunctionJob returns a function that will call the given lua function
// structured as a job call i.e. the job output and arguments are provided
// to the lua function
func LuaFunctionJob(function string) func(string, ...string) {
	return func(output string, args ...string) {
		_, err := Call(function, unpack(append([]string{output}, args...))...)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
		}
	}
}

// luaPluginName convert a human-friendly plugin name into a valid lua variable name.
func luaPluginName(name string) string {
	return strings.Replace(name, "-", "_", -1)
}

// LoadPlugins loads the plugins located in ~/.config/mi-ide/plugins
func LoadPlugins() {
	loadedPlugins = make(map[string]string)
	pluginOption = make(map[string]any)

	for _, plugin := range ListRuntimeFiles(RTPlugin) {
		pluginName := plugin.Name()
		if _, ok := loadedPlugins[pluginName]; ok {
			continue
		}

		data, err := plugin.Data()
		if err != nil {
			TermMessage("Error loading plugin: " + pluginName)
			continue
		}

		pluginLuaName := luaPluginName(pluginName)

		LoadPluginOptions(pluginName)

		if err := LoadFile(pluginLuaName, pluginLuaName, string(data)); err != nil {
			TermMessage(err)
			continue
		}

		loadedPlugins[pluginName] = pluginLuaName

	}

	if _, err := os.Stat(configDir + "/init.lua"); err == nil {
		data, _ := os.ReadFile(configDir + "/init.lua")
		if err := LoadFile("init", configDir+"/init.lua", string(data)); err != nil {
			TermMessage(err)
		}
		loadedPlugins["init"] = "init"
	}
}

// GlobalPluginCall makes a call to a function in every plugin that is currently
// loaded
func GlobalPluginCall(function string, args ...any) {
	for pl := range loadedPlugins {
		if GetPluginOption(pl, "ftype") != "*" && (GetPluginOption(pl, "ftype") == nil || GetPluginOption(pl, "ftype").(string) != CurView().Buf.FileType()) {
			continue
		}
		_, err := Call(pl+"."+function, args...)
		if err != nil && !strings.HasPrefix(err.Error(), "function does not exist") {
			TermMessage(err)
			continue
		}
	}
}

// Plugin Options

// Each plugin has its own plugin settings file, these are the functions to add, get or set options

var pluginOption map[string]any

// LoadPluginOptions load options for the plugin
func LoadPluginOptions(pname string) {
	var parsed map[string]any

	filename := configDir + "/plugins/" + pname + "/settings.json"
	if _, err := os.Stat(filename); err == nil {
		input, err := os.ReadFile(filename)
		if err != nil {
			return
		}
		err = json5.Unmarshal(input, &parsed)
		if err != nil {
			return
		}
		for k, v := range parsed {
			if !strings.HasPrefix(reflect.TypeOf(v).String(), "map") {
				pluginOption[pname+"-"+k] = v
			}
		}
	}
}

// WritePluginSettings write the plugin settings to file
func WritePluginSettings(pname string) error {
	if pname == "" {
		return errors.New("missing plugin name, can not write settings")
	}
	if _, e := os.Stat(configDir + "/plugins/" + pname); e == nil {
		filename := configDir + "/plugins/" + pname + "/settings.json"
		parsed := make(map[string]any)
		for k, v := range pluginOption {
			if strings.Contains(k, pname+"-") {
				kp := strings.Split(k, "-")
				parsed[kp[1]] = v
			}
		}
		txt, _ := json.MarshalIndent(parsed, "", "    ")
		err := os.WriteFile(filename, append(txt, '\n'), 0644)
		return err
	}
	return nil
}

// AddPluginOption add an option to the file
func AddPluginOption(pname, option string, value any) error {
	opt := pname + "-" + option
	_, ok := pluginOption[opt]
	if ok {
		return nil
	}
	pluginOption[opt] = value
	return WritePluginSettings(pname)
}

// SetPluginOption set a plugin option
func SetPluginOption(pname, option string, value any) {
	opt := pname + "-" + option
	pluginOption[opt] = value
}

// GetPluginOption get value of an option
func GetPluginOption(pname, option string) any {
	value, ok := pluginOption[pname+"-"+option]
	if !ok {
		return nil
	}
	return value
}

// PluginGetApp get access to mircoapp application to plugins
func PluginGetApp() *MicroApp {
	if apprunning != nil {
		return nil
	}
	app := new(MicroApp)
	return app
}

// PluginRunApp run plugin app
func PluginRunApp(app *MicroApp) bool {
	if apprunning == nil {
		apprunning = app
		app.Start()
		return true
	}
	return false
}

// PluginStopApp stop plugin app
func PluginStopApp(app *MicroApp) {
	if apprunning == nil {
		return
	}
	if apprunning.name == app.name {
		MicroAppStop()
	}
}

// PluginAddIcon add icon to toolbar
func PluginAddIcon(icon, callback string) {
	MicroToolBar.AddIcon([]rune(icon)[0], nil, callback)
}
