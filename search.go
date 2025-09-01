package main

import (
	//"fmt"
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/hanspr/tcell/v2"
)

var (
	// What was the last search
	lastSearch string

	// Where should we start the search down from (or up from)
	searchStart Loc

	// Is there currently a search in progress
	searching bool

	// Stores the history for searching
	//searchHistory []string
)

// StartSearchMode starts the current search
func StartSearchMode() {
	messenger.hasPrompt = false
	searching = true
	messenger.Message(Language.Translate("Find") + " :" + lastSearch + "   " + Language.Translate("F5,Backspace (Previous)  F6,Enter (Next)"))
}

// ExitSearch exits the search mode, reset active search phrase, and clear status bar
func ExitSearch(v *View) {
	lastSearch = ""
	searching = false
	messenger.hasPrompt = false
	messenger.Clear()
	messenger.Reset()
	v.Cursor.ResetSelection()
}

// HandleSearchEvent takes an event and a view and will do a real time match from the messenger's output
// to the current buffer. It searches down the buffer.
func HandleSearchEvent(event tcell.Event, v *View) {
	if !searching {
		return
	}
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyF5, tcell.KeyBackspace2:
			v.FindPrevious(true)
		case tcell.KeyF6, tcell.KeyEnter:
			v.FindNext(true)
			return
		default:
			loc := v.Cursor.CurSelection
			ExitSearch(v)
			v.Cursor.GotoLoc(loc[0])
			if e.Rune() == 0 {
				v.HandleEvent(event)
			}
			return
		}
	}
}

func searchDown(r *regexp.Regexp, v *View, start, end Loc) bool {
	var startX int
	newLineSearch := strings.Contains(r.String(), `\n`) && !strings.Contains(r.String(), `\\n`)
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i <= end.Y; i++ {
		if i == start.Y {
			startX = start.X
		} else {
			startX = -1
		}

		l := string(v.Buf.lines[i].data)
		if newLineSearch && v.Buf.NumLines > i+1 {
			l = l + "\n" + string(v.Buf.lines[i+1].data)
		}
		match := r.FindAllStringIndex(l, -1)

		if match != nil {
			for j := 0; j < len(match); j++ {
				X := runePos(match[j][0], l)
				Y := runePos(match[j][1], l)
				if X >= startX {
					nl := i
					if newLineSearch {
						nl++
						Y = Y - len(v.Buf.lines[i].data) - 1
						if Y < 0 {
							Y = 0
						}
					}
					v.Cursor.SetSelectionStart(Loc{X, i})
					v.Cursor.SetSelectionEnd(Loc{Y, nl})
					v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
					v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
					v.Cursor.Loc = v.Cursor.CurSelection[0]
					if v.Cursor.Y >= v.Bottomline()-2 || v.Cursor.Y <= v.Topline+2 {
						v.Center(false)
					}
					return true
				}
			}
		}
	}
	return false
}

func searchUp(r *regexp.Regexp, v *View, start, end Loc) bool {
	var startX int
	newLineSearch := strings.Contains(r.String(), "\\n")
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i >= end.Y; i-- {
		if i == start.Y {
			startX = v.Cursor.CurSelection[0].X
		} else {
			startX = 9999999
		}

		l := string(v.Buf.lines[i].data)
		if newLineSearch && v.Buf.NumLines > i+1 {
			l = l + "\n" + string(v.Buf.lines[i+1].data)
		}
		match := r.FindAllStringIndex(l, -1)

		if match != nil {
			for j := len(match) - 1; j >= 0; j-- {
				X := runePos(match[j][0], l)
				Y := runePos(match[j][1], l)
				if X < startX {
					nl := i
					if newLineSearch {
						nl++
						Y = Y - len(v.Buf.lines[i].data) - 1
						if Y < 0 {
							Y = 0
						}
					}
					v.Cursor.SetSelectionStart(Loc{X, i})
					v.Cursor.SetSelectionEnd(Loc{Y, nl})
					v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
					v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
					v.Cursor.Loc = v.Cursor.CurSelection[0]
					if v.Cursor.Y >= v.Bottomline()-2 || v.Cursor.Y <= v.Topline+2 {
						v.Center(false)
					}
					return true
				}
			}
		}
	}
	return false
}

// Search searches in the view for the given regex. The down bool
// specifies whether it should search down from the searchStart position
// or up from there
func Search(searchStr string, v *View, down bool) bool {
	if searchStr == "" {
		return false
	}
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return false
	}

	var found bool
	if down {
		found = searchDown(r, v, searchStart, v.Buf.End())
		if !found {
			v.searchLoops++
			found = searchDown(r, v, v.Buf.Start(), searchStart)
		}
	} else {
		found = searchUp(r, v, searchStart, v.Buf.Start())
		if !found {
			v.searchLoops++
			found = searchUp(r, v, v.Buf.End(), searchStart)
		}
	}
	if !found {
	} else {
		v.Relocate()
	}
	return found
}

// Search dialogs preview of found text

const doff int = 20

// DialogSearch display search dialog box
func DialogSearch(searchStr string) string {
	if searchStr == "" {
		return ""
	}
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return ""
	}
	v := CurView()
	found := searchDown(r, v, v.searchSave, v.Buf.End())
	if !found {
		found = searchDown(r, v, v.Buf.Start(), v.searchSave)
	}
	if found {
		xs := v.Cursor.CurSelection
		line := CurView().Buf.LineRunes(xs[0].Y)
		x1 := 0
		if xs[0].X-doff >= 0 {
			x1 = xs[0].X - doff
		}
		d1 := 0
		d2 := 0
		if xs[0].X > len(line) || xs[1].X > len(line) {
			d1 = len(line) - xs[0].X
			d2 = len(line) - xs[1].X
		}
		return string(line[x1:xs[0].X+d1]) + "{f}" + v.Cursor.GetSelection() + "{/f}" + string(line[xs[1].X+d2:])
	}
	return ""
}

// Search for function declaration

func FindLineWith(r *regexp.Regexp, v *View, start, end Loc, deep bool) (int, bool) {
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i <= end.Y; i++ {
		l := string(v.Buf.lines[i].data)
		if r.MatchString(l) {
			return i, true
		}
	}
	if deep || start.Y == 0 {
		return 0, false
	}
	return FindLineWith(r, v, Loc{0, 0}, start, true)
}

const MAX_FAILS = 20

var findFileDeep = -1

// Recursively test each file to find a matching string on regex
// If dir fails N times aborts searching to avoid large subdirectories with no code files
func FindFileWith(r *regexp.Regexp, path, filetype, ext string, depth int, hint bool) (string, int, bool) {
	var prevLines = make([]string, 5)
	if findFileDeep < 0 {
		findFileDeep = depth
	}
	abort := MAX_FAILS
	rtf := FindRuntimeFile(RTSyntax, filetype)
	files, _ := os.ReadDir(path)
	for _, f := range files {
		if f.IsDir() {
			if depth == 0 {
				continue
			}
			depth--
			path := path + "/" + f.Name()
			filename, line, ok := FindFileWith(r, path, filetype, ext, depth, hint)
			depth++
			if ok {
				return filename, line, ok
			}
			continue
		}
		filepath := path + "/" + f.Name()
		ftype := TestFileType(filepath, rtf)
		if strings.Contains(f.Name(), ext) || ftype {
			abort = MAX_FAILS
			i := 0
			pos := -1
			file, err := os.Open(filepath)
			if err != nil {
				continue
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				l := scanner.Text()
				if hint {
					pos++
					if pos > 4 {
						for i := 0; i < 4; i++ {
							prevLines[i] = prevLines[i+1]
						}
						pos = 4
					}
					prevLines[pos] = l
				}
				match := r.FindAllStringIndex(l, -1)
				if match != nil {
					if !hint {
						return filepath, i, true
					}
					comment := regexp.MustCompile(`^\s*(?:#|//|(?:<!)?--|/\*)`)
					data := ""
					for i := 0; i < 4; i++ {
						if comment.MatchString(prevLines[i]) {
							data = data + prevLines[i] + "\n"
						}
					}
					data = data + prevLines[4] + "\n"
					next4 := 0
					for scanner.Scan() {
						if next4 > 5 {
							break
						}
						l := scanner.Text()
						data = data + l + "\n"
						next4++
					}
					return data, 0, true
				}
				i++
			}
		} else if findFileDeep != depth {
			abort--
			if abort <= 0 {
				return "", 0, false
			}
		}
	}
	return "", 0, false
}
