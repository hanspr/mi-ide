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
		"Cd":        Cd,
		"Help":      Help,
		"MemUsage":  MemUsage,
		"Open":      Open,
		"Pwd":       Pwd,
		"Raw":       Raw,
		"Reload":    Reload,
		"SaveAs":    SaveAs,
		"ToggleLog": ToggleLog,
		"GroupEdit": GroupEdit,
		"GroupGit":  GroupGit,
		"GroupMenu": GroupMenu,
		"GroupShow": GroupShow,
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
		"edit:": {"GroupEdit", []Completion{GroupCompletion, NoCompletion}},
		"git:":  {"GroupGit", []Completion{GroupCompletion, NoCompletion}},
		"menu:": {"GroupMenu", []Completion{GroupCompletion, NoCompletion}},
		"show:": {"GroupShow", []Completion{GroupCompletion, NoCompletion}},
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

// Groups

func GroupEdit(args []string) {
	if args[0] == "settings" {
		EditSettings()
	} else if args[0] == "snippets" {
		EditSnippets()
	}
}

func GroupMenu(args []string) {
	if args[0] == "cloudsettings" {
		CloudSettings()
	} else if args[0] == "keybindings" {
		KeyBindings()
	} else if args[0] == "plugins" {
		PluginManager()
	} else if args[0] == "settings" {
		Settings()
	}
}

func GroupShow(args []string) {
	if args[0] == "snippets" {
		ShowSnippets()
	}
}

func GroupGit(args []string) {
	messenger.AddLog("GroupGit:", args[0], ":")
	if args[0] == "diff" {
		GitDiff("")
	} else if args[0] == "diffstaged" {
		GitDiff("--staged")
	} else if args[0] == "status" {
		GitStatus()
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
		WorkingDir = wd
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

// : Snippets

// Show loaded snippets
func ShowSnippets() {
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

func EditSettings() {
	CurView().AddTab(false)
	CurView().Open(configDir + "/settings.json")
}

func EditSnippets() {
	ftype := CurView().Buf.FileType()
	CurView().AddTab(false)
	CurView().Open(configDir + "/settings/snippets/" + ftype + ".snippets")
}

// : Git

// Load git status into a window
func GitStatus() {
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
func GitDiff(arg string) {
	messenger.AddLog("Execute Git ", arg)
	if !git.enabled {
		return
	}
	cmd := "git diff"
	if arg != "" {
		cmd = cmd + " " + arg
	}
	cmd = cmd + " -w"
	diff, err := RunShellCommand(cmd)
	if err != nil {
		messenger.AddLog(err.Error())
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

func Settings() {
	micromenu.GlobalConfigDialog()
}

func CloudSettings() {
	micromenu.MiCloudServices()
}

func PluginManager() {
	micromenu.PluginManagerDialog()
}

func KeyBindings() {
	micromenu.KeyBindingsDialog()
}
