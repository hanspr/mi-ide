package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-errors/errors"
	"github.com/hanspr/microidelibs/lang"
	"github.com/hanspr/microidelibs/terminfo"
	"github.com/hanspr/tcell"
	"github.com/hanspr/tcell/encoding"
	"github.com/mattn/go-isatty"
	"github.com/mitchellh/go-homedir"
	"github.com/yuin/gopher-lua"
	"layeh.com/gopher-luar"
)

const (
	doubleClickThreshold = 400 // How many milliseconds to wait before a second click is not a double click
	undoThreshold        = 500 // If two events are less than n milliseconds apart, undo both of them
	autosaveTime         = 8   // Number of seconds to wait before autosaving
)

var (
	// The main screen
	screen tcell.Screen

	// Object to send messages and prompts to the user
	messenger *Messenger

	// The default highlighting style
	// This simply defines the default foreground and background colors
	defStyle tcell.Style

	// Where the user's configuration is
	// This should be $XDG_CONFIG_HOME/micro-ide
	// If $XDG_CONFIG_HOME is not set, it is ~/.config/micro-ide
	configDir string

	// Version is the version number or commit hash
	// These variables should be set by the linker when compiling
	Version = "0.0.0-unknown"
	// CommitHash is the commit this version was built on
	CommitHash = "Unknown"
	// CompileDate is the date this binary was compiled on
	CompileDate = "Unknown"

	// The list of views
	tabs []*Tab
	// This is the currently open tab
	// It's just an index to the tab in the tabs array
	curTab int

	// Channel of jobs running in the background
	jobs chan JobFunction

	// Event channel
	events   chan tcell.Event
	autosave chan bool

	// Channels for the terminal emulator
	updateterm chan bool
	closeterm  chan int

	// How many redraws have happened
	numRedraw uint

	// Dir Tree View
	dirview *DirTreeView

	// Pointer Flag to check id App is running
	apprunning *MicroApp = nil

	micromenu *microMenu

	Language *lang.Lang

	MicroToolBar *ToolBar

	scrollsince time.Time
	scrollcount int

	wheelbounce int

	// Added to abort relocate cursor when changing views when using a mouse click
	MouseClick bool
)

// LoadInput determines which files should be loaded into buffers
// based on the input stored in flag.Args()
func LoadInput() []*Buffer {
	// There are a number of ways micro should start given its input

	// 1. If it is given a files in flag.Args(), it should open those

	// 2. If there is no input file and the input is not a terminal, that means
	// something is being piped in and the stdin should be opened in an
	// empty buffer

	// 3. If there is no input file and the input is a terminal, an empty buffer
	// should be opened

	var filename string
	var input []byte
	var err error
	args := flag.Args()
	buffers := make([]*Buffer, 0, len(args))

	if len(args) > 0 {
		// Option 1
		// We go through each file and load it
		for i := 0; i < len(args); i++ {
			if strings.HasPrefix(args[i], "+") {
				if strings.Contains(args[i], ":") {
					split := strings.Split(args[i], ":")
					*flagStartPos = split[0][1:] + "," + split[1]
				} else {
					*flagStartPos = args[i][1:] + ",0"
				}
				continue
			}

			buf, err := NewBufferFromFile(args[i])
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
		input, err = ioutil.ReadAll(os.Stdin)
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

// InitConfigDir finds the configuration directory for micro according to the XDG spec.
// If no directory is found, it creates one.
func InitConfigDir() {
	xdgHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgHome == "" {
		// The user has not set $XDG_CONFIG_HOME so we should act like it was set to ~/.config
		home, err := homedir.Dir()
		if err != nil {
			TermMessage("Error finding your home directory\nCan't load config files")
			return
		}
		xdgHome = home + "/.config"
	}
	configDir = xdgHome + "/micro-ide"

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
		err = os.Mkdir(xdgHome, os.ModePerm)
		if err != nil {
			TermMessage("Error creating XDG_CONFIG_HOME directory: " + err.Error())
		}
	}

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		// Copy files from /etc/micro-ide if they exists (installed by root)?
		// If the micro specific config directory doesn't exist we should download a basic one
		TermMessage("Missing configuration files.\nmicro-ide will download the necessary config files to run.\nFiles will be downloaded from\n\nhttps://raw.githubusercontent.com/hanspr/mi-channel/master/config.zip\n\nIf you do not agree, type Ctrl-C to abort, and install manually")
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
	truecolor := os.Getenv("MICRO_TRUECOLOR") == "1"

	tcelldb := os.Getenv("TCELLDB")
	os.Setenv("TCELLDB", configDir+"/.tcelldb")

	// In order to enable true color, we have to set the TERM to `xterm-truecolor` when
	// initializing tcell, but after that, we can set the TERM back to whatever it was
	oldTerm := os.Getenv("TERM")
	if truecolor {
		os.Setenv("TERM", "xterm-truecolor")
	}

	// Initilize tcell
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		if err == tcell.ErrTermNotFound {
			err = terminfo.WriteDB(configDir + "/.tcelldb")
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: Micro could not create tcelldb")
				os.Exit(1)
			}
			screen, err = tcell.NewScreen()
			if err != nil {
				fmt.Println(err)
				fmt.Println("Fatal: Micro could not initialize a screen.")
				os.Exit(1)
			}
		} else {
			fmt.Println(err)
			fmt.Println("Fatal: Micro could not initialize a screen.")
			os.Exit(1)
		}
	}
	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Now we can put the TERM back to what it was before
	if truecolor {
		os.Setenv("TERM", oldTerm)
	}

	if GetGlobalOption("mouse").(bool) {
		screen.EnableMouse()
	}

	os.Setenv("TCELLDB", tcelldb)

	// screen.SetStyle(defStyle)
}

// RedrawAll redraws everything -- all the views and the messenger
func RedrawAll(show bool) {

	messenger.Clear()

	w, h := screen.Size()
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
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

	if numRedraw%50 == 0 {
		runtime.GC()
	}
	numRedraw++
}

func LoadAll() {
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

// One Place Global Exit
// to control anything that could be necesary (to have a clean exit) in a sinlge point
func Finish(status int) {
	screen.Fini()
	os.Exit(status)
}

// Command line flags
var flagVersion = flag.Bool("version", false, "Show the version number and information")
var flagStartPos = flag.String("startpos", "", "LINE,COL to start the cursor at when opening a buffer.")
var flagConfigDir = flag.String("config-dir", "", "Specify a custom location for the configuration directory")
var flagOptions = flag.Bool("options", false, "Show all option help")

func MicroAppStop() {
	apprunning = nil
	MicroToolBar.FixTabsIconArea()
}

func main() {
	// Create Toolbar object
	MicroToolBar = NewToolBar()
	apprunning = nil

	flag.Usage = func() {
		fmt.Println("Usage: micro [OPTIONS] [FILE]...")
		fmt.Println("-config-dir dir")
		fmt.Println("    \tSpecify a custom location for the configuration directory")
		fmt.Println("-startpos LINE,COL")
		fmt.Println("+LINE:COL")
		fmt.Println("    \tSpecify a line and column to start the cursor at when opening a buffer")
		fmt.Println("    \tThis can also be done by opening file:LINE:COL")
		fmt.Println("-options")
		fmt.Println("    \tShow all option help")
		fmt.Println("-version")
		fmt.Println("    \tShow the version number and information")

		fmt.Print("\nMicro's options can also be set via command line arguments for quick\nadjustments. For real configuration, please use the settings.json\nfile (see 'help options').\n\n")
		fmt.Println("-option value")
		fmt.Println("    \tSet `option` to `value` for this session")
		fmt.Println("    \tFor example: `micro -syntax off file.c`")
		fmt.Println("\nUse `micro -options` to see the full list of configuration options")
	}

	optionFlags := make(map[string]*string)

	for k, v := range DefaultGlobalSettings() {
		optionFlags[k] = flag.String(k, "", fmt.Sprintf("The %s option. Default value: '%v'", k, v))
	}

	flag.Parse()

	if *flagVersion {
		// If -version was passed
		fmt.Println("Version:", Version)
		fmt.Println("Commit hash:", CommitHash)
		fmt.Println("Compiled on", CompileDate)
		os.Exit(0)
	}

	if *flagOptions {
		// If -options was passed
		for k, v := range DefaultGlobalSettings() {
			fmt.Printf("-%s value\n", k)
			fmt.Printf("    \tThe %s option. Default value: '%v'\n", k, v)
		}
		os.Exit(0)
	}

	// Start the Lua VM for running plugins
	L = lua.NewState()
	defer L.Close()

	// Some encoding stuff in case the user isn't using UTF-8
	encoding.Register()
	tcell.SetEncodingFallback(tcell.EncodingFallbackASCII)

	// Find the user's configuration directory (probably $XDG_CONFIG_HOME/micro-ide)
	InitConfigDir()

	// Build a list of available Extensions (Syntax, Colorscheme etc.)
	InitRuntimeFiles()

	// Load the user's settings
	InitGlobalSettings()

	// Create translation object, to begin translating messeges
	Language = lang.NewLang(globalSettings["lang"].(string), configDir+"/langs/"+globalSettings["lang"].(string)+".lang")

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
			fmt.Println("Micro-ide encountered an error:", err)
			// Print the stack trace too
			fmt.Print(errors.Wrap(err, 2).ErrorStack())
			os.Exit(1)
		}
	}()

	// Create a new messenger
	// This is used for sending the user messages in the bottom of the editor
	messenger = new(Messenger)
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

	for k, v := range optionFlags {
		if *v != "" {
			SetOption(k, *v)
		}
	}

	// Create micro-ide objects now to have them available to plugins

	// Tree view
	dirview = new(DirTreeView)
	dirview.Open = false
	x := strings.LastIndex(CurView().Buf.AbsPath, "/") + 1
	dirview.LastPath = string([]rune(CurView().Buf.AbsPath)[:x])

	// Menu
	micromenu = new(microMenu)

	// Load all the plugin stuff
	// We give plugins access to a bunch of variables here which could be useful to them
	L.SetGlobal("OS", luar.New(L, runtime.GOOS))
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
	L.SetGlobal("SetLocalOption", luar.New(L, SetLocalOption))
	L.SetGlobal("BindKey", luar.New(L, BindKey))
	L.SetGlobal("MakeCommand", luar.New(L, MakeCommand))
	L.SetGlobal("CurView", luar.New(L, CurView))
	L.SetGlobal("IsWordChar", luar.New(L, IsWordChar))
	L.SetGlobal("HandleCommand", luar.New(L, HandleCommand))
	L.SetGlobal("HandleShellCommand", luar.New(L, HandleShellCommand))
	L.SetGlobal("ExecCommand", luar.New(L, ExecCommand))
	L.SetGlobal("RunShellCommand", luar.New(L, RunShellCommand))
	L.SetGlobal("RunBackgroundShell", luar.New(L, RunBackgroundShell))
	L.SetGlobal("RunInteractiveShell", luar.New(L, RunInteractiveShell))
	L.SetGlobal("TermEmuSupported", luar.New(L, TermEmuSupported))
	L.SetGlobal("RunTermEmulator", luar.New(L, RunTermEmulator))
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
	L.SetGlobal("Reload", luar.New(L, LoadAll))
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
	events = make(chan tcell.Event, 100)
	autosave = make(chan bool)
	updateterm = make(chan bool)
	closeterm = make(chan int)

	// Loading all plugins
	LoadPlugins()

	for _, t := range tabs {
		for _, v := range t.Views {
			GlobalPluginCall("onViewOpen", v)
			GlobalPluginCall("onBufferOpen", v.Buf)
			v.Buf.LoadPluginOption()
		}
	}

	InitColorscheme()
	messenger.style = defStyle

	// Here is the event loop which runs in a separate thread
	go func() {
		for {
			if screen != nil {
				events <- screen.PollEvent()
			}
		}
	}()

	go func() {
		for {
			time.Sleep(autosaveTime * time.Second)
			if globalSettings["autosave"].(bool) {
				autosave <- true
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
		case <-autosave:
			if CurView().Buf.Path != "" {
				CurView().Save(true)
			}
		case <-updateterm:
			continue
		case vnum := <-closeterm:
			tabs[curTab].Views[vnum].CloseTerminal()
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
				MouseClick = false
				_, h := screen.Size()
				x, y := e.Position()
				if y < 1 {
					// Event is on tabbar, process all events there
					didAction = TabbarHandleMouseEvent(event)
				} else if e.Buttons() == tcell.Button1 && e.HasMotion() == false {
					MouseClick = true
					// Mouse click event
					if searching {
						// Happened during a search lock, release Search
						ExitSearch(CurView())
					} else if y == h-2 {
						// Statusline event
						if x < 3 {
							// Click on treeview icon area
							dirview.onIconClick()
						} else {
							// Other status lines events to handle
							CurView().HandleEvent(event)
						}
						didAction = true
					}
					if !didAction && CurView().mouseReleased {
						// We loop through each view in the current tab and make sure the current view
						// is the one being clicked in
						for _, v := range tabs[curTab].Views {
							if x >= v.x && x < v.x+v.Width && y >= v.y && y < v.y+v.Height {
								tabs[curTab].CurView = v.Num
							}
						}
					}
				} else if (e.Buttons() == tcell.Button3 || e.Buttons() == tcell.Button2) && e.HasMotion() == false {
					// Butttons 2,3 click on dirview; open file in view next to it
					if dirview.Open && dirview.tree_view == CurView() {
						v := CurView()
						x -= v.lineNumOffset - v.leftCol + v.x
						y += v.Topline - v.y
						v.MoveToMouseClick(x, y)
						dirview.onExecuteAction("openonView", v)
						didAction = true
					}
				} else if e.Buttons() == tcell.WheelUp || e.Buttons() == tcell.WheelDown {
					// Scroll event
					didAction = true
					if wheelbounce > 0 {
						wheelbounce = 0
						break
					}
					wheelbounce++
					var view *View
					for _, v := range tabs[curTab].Views {
						if x >= v.x && x < v.x+v.Width && y >= v.y && y < v.y+v.Height {
							view = tabs[curTab].Views[v.Num]
						}
					}
					if view != nil {
						view.HandleEvent(e)
					}
				}
			}
			if searching {
				// Search locks keyboard events to control move next/previous
				HandleSearchEvent(event, CurView())
			} else if !didAction {
				// Send it to the view, is a view event (including its own status line)
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
