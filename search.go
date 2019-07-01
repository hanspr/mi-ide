package main

import (
	//"fmt"
	"regexp"

	"github.com/hanspr/tcell"
)

var (
	// What was the last search
	lastSearch string

	// Where should we start the search down from (or up from)
	searchStart Loc

	// Is there currently a search in progress
	searching bool

	// Stores the history for searching
	searchHistory []string
)

// EndSearch stops the current search
func StartSearchMode() {
	messenger.hasPrompt = false
	searching = true
	messenger.Message("Find: " + lastSearch + "   " + "Esc,CtrlG (Exit)  F5 (Previous)  F6 (Next)")
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
	if searching == false {
		return
	}
	switch e := event.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlQ, tcell.KeyCtrlG:
			// Exit the search mode
			ExitSearch(v)
			return
		case tcell.KeyF5:
			v.FindPrevious(true)
			return
		case tcell.KeyF6:
			v.FindNext(true)
			return
		}
	}
	return
}

func searchDown(r *regexp.Regexp, v *View, start, end Loc) bool {
	var startX int
	if start.Y >= v.Buf.NumLines {
		start.Y = v.Buf.NumLines - 1
	}
	if start.Y < 0 {
		start.Y = 0
	}
	for i := start.Y; i <= end.Y; i++ {
		if i == start.Y {
			startX = v.Cursor.CurSelection[0].X
		} else {
			startX = -1
		}

		l := string(v.Buf.lines[i].data)
		match := r.FindAllStringIndex(l, -1)

		if match != nil {
			for j := 0; j < len(match); j++ {
				X := runePos(match[j][0], l)
				Y := runePos(match[j][1], l)
				if X > startX {
					v.Cursor.SetSelectionStart(Loc{X, i})
					v.Cursor.SetSelectionEnd(Loc{Y, i})
					v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
					v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
					v.Cursor.Loc = v.Cursor.CurSelection[1]
					if v.Cursor.Y >= v.Bottomline()-2 {
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
		match := r.FindAllStringIndex(l, -1)

		if match != nil {
			for j := len(match) - 1; j >= 0; j-- {
				X := runePos(match[j][0], l)
				Y := runePos(match[j][1], l)
				if X < startX {
					v.Cursor.SetSelectionStart(Loc{X, i})
					v.Cursor.SetSelectionEnd(Loc{Y, i})
					v.Cursor.OrigSelection[0] = v.Cursor.CurSelection[0]
					v.Cursor.OrigSelection[1] = v.Cursor.CurSelection[1]
					v.Cursor.Loc = v.Cursor.CurSelection[1]
					if v.Cursor.Y <= v.Topline+2 {
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
func Search(searchStr string, v *View, down bool) {
	if searchStr == "" {
		return
	}
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return
	}

	var found bool
	if down {
		found = searchDown(r, v, searchStart, v.Buf.End())
		if !found {
			found = searchDown(r, v, v.Buf.Start(), searchStart)
		}
	} else {
		found = searchUp(r, v, searchStart, v.Buf.Start())
		if !found {
			found = searchUp(r, v, v.Buf.End(), searchStart)
		}
	}
	if !found {
	} else {
		v.Relocate()
	}
}

func DialogSearch(searchStr string) string {
	if searchStr == "" {
		return ""
	}
	r, err := regexp.Compile(searchStr)
	if err != nil {
		return ""
	}
	v := CurView()
	if searchDown(r, v, v.searchSave, v.Buf.End()) == true {
		return v.Cursor.GetSelection()
	} else {
		return ""
	}
}
