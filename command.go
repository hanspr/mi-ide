package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	humanize "github.com/dustin/go-humanize"
	"github.com/hanspr/shellwords"
)

// A Command contains a action (a function to call) as well as information about how to autocomplete the command
type Command struct {
	action      func([]string)
	completions []Completion
}

// A StrCommand is similar to a command but keeps the name of the action
type StrCommand struct {
	action      string
	completions []Completion
}

var (
	commands       map[string]Command
	commandActions map[string]func([]string)
	replacing      bool = false
)

func init() {
	commandActions = map[string]func([]string){
		"Bind":          Bind,
		"Cd":            Cd,
		"CloudSettings": CloudSettings,
		"EditSettings":  EditSettings,
		"EditSnippets":  EditSnippets,
		"GitDiff":       GitDiff,
		"GitStatus":     GitStatus,
		"Help":          Help,
		"HSplit":        HSplit,
		"KeyBindings":   KeyBindings,
		"MemUsage":      MemUsage,
		"Open":          Open,
		"PluginManager": PluginManager,
		"Pwd":           Pwd,
		"Raw":           Raw,
		"Reload":        Reload,
		"Set":           Set,
		"SetLocal":      SetLocal,
		"Show":          Show,
		"ShowKey":       ShowKey,
		"Save":          Save,
		"SaveAs":        SaveAs,
		"Settings":      Settings,
		"ShowSnippets":  ShowSnippets,
		"ToggleLog":     ToggleLog,
		"VSplit":        VSplit,
	}
}

// group commands with group:command
// DefaultCommands returns a map containing mi-ide's default commands
func DefaultCommands() map[string]StrCommand {
	return map[string]StrCommand{
		"cd":       {"Cd", []Completion{FileCompletion}},
		"help":     {"Help", []Completion{HelpCompletion, NoCompletion}},
		"log":      {"ToggleLog", []Completion{NoCompletion}},
		"memusage": {"MemUsage", []Completion{NoCompletion}},
		"open":     {"Open", []Completion{FileCompletion}},
		"pwd":      {"Pwd", []Completion{NoCompletion}},
		"raw":      {"Raw", []Completion{NoCompletion}},
		"reload":   {"Reload", []Completion{NoCompletion}},
		"saveas":   {"SaveAs", []Completion{NoCompletion}},
		// Groups
		"edit:settings":      {"EditSettings", []Completion{NoCompletion}},
		"edit:snippets":      {"EditSnippets", []Completion{NoCompletion}},
		"git:status":         {"GitStatus", []Completion{NoCompletion}},
		"git:diff":           {"GitDiff", []Completion{NoCompletion}},
		"keys:bind":          {"Bind", []Completion{NoCompletion}},
		"keys:showkey":       {"ShowKey", []Completion{NoCompletion}},
		"menu:settings":      {"Settings", []Completion{NoCompletion}},
		"menu:cloudsettings": {"CloudSettings", []Completion{NoCompletion}},
		"menu:plugins":       {"PluginManager", []Completion{NoCompletion}},
		"menu:keys":          {"KeyBindings", []Completion{NoCompletion}},
		"option:set":         {"Set", []Completion{OptionCompletion, OptionValueCompletion}},
		"option:setlocal":    {"SetLocal", []Completion{OptionCompletion, OptionValueCompletion}},
		"split:vertical":     {"VSplit", []Completion{FileCompletion, NoCompletion}},
		"split:horizontal":   {"HSplit", []Completion{FileCompletion, NoCompletion}},
		"show:option":        {"Show", []Completion{OptionCompletion, NoCompletion}},
		"show:snippets":      {"ShowSnippets", []Completion{OptionCompletion, NoCompletion}},
	}
}

// InitCommands initializes the default commands
func InitCommands() {
	commands = make(map[string]Command)

	defaults := DefaultCommands()
	parseCommands(defaults)
}

func parseCommands(userCommands map[string]StrCommand) {
	for k, v := range userCommands {
		MakeCommand(k, v.action, v.completions...)
	}
}

// MakeCommand is a function to easily create new commands
// This can be called by plugins in Lua so that plugins can define their own commands
func MakeCommand(name, function string, completions ...Completion) {
	action := commandActions[function]
	if _, ok := commandActions[function]; !ok {
		// If the user seems to be binding a function that doesn't exist
		// We hope that it's a lua function that exists and bind it to that
		action = LuaFunctionCommand(function)
	}

	commands[name] = Command{action, completions}
}

// CommandEditAction returns a bindable function that opens a prompt with
// the given string and executes the command when the user presses
// enter
func CommandEditAction(prompt string) func(*View, bool) bool {
	return func(v *View, usePlugin bool) bool {
		input, canceled := messenger.Prompt("> ", prompt, "Command", CommandCompletion)
		if !canceled {
			HandleCommand(input)
		}
		return false
	}
}

// CommandAction returns a bindable function which executes the
// given command
func CommandAction(cmd string) func(*View, bool) bool {
	return func(v *View, usePlugin bool) bool {
		HandleCommand(cmd)
		return false
	}
}

// SaveAs saves the buffer with a new name
func SaveAs(args []string) {
	if len(args) > 0 {
		CurView().saveToFile(args[0])
	}
}

// Raw opens a new raw view which displays the escape sequences mi-ide
// is receiving in real-time
func Raw(args []string) {
	buf := NewBufferFromString("", "Raw events")

	view := NewView(buf)
	view.Buf.Insert(view.Cursor.Loc, Language.Translate("Warning: Showing raw event escape codes")+"\n")
	view.Buf.Insert(view.Cursor.Loc, Language.Translate("Use CtrlQ to exit")+"\n")
	view.Type = vtRaw
	tab := NewTabFromView(view)
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
}

// Cd changes the current working directory
func Cd(args []string) {
	if len(args) > 0 {
		path := ReplaceHome(args[0])
		err := os.Chdir(path)
		if err != nil {
			messenger.Alert("error", Language.Translate("Error with cd:"), " ", err)
			return
		}
		wd, _ := os.Getwd()
		for _, tab := range tabs {
			for _, view := range tab.Views {
				if len(view.Buf.name) == 0 {
					continue
				}

				view.Buf.Path, _ = MakeRelative(view.Buf.AbsPath, wd)
				if p, _ := filepath.Abs(view.Buf.Path); !strings.Contains(p, wd) {
					view.Buf.Path = view.Buf.AbsPath
				}
			}
		}
		git.CheckGit()
	}
}

// MemUsage prints mi-ide's memory usage
// Alloc shows how many bytes are currently in use
// Sys shows how many bytes have been requested from the operating system
// NumGC shows how many times the GC has been run
// Note that Go commonly reserves more memory from the OS than is currently in-use/required
// Additionally, even if Go returns memory to the OS, the OS does not always claim it because
// there may be plenty of memory to spare
func MemUsage(args []string) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	messenger.Message(fmt.Sprintf("Alloc: %v, Sys: %v, NumGC: %v", humanize.Bytes(mem.Alloc), humanize.Bytes(mem.Sys), mem.NumGC))
}

// Pwd prints the current working directory
func Pwd(args []string) {
	wd, err := os.Getwd()
	if err != nil {
		messenger.Message(err.Error())
	} else {
		messenger.Message(wd)
	}
}

// Open opens a new buffer with a given filename
func Open(args []string) {
	if len(args) > 0 {
		filename := args[0]
		// the filename might or might not be quoted, so unquote first then join the strings.
		args, err := shellwords.Split(filename)
		if err != nil {
			messenger.Alert("error", Language.Translate("Error parsing args"), " ", err)
			return
		}
		filename = strings.Join(args, " ")

		CurView().Open(filename)
	} else {
		messenger.Alert("error", Language.Translate("No filename"))
	}
}

// Save saves the buffer in the main view
func Save(args []string) {
	// Save the main view
	CurView().Save(true)
}

// ToggleLog toggles the log view
func ToggleLog(args []string) {
	buffer := messenger.getBuffer()
	if CurView().Type != vtLog {
		CurView().HSplit(buffer)
		CurView().Type = vtLog
		RedrawAll(true)
		buffer.Cursor.Loc = buffer.Start()
		CurView().Relocate()
		buffer.Cursor.Loc = buffer.End()
		CurView().Relocate()
		CurView().PreviousSplit(false)
	} else {
		CurView().Quit(true)
	}
}

// Reload reloads all files (syntax files, colorschemes...)
func Reload(args []string) {
	loadAll()
}

// Help tries to open the given help page in a horizontal split
func Help(args []string) {
	if len(args) < 1 {
		// Open the default help if the user just typed "> help"
		CurView().openHelp("help")
	} else {
		helpPage := args[0]
		if FindRuntimeFile(RTHelp, helpPage) != nil {
			CurView().openHelp(helpPage)
		} else {
			messenger.Alert("error", Language.Translate("Sorry, no help for"), " ", helpPage)
		}
	}
}

// VSplit opens a vertical split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func VSplit(args []string) {
	if len(args) == 0 {
		CurView().VSplit(NewBufferFromString("", ""))
	} else {
		messenger.Message(args[0])
		buf, err := NewBufferFromFile(args[0])
		if err != nil {
			messenger.Alert("error", err)
			return
		}
		CurView().VSplit(buf)
	}
}

// HSplit opens a horizontal split with file given in the first argument
// If no file is given, it opens an empty buffer in a new split
func HSplit(args []string) {
	if len(args) == 0 {
		CurView().HSplit(NewBufferFromString("", ""))
	} else {
		messenger.Message(args[0])
		buf, err := NewBufferFromFile(args[0])
		if err != nil {
			messenger.Alert("error", err)
			return
		}
		CurView().HSplit(buf)
	}
}

// Set sets an option
func Set(args []string) {
	if len(args) < 2 {
		messenger.Alert("error", Language.Translate("Not enough arguments"))
		return
	}

	option := args[0]
	value := args[1]

	SetOptionAndSettings(option, value)
}

// SetLocal sets an option local to the buffer
func SetLocal(args []string) {
	if len(args) < 2 {
		messenger.Alert("error", Language.Translate("Not enough arguments"))
		return
	}

	option := args[0]
	value := args[1]

	err := SetLocalOption(option, value, CurView())
	if err != nil {
		messenger.Alert("error", err.Error())
	}
}

// Show shows the value of the given option
func Show(args []string) {
	if len(args) < 1 {
		messenger.Alert("error", Language.Translate("Please provide an option to show"))
		return
	}

	option := GetOption(args[0])

	if option == nil {
		messenger.Alert("error", args[0], " "+Language.Translate("is not a valid option"))
		return
	}

	messenger.Message(option)
}

// ShowKey displays the action that a key is bound to
func ShowKey(args []string) {
	if len(args) < 1 {
		messenger.Alert("error", Language.Translate("Please provide a key to show"))
		return
	}

	if action, ok := bindingsStr[args[0]]; ok {
		messenger.Message(action)
	} else {
		messenger.Message(args[0], " "+Language.Translate("has no binding"))
	}
}

// Bind creates a new keybinding
func Bind(args []string) {
	if len(args) < 2 {
		messenger.Alert("error", Language.Translate("Not enough arguments"))
		return
	}
	BindKey(args[0], args[1])
}

// : Snippets

// Show loaded snippets
func ShowSnippets(args []string) {
	ftype := CurView().Buf.FileType()
	snips := ""
	loadSnippets(ftype)
	keys := make([]string, 0, len(snippets))
	for k := range snippets {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		snips = snips + "snippet " + name + "\n" + snippets[name].code + "\n\n"
	}
	CurView().OpenHelperView("v", "git-status", snips)
}

func EditSettings(args []string) {
	CurView().AddTab(false)
	CurView().Open(configDir + "/settings.json")
}

func EditSnippets(args []string) {
	ftype := CurView().Buf.FileType()
	CurView().AddTab(false)
	CurView().Open(configDir + "/settings/snippets/" + ftype + ".snippets")
}

// : Git

// Load git status into a window
func GitStatus(args []string) {
	if !git.enabled {
		git.CheckGit()
		return
	}
	status, err := RunShellCommand("git status")
	if err != nil {
		return
	}
	CurView().OpenHelperView("h", "git-status", status)
}

// Load git status into a window
func GitDiff(args []string) {
	if !git.enabled {
		return
	}
	diff, err := RunShellCommand("git diff -w")
	if err != nil {
		return
	}
	CurView().AddTab(false)
	CurView().Buf = NewBufferFromString(diff, "")
	CurView().Buf.Settings["filetype"] = "git-diff"
	CurView().Type = vtLog
	CurView().Buf.UpdateRules()
	CurView().Buf.fname = "git-diff"
	SetLocalOption("softwrap", "true", CurView())
	SetLocalOption("ruler", "false", CurView())
	NavigationMode = true
}

// : Helper functions

// HandleCommand handles input from the user
func HandleCommand(input string) {
	args, err := shellwords.Split(input)
	if err != nil {
		messenger.Alert("error", Language.Translate("Error parsing args"), " ", err)
		return
	}

	inputCmd := args[0]

	if _, ok := commands[inputCmd]; !ok {
		messenger.Alert("error", Language.Translate("Unknown command"), " ", inputCmd)
	} else {
		commands[inputCmd].action(args[1:])
	}
}

// ExpandString transform string secuence to its correct string value
func ExpandString(s string) string {
	s = strings.Replace(s, `\\`, `{ESCBS}`, -1)

	s = strings.Replace(s, `\t`, "\t", -1)
	s = strings.Replace(s, `\n`, "\n", -1)
	s = strings.Replace(s, `\r`, "\r", -1)

	s = strings.Replace(s, `{ESCBS}`, `\`, -1)

	return s
}

// Menu access

func Settings(args []string) {
	micromenu.GlobalConfigDialog()
}

func CloudSettings(args []string) {
	micromenu.MiCloudServices()
}

func PluginManager(args []string) {
	micromenu.PluginManagerDialog()
}

func KeyBindings(args []string) {
	micromenu.KeyBindingsDialog()
}
