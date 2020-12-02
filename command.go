package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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
		"Set":        Set,
		"SetLocal":   SetLocal,
		"Show":       Show,
		"ShowKey":    ShowKey,
		"Run":        Run,
		"Bind":       Bind,
		"Quit":       Quit,
		"Save":       Save,
		"Replace":    Replace,
		"ReplaceAll": ReplaceAll,
		"VSplit":     VSplit,
		"HSplit":     HSplit,
		"Tab":        NewTab,
		"Help":       Help,
		"Eval":       Eval,
		"ToggleLog":  ToggleLog,
		"Plugin":     PluginCmd,
		"Reload":     Reload,
		"Cd":         Cd,
		"Pwd":        Pwd,
		"Open":       Open,
		"TabSwitch":  TabSwitch,
		"Term":       Term,
		"MemUsage":   MemUsage,
		"Retab":      Retab,
		"Raw":        Raw,
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

// DefaultCommands returns a map containing micro's default commands
func DefaultCommands() map[string]StrCommand {
	return map[string]StrCommand{
		"set":        {"Set", []Completion{OptionCompletion, OptionValueCompletion}},
		"setlocal":   {"SetLocal", []Completion{OptionCompletion, OptionValueCompletion}},
		"show":       {"Show", []Completion{OptionCompletion, NoCompletion}},
		"showkey":    {"ShowKey", []Completion{NoCompletion}},
		"bind":       {"Bind", []Completion{NoCompletion}},
		"run":        {"Run", []Completion{NoCompletion}},
		"quit":       {"Quit", []Completion{NoCompletion}},
		"save":       {"Save", []Completion{NoCompletion}},
		"replace":    {"Replace", []Completion{NoCompletion}},
		"replaceall": {"ReplaceAll", []Completion{NoCompletion}},
		"vsplit":     {"VSplit", []Completion{FileCompletion, NoCompletion}},
		"hsplit":     {"HSplit", []Completion{FileCompletion, NoCompletion}},
		"tab":        {"Tab", []Completion{FileCompletion, NoCompletion}},
		"help":       {"Help", []Completion{HelpCompletion, NoCompletion}},
		"eval":       {"Eval", []Completion{NoCompletion}},
		"log":        {"ToggleLog", []Completion{NoCompletion}},
		"plugin":     {"Plugin", []Completion{PluginCmdCompletion, PluginNameCompletion}},
		"reload":     {"Reload", []Completion{NoCompletion}},
		"cd":         {"Cd", []Completion{FileCompletion}},
		"pwd":        {"Pwd", []Completion{NoCompletion}},
		"open":       {"Open", []Completion{FileCompletion}},
		"tabswitch":  {"TabSwitch", []Completion{NoCompletion}},
		"term":       {"Term", []Completion{NoCompletion}},
		"memusage":   {"MemUsage", []Completion{NoCompletion}},
		"retab":      {"Retab", []Completion{NoCompletion}},
		"raw":        {"Raw", []Completion{NoCompletion}},
	}
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

// PluginCmd installs, removes, updates, lists, or searches for given plugins
func PluginCmd(args []string) {
	if len(args) >= 1 {
		switch args[0] {
		case "install":
			installedVersions := GetInstalledVersions(false)
			for _, plugin := range args[1:] {
				pp := GetAllPluginPackages().Get(plugin)
				if pp == nil {
					messenger.Error(Language.Translate("Unknown plugin:") + plugin)
				} else if err := pp.IsInstallable(); err != nil {
					messenger.Error(Language.Translate("Error installing"), " ", plugin, ": ", err)
				} else {
					for _, installed := range installedVersions {
						if pp.Name == installed.pack.Name {
							if pp.Versions[0].Version.Compare(installed.Version) == 1 {
								messenger.Error(pp.Name, " "+Language.Translate("is already installed but out-of-date: use 'plugin update"), " ", pp.Name, Language.Translate("' to update"))
							} else {
								messenger.Error(pp.Name, " "+Language.Translate("is already installed"))
							}
						}
					}
					pp.Install()
				}
			}
		case "remove":
			removed := ""
			for _, plugin := range args[1:] {
				// check if the plugin exists.
				if _, ok := loadedPlugins[plugin]; ok {
					UninstallPlugin(plugin)
					removed += plugin + " "
					continue
				}
			}
			if !IsSpaces([]byte(removed)) {
				messenger.Message(Language.Translate("Removed"), " ", removed)
			} else {
				messenger.Error(Language.Translate("The requested plugins do not exist"))
			}
		case "update":
			UpdatePlugins(args[1:])
		case "list":
			plugins := GetInstalledVersions(false)
			messenger.AddLog("----------------")
			messenger.AddLog(Language.Translate("The following plugins are currently installed:") + "\n")
			for _, p := range plugins {
				messenger.AddLog(fmt.Sprintf("%s (%s)", p.pack.Name, p.Version))
			}
			messenger.AddLog("----------------")
			if len(plugins) > 0 {
				if CurView().Type != vtLog {
					ToggleLog([]string{})
				}
			}
		case "search":
			plugins := SearchPlugin(args[1:])
			messenger.Message(len(plugins), " "+Language.Translate("plugins found"))
			for _, p := range plugins {
				messenger.AddLog("----------------")
				messenger.AddLog(p.String())
			}
			messenger.AddLog("----------------")
			if len(plugins) > 0 {
				if CurView().Type != vtLog {
					ToggleLog([]string{})
				}
			}
		case "available":
			packages := GetAllPluginPackages()
			messenger.AddLog(Language.Translate("Available Plugins:"))
			for _, pkg := range packages {
				messenger.AddLog(pkg.Name)
			}
			if CurView().Type != vtLog {
				ToggleLog([]string{})
			}
		}
	} else {
		messenger.Error(Language.Translate("Not enough arguments"))
	}
}

// Retab changes all spaces to tabs or all tabs to spaces
// depending on the user's settings
func Retab(args []string) {
	CurView().Retab(true)
}

// Raw opens a new raw view which displays the escape sequences micro
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

// TabSwitch switches to a given tab either by name or by number
func TabSwitch(args []string) {
	if len(args) > 0 {
		num, err := strconv.Atoi(args[0])
		if err != nil {
			// Check for tab with this name

			found := false
			for _, t := range tabs {
				v := t.Views[t.CurView]
				if v.Buf.GetName() == args[0] {
					curTab = v.TabNum
					found = true
				}
			}
			if !found {
				messenger.Error(Language.Translate("Could not find tab:"), " ", err)
			}
		} else {
			num--
			if num >= 0 && num < len(tabs) {
				curTab = num
			} else {
				messenger.Error(Language.Translate("Invalid tab index"))
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
			messenger.Error(Language.Translate("Error with cd:"), " ", err)
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
	}
}

// MemUsage prints micro's memory usage
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
			messenger.Error(Language.Translate("Error parsing args"), " ", err)
			return
		}
		filename = strings.Join(args, " ")

		CurView().Open(filename)
	} else {
		messenger.Error(Language.Translate("No filename"))
	}
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
			messenger.Error(Language.Translate("Sorry, no help for"), " ", helpPage)
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
			messenger.Error(err)
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
			messenger.Error(err)
			return
		}
		CurView().HSplit(buf)
	}
}

// Eval evaluates a lua expression
func Eval(args []string) {
	if len(args) >= 1 {
		err := L.DoString(args[0])
		if err != nil {
			messenger.Error(err)
		}
	} else {
		messenger.Error(Language.Translate("Not enough arguments"))
	}
}

// NewTab opens the given file in a new tab
func NewTab(args []string) {
	if len(args) == 0 {
		CurView().AddTab(true)
	} else {
		buf, err := NewBufferFromFile(args[0])
		if err != nil {
			messenger.Error(err)
			return
		}

		tab := NewTabFromView(NewView(buf))
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
}

// Set sets an option
func Set(args []string) {
	if len(args) < 2 {
		messenger.Error(Language.Translate("Not enough arguments"))
		return
	}

	option := args[0]
	value := args[1]

	SetOptionAndSettings(option, value)
}

// SetLocal sets an option local to the buffer
func SetLocal(args []string) {
	if len(args) < 2 {
		messenger.Error(Language.Translate("Not enough arguments"))
		return
	}

	option := args[0]
	value := args[1]

	err := SetLocalOption(option, value, CurView())
	if err != nil {
		messenger.Error(err.Error())
	}
}

// Show shows the value of the given option
func Show(args []string) {
	if len(args) < 1 {
		messenger.Error(Language.Translate("Please provide an option to show"))
		return
	}

	option := GetOption(args[0])

	if option == nil {
		messenger.Error(args[0], " "+Language.Translate("is not a valid option"))
		return
	}

	messenger.Message(option)
}

// ShowKey displays the action that a key is bound to
func ShowKey(args []string) {
	if len(args) < 1 {
		messenger.Error(Language.Translate("Please provide a key to show"))
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
		messenger.Error(Language.Translate("Not enough arguments"))
		return
	}
	BindKey(args[0], args[1])
}

// Run runs a shell command in the background
func Run(args []string) {
	// Run a shell command in the background (openTerm is false)
	HandleShellCommand(shellwords.Join(args...), false, true)
}

// Quit closes the main view
func Quit(args []string) {
	// Close the main view
	CurView().Quit(true)
}

// Save saves the buffer in the main view
func Save(args []string) {
	if len(args) == 0 {
		// Save the main view
		CurView().Save(true)
	} else {
		CurView().Buf.SaveAs(args[0])
	}
}

// Replace runs search and replace
func Replace(args []string) {

	if len(args) < 2 || args[0] == args[1] {
		// We need to find both a search and replace expression
		messenger.Error(Language.Translate("Invalid replace statement:"), " "+strings.Join(args, " "))
		return
	}

	all := false
	noRegex := false
	mods := "(?m)"
	replacing = true
	if len(args) > 2 {
		for _, arg := range args[2:] {
			switch arg {
			case "a":
				all = true
			case "i":
				mods = mods + "(?i)"
			case "s":
				mods = mods + "(?s)"
			case "l":
				noRegex = true
			case "":
			default:
				messenger.Error(Language.Translate("Invalid flag:"), " ", arg)
				return
			}
		}
	}

	search := string(args[0])

	if noRegex {
		search = regexp.QuoteMeta(search)
	} else {
		search = mods + search
	}

	replace := string(args[1])
	replace = ExpandString(replace)
	regex, err := regexp.Compile(search)
	if err != nil {
		// There was an error with the user's regex
		messenger.Error(err.Error())
		replacing = false
		return
	}

	view := CurView()
	startLine := view.searchSave.Y
	found := 0
	view.searchLoops = 0

	for {
		var choice rune
		var canceled bool
		// The 'check' flag was used
		Search(search, view, true)
		if !view.Cursor.HasSelection() {
			replacing = false
			break
		}
		view.Relocate()
		RedrawAll(true)
		if view.searchLoops > 0 && view.Cursor.Y > startLine {
			view.searchLoops = 0
			if view.Cursor.HasSelection() {
				view.Cursor.Loc = view.Cursor.CurSelection[0]
				view.Cursor.ResetSelection()
			}
			messenger.Reset()
			replacing = false
			break
		}
		y := []rune(Language.Translate("y"))[0]
		n := []rune(Language.Translate("n"))[0]
		q := []rune(Language.Translate("q"))[0]
		I := []rune(Language.Translate("!"))[0]
		if all {
			freeze = true
		} else {
			choice, canceled = messenger.LetterPrompt(Language.Translate("Perform replacement? (y,n,q,!)"), y, n, q, I)
		}
		if canceled {
			if view.Cursor.HasSelection() {
				view.Cursor.Loc = view.Cursor.CurSelection[0]
				view.Cursor.ResetSelection()
			}
			messenger.Reset()
			replacing = false
			break
		} else if choice == y || choice == I || all {
			if choice == I && !all {
				all = true
			}
			sel := view.Cursor.GetSelection()
			rep := regex.ReplaceAllString(sel, replace)
			if Count(rep) > Count(sel) {
				searchStart = Loc{view.Cursor.CurSelection[1].X + 1, view.Cursor.Loc.Y}
			} else {
				searchStart = Loc{view.Cursor.CurSelection[1].X - 1, view.Cursor.Loc.Y}
			}
			if searchStart.X > len(view.Buf.LineBytes(searchStart.Y))-1 {
				searchStart = Loc{0, searchStart.Y + 1}
				if searchStart.Y > view.Buf.Len()-1 {
					searchStart.Y = view.Buf.Len() - 1
				}
			}
			view.Cursor.DeleteSelection()
			view.Buf.Insert(view.Cursor.Loc, rep)
			view.Cursor.ResetSelection()
			messenger.Reset()
			found++
		} else if choice == q {
			if view.Cursor.HasSelection() {
				view.Cursor.Loc = view.Cursor.CurSelection[0]
				view.Cursor.ResetSelection()
			}
			messenger.Reset()
			replacing = false
			break
		} else {
			searchStart = view.Cursor.CurSelection[1]
		}
	}
	freeze = false
	view.Cursor.Relocate()

	if found > 1 {
		messenger.Message(Language.Translate("Replaced"), " ", found, " "+Language.Translate("occurrences of"), " ", search)
	} else if found == 1 {
		messenger.Message(Language.Translate("Replaced"), " ", found, " "+Language.Translate("occurrence of "), search)
	} else {
		messenger.Message(Language.Translate("Nothing matched"), " ", search)
	}
}

// ReplaceAll replaces search term all at once
func ReplaceAll(args []string) {
	// aliased to Replace command
	Replace(append(args, "-a"))
}

// Term opens a terminal in the current view
func Term(args []string) {
	var err error
	if len(args) == 0 {
		err = CurView().StartTerminal([]string{os.Getenv("SHELL"), "-i"}, true, false, "")
	} else {
		err = CurView().StartTerminal(args, true, false, "")
	}
	if err != nil {
		messenger.Error(err)
	}
}

// HandleCommand handles input from the user
func HandleCommand(input string) {
	args, err := shellwords.Split(input)
	if err != nil {
		messenger.Error(Language.Translate("Error parsing args"), " ", err)
		return
	}

	inputCmd := args[0]

	if _, ok := commands[inputCmd]; !ok {
		messenger.Error(Language.Translate("Unknown command"), " ", inputCmd)
	} else {
		commands[inputCmd].action(args[1:])
	}
}

// ExpandString transform string secuence to its correct string value
func ExpandString(s string) string {
	s = strings.Replace(s, `\t`, "\t", -1)
	s = strings.Replace(s, `\n`, "\n", -1)
	s = strings.Replace(s, `\r`, "\r", -1)
	return s
}
