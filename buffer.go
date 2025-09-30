package main

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/hanspr/highlight"
	"github.com/hanspr/ioencoder"
	"github.com/phayes/permbits"
)

// LargeFileThreshold Define a large file limit reference
const LargeFileThreshold = 50000

var (
	// 0 - no line type detected
	// 1 - lf detected
	// 2 - crlf detected
	fileformat = 0
)

// Buffer stores the text for files that are loaded into the text editor
// It uses a rope to efficiently store the string and contains some
// simple functions for saving and wrapper functions for modifying the rope
type Buffer struct {
	// The eventhandler for undo/redo
	*EventHandler
	// This stores all the text in the buffer as an array of lines
	*LineArray

	Cursor    Cursor
	cursors   []*Cursor // for multiple cursors
	curCursor int       // the current cursor

	// Path to the file on disk (relative to my initial current path)
	Path string
	// Absolute path to the file on disk
	AbsPath string
	// Name of the buffer on the status line
	name string
	// Name with no path to use on Tabs
	Fname string

	//Encoding
	encoder  string // encoder currently being used
	encoding bool   // file requires encoding (any encoder different than UTF8)

	// Whether or not the buffer has been modified since it was opened
	IsModified bool
	RO         bool

	// Stack Reference since saved, used to set buffer clean if undos arrives to
	// last clean state (monitors in Undo Redo Actions)
	UndoStackRef int

	// Stores the last modification time of the file the buffer is pointing to
	ModTime time.Time

	// NumLines is the number of lines in the buffer
	NumLines int

	syntaxDef   *highlight.Def
	highlighter *highlight.Highlighter

	// Buffer local settings
	Settings map[string]any

	pasteLoc Loc
	//pasteTime time
}

// GetFileSettings define basic preconfigured settings from a file, guessed or previously saved
func (b *Buffer) GetFileSettings(filename string) {
	filename, _ = filepath.Abs(filename)
	// Check read only flags in none windows environments
	if currEnv.OS != "windows" {
		// Test if file is write enabled for this user and file permissions
		if fi, err := os.Stat(filename); err == nil {
			b.RO = true
			perm, _ := permbits.Stat(filename)
			UID := fi.Sys().(*syscall.Stat_t).Uid
			gid := fi.Sys().(*syscall.Stat_t).Gid
			if currEnv.UID == 0 {
				b.RO = false
			} else if uint32(currEnv.UID) == UID && perm.UserWrite() {
				b.RO = false
			} else if uint32(currEnv.UID) != UID && uint32(currEnv.Gid) == gid && perm.GroupWrite() {
				b.RO = false
			} else if uint32(currEnv.UID) != UID && uint32(currEnv.Gid) != gid && perm.OtherWrite() {
				b.RO = false
			} else if uint32(currEnv.UID) != UID && uint32(currEnv.Gid) != gid && perm.GroupWrite() {
				for _, g := range currEnv.Groups {
					if uint32(g) == gid {
						b.RO = false
						break
					}
				}
			}
		}
	}
	b.encoder = "UTF8"
	// Find last encoding used for this file
	setpath := filename + ".settings"
	setpath = configDir + "/buffers/" + strings.ReplaceAll(setpath, "/", "")
	settings, jerr := ReadFileJSON(setpath)
	if jerr == nil {
		if settings["encoder"] != nil {
			b.encoder = settings["encoder"].(string)
		}
		if settings["blockopen"] == nil {
			settings["blockopen"] = ""
		}
		if settings["blockclose"] == nil {
			settings["blockclose"] = ""
		}
	}
}

// NewBufferFromFile opens a new buffer using the given path
// It will also automatically handle `~`, and line/column with filename:l:c
// It will return an empty buffer if the path does not exist
// and an error if the file is a directory
func NewBufferFromFile(path string) (*Buffer, error) {
	filename := path
	filename = ReplaceHome(filename)
	file, err := os.Open(filename)
	fileInfo, _ := os.Stat(filename)

	if err == nil && fileInfo.IsDir() {
		return nil, errors.New(filename + " " + Language.Translate("is a directory"))
	}

	defer file.Close()

	var buf *Buffer
	if err != nil {
		// File does not exist -- create an empty buffer with that name
		buf = NewBufferFromString("", filename)
	} else {
		buf = NewBuffer(file, FSize(file), filename)
	}
	go cc.loadCompletions(buf.FileType())

	return buf, nil
}

// NewBufferFromString creates a new buffer containing the given string
func NewBufferFromString(text, path string) *Buffer {
	return NewBuffer(strings.NewReader(text), int64(len(text)), path)
}

// NewBuffer creates a new buffer from a given reader with a given path
func NewBuffer(reader io.Reader, size int64, path string) *Buffer {
	var utf8reader io.Reader

	// check if the file is already open in a tab. If it's open return the buffer to that tab
	if path != "" {
		for _, tab := range tabs {
			for _, view := range tab.Views {
				if view.Buf.Path == path {
					return view.Buf
				}
			}
		}
	}

	b := new(Buffer)

	if reflect.TypeOf(reader).String() == "*os.File" && path != "" {
		// Check for previous saved settings
		b.GetFileSettings(path)
		if b.encoder == "UTF8" {
			utf8reader = reader
			b.encoding = false
		} else {
			var err error
			enc := ioencoder.New()
			utf8reader, err = enc.GetReader(b.encoder, reader)
			if err == nil {
				b.encoding = true
			} else {
				b.encoder = "UTF8"
				b.encoding = false
			}
		}
	} else {
		utf8reader = reader
		b.encoder = "UTF8"
		b.encoding = false
	}

	b.LineArray = NewLineArray(size, utf8reader)

	b.Settings = DefaultLocalSettings()
	for k, v := range globalSettings {
		if _, ok := b.Settings[k]; ok {
			b.Settings[k] = v
		}
	}

	switch fileformat {
	case 1:
		b.Settings["fileformat"] = "unix"
	case 2:
		b.Settings["fileformat"] = "dos"
	}

	absPath, _ := filepath.Abs(path)

	b.Path = path
	b.AbsPath = absPath
	b.Fname = filepath.Base(path)

	// The last time this file was modified
	b.ModTime, _ = GetModTime(b.Path)

	b.EventHandler = NewEventHandler(b)

	b.Update()
	b.UpdateRules()

	if _, err := os.Stat(configDir + "/buffers/"); os.IsNotExist(err) {
		os.Mkdir(configDir+"/buffers/", os.ModePerm)
	}

	b.Cursor = Cursor{
		Loc: Loc{0, 0},
		buf: b,
	}
	InitLocalSettings(b)
	b.cursors = []*Cursor{&b.Cursor}
	b.pasteLoc.X = -1
	return b
}

// GetName returns the name that should be displayed in the statusline
// for this buffer
func (b *Buffer) GetName() string {
	if b.name == "" {
		if b.Path == "" {
			return "No name"
		}
		return b.Path
	}
	return b.name
}

// UpdateRules updates the syntax rules and filetype for this buffer
// This is called when the colorscheme changes
func (b *Buffer) UpdateRules() {
	rehighlight := false
	var files []*highlight.File
	for _, f := range ListRuntimeFiles(RTSyntax) {
		data, err := f.Data()
		if err != nil {
			TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
		} else {
			file, err := highlight.ParseFile(data)
			if err != nil {
				TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
				continue
			}
			ftdetect, err := highlight.ParseFtDetect(file)
			if err != nil {
				TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
				continue
			}

			ft := b.Settings["filetype"].(string)
			if ft == "" && !rehighlight {
				if highlight.MatchFiletype(ftdetect, b.Path, b.lines[0].data) {
					header := new(highlight.Header)
					header.FileType = file.FileType
					header.FtDetect = ftdetect
					b.syntaxDef, err = highlight.ParseDef(file, header)
					if err != nil {
						TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
						continue
					}
					rehighlight = true
				}
			} else {
				if file.FileType == ft && !rehighlight {
					header := new(highlight.Header)
					header.FileType = file.FileType
					header.FtDetect = ftdetect
					b.syntaxDef, err = highlight.ParseDef(file, header)
					if err != nil {
						TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
						continue
					}
					rehighlight = true
				}
			}
			files = append(files, file)
		}
	}

	if b.syntaxDef == nil {
		f := FindRuntimeFile(RTSyntax, "text")
		if f == nil {
			TermMessage(Language.Translate("Error loading syntax file") + ": text.yalm")
		} else {
			data, err := f.Data()
			if err != nil {
				TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
			} else {
				file, err := highlight.ParseFile(data)
				if err != nil {
					TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
				} else {
					header := new(highlight.Header)
					header.FileType = "text"
					b.syntaxDef, err = highlight.ParseDef(file, header)
					if err != nil {
						TermMessage(Language.Translate("Error loading syntax file") + " " + f.Name() + ": " + err.Error())
					}
					rehighlight = true
				}
			}
		}
	} else {
		highlight.ResolveIncludes(b.syntaxDef, files)
	}

	if b.highlighter == nil || rehighlight {
		if b.syntaxDef != nil {
			b.Settings["filetype"] = b.syntaxDef.FileType
			b.highlighter = highlight.NewHighlighter(b.syntaxDef)
			if b.Settings["syntax"].(bool) {
				b.highlighter.HighlightStates(b)
			}
		}
	}
}

// FileType returns the buffer's filetype
func (b *Buffer) FileType() string {
	return b.Settings["filetype"].(string)
}

// IndentString returns a string representing one level of indentation
func (b *Buffer) IndentString() string {
	if b.Settings["tabstospaces"].(bool) {
		return Spaces(int(b.Settings["tabsize"].(float64)))
	}
	return "\t"
}

// AddMultiComment adds comment to multiple lines
func (b *Buffer) AddMultiComment(Start, Stop Loc) {
	cstring := b.Settings["comment"].(string)
	comment := regexp.MustCompile(`^\s*` + cstring)
	start := Start.Y
	end := Stop.Y
	if Start.Y == Stop.Y || Stop.X > 0 {
		end++
	}
	// Scan to decide if we add o remove comments
	action := ""
	for y := start; y < end; y++ {
		str := b.Line(y)
		if str == "" || IsSpacesOrTabs(str) {
			continue
		}
		if comment.MatchString(str) {
			if action == "add" {
				break
			} else if action == "" {
				action = "del"
			}
		} else {
			if action == "del" {
				action = "add"
				break
			} else if action == "" {
				action = "add"
			}
		}
	}
	for y := start; y < end; y++ {
		str := b.Line(y)
		if str == "" || IsSpacesOrTabs(str) {
			continue
		}
		if action == "del" {
			// Remove comment from line
			str = strings.Replace(str, cstring, "", 1)
		} else {
			// Add comment to line
			wsp := GetLeadingWhitespace(str)
			str = strings.Replace(str, wsp, "", 1)
			str = wsp + cstring + str
		}
		x := Count(b.Line(y))
		b.Replace(Loc{0, y}, Loc{x, y}, str)
	}
}

// SmartIndent indent the line
func (b *Buffer) SmartIndent(Start, Stop Loc) {
	bopen := `[{[(]$`
	bclose := `^[}\])]`
	binter := `^[}\])].+?[{[(]$`
	if b.Settings["blockopen"].(string) != "" && b.Settings["blockclose"].(string) != "" && b.Settings["blockinter"].(string) != "" {
		bopen = b.Settings["blockopen"].(string)
		bclose = b.Settings["blockclose"].(string)
		binter = b.Settings["blockinter"].(string)
	}
	iChar := b.Settings["indentchar"].(string)
	iMult := 1
	comment := regexp.MustCompile(`^\s*(?:#|//|(?:<!)?--|/\*)`)
	skipBlockStart := regexp.MustCompile(`^(?:#|//|(?:<!)?--|/\*)<<<`)
	skipBlockEnd := regexp.MustCompile(`^(?:#|//|(?:<!)?--|/\*)>>>`)
	skipBlock := false
	openBlock := regexp.MustCompile(bopen)
	closeBlock := regexp.MustCompile(bclose)
	interBlock := regexp.MustCompile(binter)
	if iChar == " " {
		iMult = int(b.Settings["tabsize"].(float64))
		iChar = strings.Repeat(" ", iMult)
	}
	indentChangeDone := false
	I := 0
	Ys := Start.Y - 1
	Ye := Stop.Y
	if Ys < 0 {
		Ys = 0
	} else {
		// Look back for the first line that is not empty, is not a comment, get that indentation as reference
	restart:
		for y := Ys; y >= 0; y-- {
			l := b.Line(y)
			isComment := comment.MatchString(l)
			if len(l) > 0 && !isComment {
				I = GetLineIndentetion(b.Line(y), iChar, iMult)
				if I < 0 && !indentChangeDone {
					indentChangeDone = true
					cfrom := "\t"
					if b.Settings["indentchar"].(string) == "\t" {
						cfrom = " "
					}
					//messenger.AddLog("smartindent, mixed: from >", cfrom, "< to >", b.Settings["indentchar"].(string), "< size:", b.Settings["tabsize"].(float64))
					b.ChangeIndentation(cfrom, b.Settings["indentchar"].(string), int(b.Settings["tabsize"].(float64)), int(b.Settings["tabsize"].(float64)))
					goto restart
				}
				break
			} else if isComment {
				Ys--
			}
		}
	}
	if Ys < 0 {
		Ys = 0
	}
	ci := 0
	IndRef := 0
	// messenger.AddLog("inicio encontrado:", Ys)
	// messenger.AddLog("I:", I)
	for y := Ys; y <= Ye; y++ {
		// messenger.AddLog("Linea:", string(b.Line(y)))
		lc := SmartIndentPrepareLine(b.Line(y))
		if comment.MatchString(lc) {
			ci = I + IndRef
		} else {
			// messenger.AddLog("linea:", lc)
			if skipBlock {
				// messenger.AddLog("skipblock")
				if skipBlockEnd.MatchString(lc) {
					skipBlock = false
				}
				continue
			}
			if skipBlockStart.MatchString(lc) {
				// messenger.AddLog("start block")
				skipBlock = true
				continue
			}
			// messenger.AddLog("ci + IndRef:", ci, "+", IndRef)
			ci = I + IndRef
			if interBlock.MatchString(lc) {
				// messenger.AddLog("interblock")
				if y == Ys {
					// if reference is an interblock, the reference already is correct, treat as indenting
					// messenger.AddLog("startline")
					IndRef++
					continue
				}
				ci--
			} else if openBlock.MatchString(lc) {
				// messenger.AddLog("indent")
				IndRef++
			} else if closeBlock.MatchString(lc) {
				// messenger.AddLog("outdent")
				if y == Ys {
					// messenger.AddLog("start line")
					// if reference outdent, is already outdented
					continue
				}
				IndRef--
				ci--
			}
		}
		if ci < 0 {
			//messenger.AddLog("NEGATIVO !!!!!")
			ci = 0
		}
		if GetLineIndentetion(b.Line(y), iChar, iMult) != ci {
			indentation := strings.Repeat(iChar, ci)
			li := CountLeadingWhitespace(b.Line(y))
			b.Replace(Loc{0, y}, Loc{li, y}, indentation)
		}
	}
}

// CheckModTime makes sure that the file this buffer points to hasn't been updated
// by an external program since it was last read
// If it has, we ask the user if they would like to reload the file
func (b *Buffer) CheckModTime() {
	modTime, ok := GetModTime(b.Path)
	if ok {
		if modTime != b.ModTime {
			if b.Settings["autoreload"].(bool) && !b.IsModified {
				messenger.Alert("info", Language.Translate("Buffer reloaded"))
				b.ReOpen()
			} else {
				choice, canceled := messenger.YesNoPrompt(Language.Translate("The file has changed since it was last read. Reload file? (y,n)"))
				messenger.Reset()
				messenger.Clear()
				if !choice || canceled {
					// Don't load new changes -- do nothing
					b.ModTime, _ = GetModTime(b.Path)
				} else {
					// Load new changes
					b.ReOpen()
				}
			}
		}
	}
}

var reopen = false

// ReOpen reloads the current buffer from disk
func (b *Buffer) ReOpen() {
	if reopen {
		return
	}
	var txt string
	reopen = true
	defer func() {
		reopen = false
	}()
	data, err := os.ReadFile(b.Path)
	if b.encoding {
		enc := ioencoder.New()
		txt = enc.DecodeString(b.encoder, string(data))
	} else {
		txt = string(data)
	}

	if err != nil {
		messenger.Alert("error", err.Error())
		return
	}
	b.EventHandler.ApplyDiff(txt)

	b.ModTime, _ = GetModTime(b.Path)
	b.IsModified = false
	b.Update()
	b.SmartDetections()
	git.GitSetStatus()
	b.Cursor.Relocate()
}

// Update fetches the string from the rope and updates the `text` and `lines` in the buffer
func (b *Buffer) Update() {
	b.NumLines = len(b.lines)
}

// MergeCursors merges any cursors that are at the same position
// into one cursor
func (b *Buffer) MergeCursors() {
	var cursors []*Cursor
	for i := range b.cursors {
		c1 := b.cursors[i]
		if c1 != nil {
			for j := 0; j < len(b.cursors); j++ {
				c2 := b.cursors[j]
				if c2 != nil && i != j && c1.Loc == c2.Loc {
					b.cursors[j] = nil
				}
			}
			cursors = append(cursors, c1)
		}
	}

	b.cursors = cursors

	for i := range b.cursors {
		b.cursors[i].Num = i
	}

	if b.curCursor >= len(b.cursors) {
		b.curCursor = len(b.cursors) - 1
	}
}

// UpdateCursors updates all the cursors indicies
func (b *Buffer) UpdateCursors() {
	for i, c := range b.cursors {
		c.Num = i
	}
}

// Save saves the buffer to its default path
func (b *Buffer) Save() error {
	return b.SaveAs(b.Path)
}

// SaveAs saves the buffer to a specified path (filename), creating the file if it does not exist
func (b *Buffer) SaveAs(filename string) error {
	b.UpdateRules()
	if b.Settings["rmtrailingws"].(bool) {
		for i, l := range b.lines {
			pos := len(bytes.TrimRightFunc(l.data, unicode.IsSpace))

			if pos < len(l.data) {
				b.deleteToEnd(Loc{pos, i})
			}
		}

		b.Cursor.Relocate()
	}

	if b.Settings["eofnewline"].(bool) {
		end := b.End()
		if b.RuneAt(Loc{end.X - 1, end.Y}) != '\n' {
			b.Insert(end, "\n")
		}
	}

	defer func() {
		b.ModTime, _ = GetModTime(filename)
	}()

	// Removes any tilde and replaces with the absolute path to home
	absFilename := ReplaceHome(filename)

	// Get the leading path to the file | "." is returned if there's no leading path provided
	if dirname := filepath.Dir(absFilename); dirname != "." {
		// Check if the parent dirs don't exist
		if _, statErr := os.Stat(dirname); os.IsNotExist(statErr) {
			// Prompt to make sure they want to create the dirs that are missing
			if yes, canceled := messenger.YesNoPrompt(Language.Translate("Parent folders") + " '" + dirname + "' " + Language.Translate("do not exist. Create them? (y,n)")); yes && !canceled {
				// Create all leading dir(s) since they don't exist
				if mkdirallErr := os.MkdirAll(dirname, os.ModePerm); mkdirallErr != nil {
					// If there was an error creating the dirs
					return mkdirallErr
				}
			} else {
				// If they canceled the creation of leading dirs
				return errors.New(Language.Translate("Save aborted"))
			}
		}
	}

	var fileSize int

	err := overwriteFile(absFilename, func(file io.Writer) (e error) {
		var fileutf8 io.Writer
		var err error
		var eol []byte
		enc := ioencoder.New()

		if b.encoding {
			fileutf8, err = enc.GetWriter(b.encoder, file)
			if err != nil {
				messenger.AddLog("Error!:", err.Error())
				messenger.AddLog("Original encoding:", b.encoder)
				messenger.AddLog("Save as UTF8")
				messenger.Alert("info", Language.Translate("File saved as UTF8"))
				fileutf8 = file
				b.encoder = "UTF8"
				b.encoding = false
			}
		} else {
			fileutf8 = file
		}
		if len(b.lines) == 0 {
			return
		}

		if b.Settings["fileformat"] == "dos" {
			eol = []byte{'\r', '\n'}
		} else {
			eol = []byte{'\n'}
		}

		// write lines
		if fileSize, e = fileutf8.Write(b.lines[0].data); e != nil {
			return
		}

		for _, l := range b.lines[1:] {
			if _, e = fileutf8.Write(eol); e != nil {
				return
			}

			if _, e = fileutf8.Write(l.data); e != nil {
				return
			}

			fileSize += len(eol) + len(l.data)
		}

		return
	})

	if err != nil {
		//messenger.AddLog(err.Error())
		if b.encoding {
			newerr := b.RetryOnceSaveAs(filename)
			if newerr != nil {
				return err
			}
			err = nil
		} else {
			return err
		}
	}

	b.Path = filename
	b.IsModified = false
	if b.encoder != "UTF8" {
		settings := make(map[string]string)
		cachename, _ := filepath.Abs(filename)
		cachename = configDir + "/buffers/" + strings.ReplaceAll(cachename+".settings", "/", "")
		settings["encoder"] = b.encoder
		err := UpdateFileJSON(cachename, settings)
		if err != nil {
			messenger.Alert("error", Language.Translate("Could not save settings")+" : "+err.Error())
		}
	}
	// Save current Undo Stack Len to later check Modified status in Actions
	b.UndoStackRef = b.EventHandler.UndoStack.Len()
	return nil
}

// RetryOnceSaveAs retray saving file as UTF8 in case of error
func (b *Buffer) RetryOnceSaveAs(filename string) error {
	enc := b.encoder
	b.encoder = "UTF8"
	b.encoding = false
	err := b.SaveAs(filename)
	if err != nil {
		b.encoder = enc
		b.encoding = true
		return errors.New(Language.Translate("File can not be saved with selected encoding"))
	}
	return err
}

// overwriteFile opens the given file for writing, truncating if one exists, and then calls
// the supplied function with the file as io.Writer object, also making sure the file is
// closed afterwards.
func overwriteFile(name string, fn func(io.Writer) error) (err error) {
	var file *os.File

	if file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}

	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = e
		}
	}()

	w := bufio.NewWriter(file)

	if err = fn(w); err != nil {
		return
	}

	err = w.Flush()
	return
}

// Modified returns if this buffer has been modified since
// being opened
func (b *Buffer) Modified() bool {
	return b.IsModified
}

func (b *Buffer) insert(pos Loc, value []byte) {
	b.IsModified = true
	b.LineArray.insert(pos, value)
	b.Update()
}
func (b *Buffer) remove(start, end Loc) string {
	b.IsModified = true
	sub := b.LineArray.remove(start, end)
	b.Update()
	return sub
}
func (b *Buffer) deleteToEnd(start Loc) {
	b.IsModified = true
	b.LineArray.DeleteToEnd(start)
	b.Update()
}

// Start returns the location of the first character in the buffer
func (b *Buffer) Start() Loc {
	return Loc{0, 0}
}

// End returns the location of the last character in the buffer
func (b *Buffer) End() Loc {
	return Loc{utf8.RuneCount(b.lines[b.NumLines-1].data), b.NumLines - 1}
}

// RuneAt returns the rune at a given location in the buffer
func (b *Buffer) RuneAt(loc Loc) rune {
	line := b.LineRunes(loc.Y)
	if len(line) > 0 {
		return line[loc.X]
	}
	return '\n'
}

// LineBytes returns a single line as an array of bytes
func (b *Buffer) LineBytes(n int) []byte {
	if n >= len(b.lines) {
		return []byte{}
	}
	return b.lines[n].data
}

// LineRunes returns a single line as an array of runes
func (b *Buffer) LineRunes(n int) []rune {
	if n >= len(b.lines) {
		return []rune{}
	}
	return toRunes(b.lines[n].data)
}

// Line returns a single line
func (b *Buffer) Line(n int) string {
	if n >= len(b.lines) {
		return ""
	}
	return string(b.lines[n].data)
}

// LinesNum returns the number of lines in the buffer
func (b *Buffer) LinesNum() int {
	return len(b.lines)
}

// Lines returns an array of strings containing the lines from start to end
func (b *Buffer) Lines(start, end int) []string {
	lines := b.lines[start:end]
	var slice []string
	for _, line := range lines {
		slice = append(slice, string(line.data))
	}
	return slice
}

// Len gives the length of the buffer
func (b *Buffer) Len() (n int) {
	for _, l := range b.lines {
		n += utf8.RuneCount(l.data)
	}

	if len(b.lines) > 1 {
		n += len(b.lines) - 1 // account for newlines
	}

	return
}

// MoveLinesUp moves the range of lines up one row
func (b *Buffer) MoveLinesUp(start int, end int) {
	// 0 < start < end <= len(b.lines)
	if start < 1 || start >= end || end > len(b.lines) {
		return // what to do? FIXME
	}
	if end == len(b.lines) {
		b.Insert(
			Loc{
				utf8.RuneCount(b.lines[end-1].data),
				end - 1,
			},
			"\n"+b.Line(start-1),
		)
	} else {
		b.Insert(
			Loc{0, end},
			b.Line(start-1)+"\n",
		)
	}
	b.Remove(
		Loc{0, start - 1},
		Loc{0, start},
	)
}

// MoveLinesDown moves the range of lines down one row
func (b *Buffer) MoveLinesDown(start int, end int) {
	// 0 <= start < end < len(b.lines)
	// if end == len(b.lines), we can't do anything here because the
	// last line is unaccessible, FIXME
	if start < 0 || start >= end || end >= len(b.lines)-1 {
		return // what to do? FIXME
	}
	b.Insert(
		Loc{0, start},
		b.Line(end)+"\n",
	)
	end++
	b.Remove(
		Loc{0, end},
		Loc{0, end + 1},
	)
}

// RemoveTrailingSpace function to remove any leftover white spaces at the end of the buffer
func (b *Buffer) RemoveTrailingSpace(pos Loc) {
	line := b.Line(pos.Y)[pos.X:]
	end := len(b.LineRunes(pos.Y))
	re, _ := regexp.Compile(`[\t ]+$`)
	if re.MatchString(line) {
		line = re.ReplaceAllString(line, "")
		b.Replace(Loc{pos.X, pos.Y}, Loc{end, pos.Y}, line)
	}
}

// ClearMatches clears all of the syntax highlighting for this buffer
func (b *Buffer) ClearMatches() {
	for i := range b.lines {
		b.SetMatch(i, nil)
		b.SetState(i, nil)
	}
}

func (b *Buffer) clearCursors() {
	for i := 1; i < len(b.cursors); i++ {
		b.cursors[i] = nil
	}
	b.cursors = b.cursors[:1]
	b.UpdateCursors()
	b.Cursor.ResetSelection()
}

var bracePairs = [][2]rune{
	{'(', ')'},
	{'{', '}'},
	{'[', ']'},
}

// FindMatchingBrace returns the location in the buffer of the matching bracket
// It is given a brace type containing the open and closing character, (for example
// '{' and '}') as well as the location to match from
func (b *Buffer) FindMatchingBrace(braceType [2]rune, start Loc) Loc {
	curLine := b.LineRunes(start.Y)
	startChar := curLine[start.X]
	var i int
	switch startChar {
	case braceType[0]:
		for y := start.Y; y < b.NumLines; y++ {
			l := b.LineRunes(y)
			xInit := 0
			if y == start.Y {
				xInit = start.X
			}
			for x := xInit; x < len(l); x++ {
				r := l[x]
				switch r {
				case braceType[0]:
					i++
				case braceType[1]:
					i--
					if i == 0 {
						return Loc{x, y}
					}
				}
			}
		}
	case braceType[1]:
		for y := start.Y; y >= 0; y-- {
			l := []rune(string(b.lines[y].data))
			xInit := len(l) - 1
			if y == start.Y {
				xInit = start.X
			}
			for x := xInit; x >= 0; x-- {
				r := l[x]
				switch r {
				case braceType[0]:
					i--
					if i == 0 {
						return Loc{x, y}
					}
				case braceType[1]:
					i++
				}
			}
		}
	}
	return start
}

// ChangeIndentation re indent buffer if user changed indentation rules
func (b *Buffer) ChangeIndentation(cfrom, cto string, nfrom, nto int) {
	var ss, sr string
	if cfrom == "\t" {
		ss = cfrom
	} else {
		ss = strings.Repeat(" ", nfrom)
	}
	if cto == "\t" {
		sr = cto
		b.setIndentationOptions("tab", nto)
	} else {
		sr = strings.Repeat(" ", nto)
		b.setIndentationOptions("space", nto)
	}
	for y := range b.LinesNum() - 1 {
		l := b.Line(y)
		ws := GetLeadingWhitespace(l)
		wt := strings.ReplaceAll(ws, ss, sr)
		newline := strings.Replace(l, ws, wt, 1)
		b.lines[y].data = []byte(newline)
	}
	b.IsModified = true
}

// SmartDetections Check buffer to confirm current settings are consistent
func (b *Buffer) SmartDetections() {
	check := 0
	end := b.LinesNum()
	tablines := 0
	spacelines := 0
	spclen := 999
	found := false
	retab := regexp.MustCompile(`^\t+`)
	respc := regexp.MustCompile(`^  +`)
	for i := range end {
		l := b.Line(i)
		if Count(l) < 2 {
			continue
		}
		// Check 40 lines, after we find the first indentation
		if check > 40 {
			break
		}
		if retab.FindString(l) != "" {
			tablines++
			found = true
		} else {
			spc := respc.FindString(l)
			if spc != "" {
				spacelines++
				found = true
				if len(spc) < spclen {
					spclen = len(spc)
				}
			}
		}
		if found {
			check++
		}
	}
	if found {
		if tablines > spacelines {
			b.setIndentationOptions("tab", int(b.Settings["tabsize"].(float64)))
		} else {
			b.setIndentationOptions("space", spclen)
		}
	}
}

func (b *Buffer) setIndentationOptions(indent string, n int) {
	if indent == "tab" {
		b.Settings["indentchar"] = "\t"
		b.Settings["tabstospaces"] = false
		b.Settings["tabmovement"] = false
	} else {
		b.Settings["indentchar"] = " "
		b.Settings["tabstospaces"] = true
		b.Settings["tabmovement"] = true
	}
	b.Settings["tabsize"] = float64(n)
}

// Buffer Utilities

// RunFormatter check if formmatter is enabled for the current buffer and a formatter exists
// If formatter exists, run it
func (b *Buffer) RunFormatter() bool {
	if !b.Settings["useformatter"].(bool) {
		return false
	}
	formatterPath := configDir + "/formatters/" + b.FileType()
	info, err := os.Stat(formatterPath)
	if err != nil {
		return false
	}
	mode := info.Mode()
	if !strings.Contains(mode.String(), "rwx") {
		return false
	}
	_, err = ExecCommand(formatterPath, b.AbsPath)
	return err == nil
}
