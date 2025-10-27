package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-errors/errors"
	"github.com/hanspr/clipboard"
	"github.com/hanspr/lang"
	"github.com/hanspr/terminfo"

	"github.com/hanspr/tcell/v2"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
	lua "github.com/yuin/gopher-lua"
	luar "layeh.com/gopher-luar"
)

const (
	doubleClickThreshold = 400 // How many milliseconds to wait before a second click is not a double click
	undoThreshold        = 500 // If two events are less than n milliseconds apart, undo both of them
)

// Force build, one time

// MouseClick mouse clic structure to follow mouse clics
type MouseClick struct {
	Click     bool
	Since     time.Time
	Pos       Loc
	MouseDown bool
	Button    int
}

// CursorType current cursor selected by user
type CursorType struct {
	color string
	shape string
}

// appEnv what environment we are
type appEnv struct {
	OS        string
	UID       int
	Gid       int
	Groups    []int
	ClipWhere string
}

var (
	// The main screen
	screen tcell.Screen

	// Object to send messages and prompts to the user
	messenger *Messenger

	// The default highlighting style
	// This simply defines the default foreground and background colors
	defStyle tcell.Style

	// Where the user's configuration is
	// This should be $XDG_CONFIG_HOME/mi-ide
	// If $XDG_CONFIG_HOME is not set, it is ~/.config/mi-ide
	configDir string

	// Version is the version number or commit hash
	// These variables should be set by the linker when compiling
	Version = "1.2.57"

	// The list of views
	tabs []*Tab
	// This is the currently open tab
	// It's just an index to the tab in the tabs array
	curTab int

	// Channel of jobs running in the background
	jobs chan JobFunction

	// Event channel
	events chan tcell.Event

	// Pointer Flag to check if App is running
	apprunning *MicroApp

	micromenu *microMenu

	// Language Translation object
	Language *lang.Lang

	// MicroToolBar toolbar object
	MicroToolBar *ToolBar

	scrollsince time.Time
	scrollcount int
	freeze      = false

	// Mouse follow the mouseclick
	Mouse MouseClick

	cursorInsert        CursorType
	cursorOverwrite     CursorType
	cursorHadShape      = false
	cursorHadColor      = false
	lastOverwriteStatus = false

	// Clip Clipboard object
	Clip *clipboard.Clipboard

	currEnv appEnv

	mouseEnabled        = false
	previousMouseStatus = false

	navigationMode = false

	cloudPath = "https://clip.microflow.com.mx:8443" // Cloud service url

	git *Gitstatus

	workingDir = ""

	homeDir = ""

	drawingLoc sync.Mutex

	gemini *geminiConnect
)

// LoadInput determines which files should be loaded into buffers
// based on the input stored in flag.Args()
func LoadInput() []*Buffer {
	// There are a number of ways mi-ide should start given its input

	// 1. If it is given a files in flag.Args(), it should open those

	// 2. If there is no input file and the input is not a terminal, that means
	// something is being piped in and the stdin should be opened in an
	// empty buffer

	// 3. If there is no input file and the input is a terminal, an empty buffer
	// should be opened

	var filename string
	var input []byte
	var err error
	var buf *Buffer

	args := flag.Args()
	buffers := make([]*Buffer, 0, len(args))
	if _, err := os.Stat(configDir + "/new.txt"); err == nil {
		os.Remove(configDir + "/new.txt")
		buf, err = NewBufferFromFile(configDir + "/help/welcome.md")
		if err == nil {
			buffers = append(buffers, buf)
		}
	}

	if len(args) > 0 {
		// Option 1
		// We go through each file and load it
		for i := range args {
			buf, err = NewBufferFromFile(args[i])
			if err != nil {
				TermMessage(err)
				continue
			}
			// If the file didn't exist, input will be empty, and we'll open an empty buffer
			buffers = append(buffers, buf)
		}
	} else if !isatty.IsTerminal(os.Stdin.Fd()) {
		// Option 2
		// The input is not a terminal, so something is being piped in
		// and we should read from stdin
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			TermMessage("Error reading from stdin: ", err)
			input = []byte{}
		}
		buffers = append(buffers, NewBufferFromString(string(input), filename))
	} else {
		// Option 3, just open an empty buffer
		buffers = append(buffers, NewBufferFromString(string(input), filename))
	}

	return buffers
}

// InitConfigDir finds the configuration directory for mi-ide according to the XDG spec.
// If no directory is found, it creates one.
func InitConfigDir() {
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		// The user has not set $XDG_CONFIG_HOME so we should act like it was set to ~/.config
		xdgHome = homeDir + "/.config"
	}
	configDir = xdgHome + "/mi-ide"

	if len(*flagConfigDir) > 0 {
		if _, err := os.Stat(*flagConfigDir); os.IsNotExist(err) {
			TermMessage("Error: " + *flagConfigDir + " does not exist. Defaulting to " + configDir + ".")
		} else {
			configDir = *flagConfigDir
			return
		}
	}

	if _, err := os.Stat(xdgHome); os.IsNotExist(err) {
		// If the xdgHome doesn't exist we should create it
		err = os.Mkdir(xdgHome, 0700)
		if err != nil {
			TermMessage("Error creating XDG_CONFIG_HOME directory: " + err.Error())
		}
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Copy files from /etc/mi-ide if they exists (installed by root)?
		// If the mi-ide specific config directory doesn't exist we should download a basic one
		TermMessage("Missing configuration files.\nmi-ide will download the necessary config files to run.\nFiles will be downloaded from\n\nhttps://raw.githubusercontent.com/hanspr/mi-channel/master/config.zip\n\nIf you do not agree, type Ctrl-C to abort, and install manually")
		err := DownLoadExtractZip("https://raw.githubusercontent.com/hanspr/mi-channel/master/config.zip", configDir)
		if err != nil {
			TermMessage("Could not download config files, please install manually.\n\nHave to abort.")
			os.Exit(0)
		}
	}
}

// InitScreen creates and initializes the tcell screen
func InitScreen() {
	// Should we enable true color?
	truecolor := os.Getenv("MIIDE_TRUECOLOR") == "1"

	tcelldb := os.Getenv("TCELLDB")
	os.Setenv("TCELLDB", configDir+"/.tcelldb")

	// In order to enable true color, we have to set the TERM to `xterm-truecolor` when
	// initializing tcell, but after that, we can set the TERM back to whatever it was
	oldTerm := os.Getenv("TERM")
	if truecolor {
		os.Setenv("TERM", "xterm-truecolor")
	}

	// Initialize tcell
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		if err == tcell.ErrTermNotFound {
			err = terminfo.WriteDB(configDir + "/.tcelldb")
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: mi-ide could not create tcelldb")
				os.Exit(1)
			}
			screen, err = tcell.NewScreen()
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: mi-ide could not initialize a screen.")
				os.Exit(1)
			}
		} else {
			fmt.Println(err)
			fmt.Println("Fatal: mi-ide could not initialize a screen.")
			os.Exit(1)
		}
	}
	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	screen.EnablePaste()
	if mouseEnabled {
		screen.EnableMouse()
	}

	// Now we can put the TERM back to what it was before
	if truecolor {
		os.Setenv("TERM", oldTerm)
	}

	os.Setenv("TCELLDB", tcelldb)
}

// RedrawAll redraws everything -- all the views and the messenger
func RedrawAll(show bool) {
	if freeze {
		return
	}
	drawingLoc.Lock()

	messenger.Clear()

	w, h := screen.Size()
	for x := range w {
		for y := range h {
			screen.SetContent(x, y, ' ', nil, defStyle)
		}
	}

	for _, v := range tabs[curTab].Views {
		v.Display()
	}
	DisplayTabs()
	messenger.Display()

	if show {
		screen.Show()
	}
	drawingLoc.Unlock()
}

func loadAll() {
	// Find the user's configuration directory (probably $XDG_CONFIG_HOME/micro)-ide
	InitConfigDir()

	// Build a list of available Extensions (Syntax, Colorscheme etc.)
	InitRuntimeFiles()

	// Load the user's settings
	InitGlobalSettings()

	InitCommands()
	InitBindings()

	InitColorscheme()
	LoadPlugins()

	for _, tab := range tabs {
		for _, v := range tab.Views {
			v.Buf.UpdateRules()
		}
	}
}

// Finish One Place Global Exit
// to control anything that could be necessary (to have a clean exit) in a single point
func Finish(status int) {
	screen.DisableMouse()
	time.Sleep(100 * time.Millisecond)
	if cursorHadColor {
		screen.SetCursorColorShape("white", "")
	}
	if cursorHadShape {
		screen.SetCursorColorShape("white", "block")
	}
	screen.Fini()
	os.Exit(status)
}

// Command line flags
var flagVersion = flag.Bool("version", false, "Show the version number and information")
var flagConfigDir = flag.String("config-dir", "", "Specify a custom location for the configuration directory")

// var flagOptions = flag.Bool("options", false, "Show all option help")

// MicroAppStop stop mi-ide app redraw icon bar
func MicroAppStop() {
	apprunning = nil
	MicroToolBar.FixTabsIconArea()
	MouseOnOff(previousMouseStatus)
}

func main() {
	// Create Toolbar object
	MicroToolBar = NewToolBar()
	apprunning = nil
	HelperWindow = nil
	currEnv.OS = runtime.GOOS
	currEnv.UID = os.Getuid()
	currEnv.Gid = os.Getgid()
	currEnv.Groups, _ = os.Getgroups()
	currEnv.ClipWhere = "local"
	workingDir, _ = os.Getwd()
	git = NewGitStatus()
	gemini = nil
	homeDir, _ = homedir.Dir()

	flag.Usage = func() {
		fmt.Println(ColorYellow + "Usage: mi-ide [OPTIONS] [FILE]..." + ColorReset)
		fmt.Println(ColorBold + "Available options" + ColorReset)
		fmt.Println(ColorBCyan + "--config-dir dir" + ColorReset)
		fmt.Println("    Specify a custom location for the configuration directory")
		fmt.Println(ColorBCyan + "--version" + ColorReset)
		fmt.Println("    Show the version number")
		fmt.Println(ColorBCyan + "\nQuick intro" + ColorReset)
		fmt.Println(ColorBold + "    Ctrl-o     : " + ColorReset + "Open file")
		fmt.Println(ColorBold + "    Ctrl-s     : " + ColorReset + "Save")
		fmt.Println(ColorBold + "    Ctrl-t     : " + ColorReset + "Create new tab")
		fmt.Println(ColorBold + "    Ctrl-w     : " + ColorReset + "Close current tab, helper window")
		fmt.Println(ColorBold + "    Ctrl-q     : " + ColorReset + "Exit")
		fmt.Println(ColorBold + "    Arrows     : " + ColorReset + "Move cursor around")
		fmt.Println(ColorBold + "    Ctrl-e     : " + ColorReset + "Enter command mode, tab to see list of commands")
		fmt.Println("                 Ctrl-e help defaultkeys")
	}

	optionFlags := make(map[string]*string)

	for k, v := range DefaultGlobalSettings() {
		optionFlags[k] = flag.String(k, "", fmt.Sprintf("The %s option. Default value: '%v'", k, v))
	}

	flag.Parse()

	if *flagVersion {
		// If -version was passed
		fmt.Println(Version)
		os.Exit(0)
	}

	// Start the Lua VM for running plugins
	L = lua.NewState()
	defer L.Close()

	// Find the user's configuration directory (probably $XDG_CONFIG_HOME/mi-ide)
	InitConfigDir()

	// Build a list of available Extensions (Syntax, Colorscheme etc.)
	InitRuntimeFiles()

	// Load the user's settings
	InitGlobalSettings()

	if _, err := os.Stat(configDir + "/libs/shared.lua"); err == nil {
		utf8ToShare, _ := CompileLua(configDir + "/libs/shared.lua")
		DoCompiledFile(L, utf8ToShare)
	}

	// Create translation object, to begin translating messages
	Language = lang.NewLang(globalSettings["lang"].(string), configDir+"/langs/"+globalSettings["lang"].(string)+".lang")

	// Create the clipboard
	Clip = clipboard.New()
	Clip.SetLocalPath(configDir + "/buffers/clips.txt")
	Clip.SetCloudPath(cloudPath, globalSettings["mi-key"].(string), globalSettings["mi-pass"].(string), globalSettings["mi-phrase"].(string))

	InitCommands()

	InitBindings()

	// Start the screen
	InitScreen()

	// This is just so if we have an error, we can exit cleanly and not completely
	// mess up the terminal being worked in
	// In other words we need to shut down tcell before the program crashes
	defer func() {
		if err := recover(); err != nil {
			screen.Fini()
			fmt.Println("mi-ide encountered an error:", err)
			// Print the stack trace too
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			os.Exit(1)
		}
	}()

	// Create a new messenger
	// This is used for sending the user messages in the bottom of the editor
	messenger = new(Messenger)
	messenger.timerOn = false
	messenger.LoadHistory()

	// Now we load the input
	buffers := LoadInput()
	if len(buffers) == 0 {
		Finish(1)
	}
	for _, buf := range buffers {
		// For each buffer we create a new tab and place the view in that tab
		tab := NewTabFromView(NewView(buf))
		tab.SetNum(len(tabs))
		tabs = append(tabs, tab)
		for _, t := range tabs {
			for _, v := range t.Views {
				v.Center(false)
			}
			t.Resize()

		}
	}

	// release the references we created, they are no longer needed
	buffers = nil

	for k, v := range optionFlags {
		if *v != "" {
			SetOption(k, *v)
		}
	}

	// Create mi-ide objects now to have them available to plugins
	// Menu
	micromenu = new(microMenu)
	x := strings.LastIndex(CurView().Buf.AbsPath, "/") + 1
	micromenu.LastPath = string([]rune(CurView().Buf.AbsPath)[:x])
	// Load all the plugin stuff
	// We give plugins access to a bunch of variables here which could be useful to them
	L.SetGlobal("OS", luar.New(L, currEnv.OS))
	L.SetGlobal("tabs", luar.New(L, tabs))
	L.SetGlobal("GetTabs", luar.New(L, func() []*Tab {
		return tabs
	}))
	L.SetGlobal("curTab", luar.New(L, curTab))
	L.SetGlobal("messenger", luar.New(L, messenger))
	L.SetGlobal("GetOption", luar.New(L, GetOption))
	L.SetGlobal("AddPluginOption", luar.New(L, AddPluginOption))
	L.SetGlobal("SetPluginOption", luar.New(L, SetPluginOption))
	L.SetGlobal("GetPluginOption", luar.New(L, GetPluginOption))
	L.SetGlobal("WritePluginSettings", luar.New(L, WritePluginSettings))
	L.SetGlobal("PluginGetApp", luar.New(L, PluginGetApp))
	L.SetGlobal("PluginRunApp", luar.New(L, PluginRunApp))
	L.SetGlobal("PluginStopApp", luar.New(L, PluginStopApp))
	L.SetGlobal("PluginAddIcon", luar.New(L, PluginAddIcon))
	L.SetGlobal("StringToStyle", luar.New(L, StringToStyle))
	L.SetGlobal("SetLocalOption", luar.New(L, SetLocalOption))
	L.SetGlobal("BindKey", luar.New(L, BindKey))
	L.SetGlobal("MakeCommand", luar.New(L, MakeCommand))
	L.SetGlobal("RemoveCommand", luar.New(L, RemoveCommand))
	L.SetGlobal("CurView", luar.New(L, CurView))
	L.SetGlobal("IsWordChar", luar.New(L, IsWordChar))
	L.SetGlobal("HandleCommand", luar.New(L, HandleCommand))
	L.SetGlobal("ExecCommand", luar.New(L, ExecCommand))
	L.SetGlobal("RunShellCommand", luar.New(L, RunShellCommand))
	L.SetGlobal("RunBackgroundShell", luar.New(L, RunBackgroundShell))
	L.SetGlobal("GetLeadingWhitespace", luar.New(L, GetLeadingWhitespace))
	L.SetGlobal("MakeCompletion", luar.New(L, MakeCompletion))
	L.SetGlobal("NewBuffer", luar.New(L, NewBufferFromString))
	L.SetGlobal("NewBufferFromFile", luar.New(L, NewBufferFromFile))
	L.SetGlobal("RuneStr", luar.New(L, func(r rune) string {
		return string(r)
	}))
	L.SetGlobal("Loc", luar.New(L, func(x, y int) Loc {
		return Loc{x, y}
	}))
	L.SetGlobal("WorkingDirectory", luar.New(L, os.Getwd))
	L.SetGlobal("JoinPaths", luar.New(L, filepath.Join))
	L.SetGlobal("DirectoryName", luar.New(L, filepath.Dir))
	L.SetGlobal("configDir", luar.New(L, configDir))
	L.SetGlobal("Reload", luar.New(L, loadAll))
	L.SetGlobal("ByteOffset", luar.New(L, ByteOffset))
	L.SetGlobal("ToCharPos", luar.New(L, ToCharPos))

	// Used for asynchronous jobs
	L.SetGlobal("JobStart", luar.New(L, JobStart))
	L.SetGlobal("JobSpawn", luar.New(L, JobSpawn))
	L.SetGlobal("JobSend", luar.New(L, JobSend))
	L.SetGlobal("JobStop", luar.New(L, JobStop))

	// Extension Files
	L.SetGlobal("ReadRuntimeFile", luar.New(L, PluginReadRuntimeFile))
	L.SetGlobal("ListRuntimeFiles", luar.New(L, PluginListRuntimeFiles))
	L.SetGlobal("AddRuntimeFile", luar.New(L, PluginAddRuntimeFile))
	L.SetGlobal("AddRuntimeFilesFromDirectory", luar.New(L, PluginAddRuntimeFilesFromDirectory))
	L.SetGlobal("AddRuntimeFileFromMemory", luar.New(L, PluginAddRuntimeFileFromMemory))

	// Access to Go stdlib
	L.SetGlobal("import", luar.New(L, Import))

	jobs = make(chan JobFunction, 100)
	events = make(chan tcell.Event, 1000)

	// Loading all plugins
	LoadPlugins()

	for _, t := range tabs {
		for _, v := range t.Views {
			GlobalPluginCall("onViewOpen", v)
			GlobalPluginCall("onBufferOpen", v.Buf)
		}
	}

	InitColorscheme()
	messenger.style = defStyle
	CurView().SetCursorEscapeString()
	git.GitSetStatus()

	// Here is the event loop which runs in a separate thread
	go func() {
		for {
			if screen != nil {
				events <- screen.PollEvent()
			}
		}
	}()

	for {
		// Display everything (if app is not running)
		if apprunning == nil {
			RedrawAll(true)
		}

		var event tcell.Event

		// Check for new events
		select {
		case f := <-jobs:
			// If a new job has finished while running in the background we should execute the callback
			f.function(f.output, f.args...)
			continue
		case event = <-events:
		}

		// If an app is running, send all events to it
		if apprunning != nil {
			apprunning.HandleEvents(event)
			continue
		}

		for event != nil {
			didAction := false

			switch e := event.(type) {
			case *tcell.EventResize:
				for _, t := range tabs {
					t.Resize()
				}
				didAction = true
			case *tcell.EventMouse:
				_, h := screen.Size()
				x, y := e.Position()
				button := e.Buttons()
				// Try to get full mouse clicks on different terminals
				if button == tcell.ButtonNone && Mouse.MouseDown {
					if Mouse.Pos.X == x && Mouse.Pos.Y == y {
						Mouse.Click = true
					}
					Mouse.MouseDown = false
				} else if button == tcell.ButtonNone {
					Mouse.Click = false
				} else {
					switch button {
					case tcell.Button1:
						Mouse.Button = 1
					case tcell.Button2:
						Mouse.Button = 2
					case tcell.Button3:
						Mouse.Button = 3
					default:
						Mouse.Button = 99
					}
					Mouse.Pos.X, Mouse.Pos.Y = e.Position()
					Mouse.MouseDown = true
					Mouse.Click = false
				}
				if Mouse.Click && y < 1 {
					// Event is on tabbar, process all events there
					TabbarHandleMouseEvent(event)
					didAction = true
				} else if Mouse.Click && Mouse.Button == 1 {
					// Mouse 1 click events only
					if y >= h-2 {
						// Ignore clicks on status and message area
						didAction = true
					}
				} else {
					if y >= h-1 {
						// Ignore moves at the message area
						didAction = true
					}
				}
			}
			if searching {
				// Search locks keyboard events to control move next/previous
				HandleSearchEvent(event, CurView())
			} else if !didAction {
				// Handle view and own statusline events
				CurView().HandleEvent(event)
			}

			select {
			case event = <-events:
			default:
				event = nil
			}
		}
	}
}
