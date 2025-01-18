package main

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	snipFileType   = ""
	snippets       = map[string]*snippet{}
	snAutoclose    = false
	currentSnippet = &snippet{}
)

// Location
type SnippetLocation struct {
	idx     int
	ph      *ph
	snippet *snippet
}

func NewSnippetLocation(idx int, ph *ph, snip *snippet) *SnippetLocation {
	sl := &SnippetLocation{}
	sl.idx = idx
	sl.ph = ph
	sl.snippet = snip
	return sl
}

func (sl *SnippetLocation) offset() int {
	add := 0
	for _, loc := range sl.snippet.locations {
		if loc == sl {
			break
		}
		if loc.ph.value != "" {
			add = add + Count(loc.ph.value)
		}
	}
	return sl.idx + add
}

func (sl *SnippetLocation) startPos() Loc {
	loc := sl.snippet.startPos
	return loc.Move(sl.offset(), sl.snippet.view.Buf)
}

func (sl *SnippetLocation) len() int {
	l := 0
	if sl.ph.value != "" {
		l = Count(sl.ph.value)
	}
	if l <= 0 {
		return 1
	}
	return l
}

func (sl *SnippetLocation) endPos() Loc {
	start := sl.startPos()
	return start.Move(sl.len(), sl.snippet.view.Buf)
}

// check if the given loc is within the location
func (sl *SnippetLocation) isWithin(loc Loc) bool {
	return loc.GreaterEqual(sl.startPos()) && loc.LessEqual(sl.endPos())
}

func (sl *SnippetLocation) focus() {
	view := sl.snippet.view
	startP := sl.startPos().Move(-1, view.Buf)
	endP := sl.endPos().Move(-1, view.Buf)

	for view.Cursor.LessThan(startP) {
		view.Cursor.Right()
	}
	for view.Cursor.GreaterThan(endP) {
		view.Cursor.Left()
	}
	if Count(sl.ph.value) > 0 {
		view.Cursor.SetSelectionStart(startP)
		view.Cursor.SetSelectionEnd(endP)
	} else {
		view.Cursor.ResetSelection()
	}
}

func (sl *SnippetLocation) handleInput(ev *TextEvent) bool {
	if ev.EventType == 1 {
		if ev.Deltas[1].Text == "\n" {
			sl.snippet.view.SnippetAccept(false)
			return false
		} else {
			offset := 1
			sp := sl.startPos()
			for sp.LessEqual(ev.Deltas[1].Start) {
				sp = sp.Move(1, sl.snippet.view.Buf)
				offset = offset + 1
			}
			sl.snippet.remove()
			if offset == 1 {
				sl.ph.value = ev.Deltas[1].Text + sl.ph.value[:offset]
			} else {
				sl.ph.value = sl.ph.value[0:offset-1] + ev.Deltas[1].Text + sl.ph.value[:offset]
			}
			sl.snippet.insert()
			return true
		}
	} else if ev.EventType == -1 {
		offset := 1
		sp := sl.startPos()
		for sp.LessEqual(ev.Deltas[1].Start) {
			sp = sp.Move(1, sl.snippet.view.Buf)
			offset = offset + 1
		}
		if ev.Deltas[1].Start.Y != ev.Deltas[1].End.Y {
			return false
		}
		sl.snippet.remove()
		l := ev.Deltas[1].End.X - ev.Deltas[1].Start.X
		if offset == 1 {
			sl.ph.value = sl.ph.value[:offset+l]
		} else {
			sl.ph.value = sl.ph.value[0:offset-1] + sl.ph.value[offset+l:]
		}
		sl.snippet.insert()
		return true
	}
	return false
}

// Snippet

type snippet struct {
	code         string
	locations    []*SnippetLocation
	placeholders []*ph
	startPos     Loc
	modText      bool
	view         *View
	focused      int
}

type ph struct {
	num   int
	value string
}

func NewSnippet() *snippet {
	s := &snippet{}
	s.code = ""
	return s
}

func (s *snippet) addCodeLine(line string) {
	if s.code != "" {
		s.code = s.code + "\n"
	}
	s.code = s.code + line
}

func (s *snippet) prepare() {
	if s.placeholders == nil {
		s.placeholders = make([]*ph, 1)
		s.locations = make([]*SnippetLocation, 1)
		rgx, _ := regexp.Compile(`\${(\d+):?([^}]*)}`)
		for {
			match := rgx.FindStringSubmatch(s.code)
			if match == nil {
				break
			}
			num, _ := strconv.ParseInt(match[1], 10, 0)
			value := match[2]
			idx := rgx.FindStringIndex(s.code)
			r := rgx.FindString(s.code)
			if r != "" {
				s.code = strings.Replace(s.code, r, "", 1)
			}
			p := s.placeholders[num]
			if p == nil {
				p = &ph{num: int(num), value: value}
				s.placeholders = append(s.placeholders, p)
			}
			s.locations = append(s.locations, NewSnippetLocation(idx[0], p, s))
		}
	}
}

func (s *snippet) clone() *snippet {
	result := NewSnippet()
	result.addCodeLine(s.code)
	result.prepare()
	return result
}

func (s *snippet) str() string {
	res := s.code
	for i := len(s.locations); i > 0; i-- {
		loc := s.locations[i]
		res = res[0:loc.idx-1] + loc.ph.value + res[loc.idx+1:]
	}
	return res
}

func (s *snippet) findLocation(loc Loc) *SnippetLocation {
	for _, l := range s.locations {
		if l.isWithin(loc) {
			return l
		}
	}
	return nil
}

func (s *snippet) remove() {
	endPos := s.startPos.Move(len(s.str()), s.view.Buf)
	s.modText = true
	s.view.Cursor.SetSelectionStart(s.startPos)
	s.view.Cursor.SetSelectionEnd(endPos)
	s.view.Cursor.DeleteSelection()
	s.view.Cursor.ResetSelection()
	s.modText = false
}

func (s *snippet) insert() {
	s.modText = true
	s.view.Buf.insert(s.startPos, []byte(s.str()))
	s.modText = false
}

func (s *snippet) focusNext() {
	if s.focused != 0 {
		s.focused = (s.focused + 1) % len(s.placeholders)
	}
	ph := s.placeholders[s.focused+1]
	for _, l := range s.locations {
		if l.ph == ph {
			l.focus()
			return
		}
	}
}

func loadSnippets(filetype string) {
	messenger.AddLog("leyendo snippets, para lenguaje : ", filetype)
	if filetype != snipFileType {
		snippets = ReadSnippets(filetype)
		snipFileType = filetype
	}
}

func ReadSnippets(filetype string) map[string]*snippet {
	messenger.AddLog("Leyendo snippets")
	lineNum := 0
	rgxComment, _ := regexp.Compile(`^#`)
	rgxSnip, _ := regexp.Compile(`^snippet \w`)
	rgxCode, _ := regexp.Compile(`^\t`)
	snippets := map[string]*snippet{}
	filename := configDir + "settings/snippets/" + filetype + ".snippets"
	file, err := os.Open(filename)
	if err != nil {
		messenger.AddLog("No snippets for file ", filetype)
		messenger.Error("No snippets file for ", filetype)
		return snippets
	}
	defer file.Close()
	curSnip := &snippet{}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if rgxComment.Match(line) {
			// comment line
			continue
		} else if rgxCode.Match(line) {
			// snippet word
			name := strings.Replace(string(line), "snippet ", "", 1)
			curSnip = NewSnippet()
			snippets[name] = curSnip
		} else if rgxSnip.Match(line) {
			// snippet code
			cline := strings.Replace(string(line), "\t", "", 1)
			curSnip.addCodeLine(cline)
		} else {
			messenger.AddLog("ReadSnippets error ", lineNum, " : ", string(line))
		}
	}
	return snippets
}

func SnippetOnBeforeTextEvent(ev *TextEvent) bool {
	if currentSnippet != nil && currentSnippet.view == CurView() {
		if currentSnippet.modText {
			return true
		}
		locStart := &SnippetLocation{}
		locEnd := &SnippetLocation{}
		if currentSnippet != nil {
			locStart = currentSnippet.findLocation(ev.Deltas[1].Start.Move(1, currentSnippet.view.Buf))
			locEnd = currentSnippet.findLocation(ev.Deltas[1].End)
		}
		if locStart != nil && (locStart == locEnd || (ev.Deltas[1].End.Y == 0 && ev.Deltas[1].End.X == 0)) {
			if locStart.handleInput(ev) {
				currentSnippet.view.Cursor.Goto(ev.C)
				return false
			}
		}
		currentSnippet.view.SnippetAccept(false)
	}
	return true
}

// Keybinding calls
func (v *View) SnippetInsert(usePlugin bool) bool {
	c := v.Cursor
	buf := v.Buf
	xy := Loc{c.X, c.Y}
	name := ""
	ok := false
	curSn := &snippet{}
	snAutoclose = false
	if v.SelectWordLeft(false) {
		name = c.GetSelection()
	} else {
		messenger.AddLog("No word selected")
		return false
	}

	loadSnippets(buf.FileType())
	if curSn, ok = snippets[name]; !ok {
		messenger.Message("Unknown snippet : ", name)
		return false
	}
	if buf.Settings["autoclose"].(bool) {
		snAutoclose = true
		buf.Settings["autoclose"] = false
	}
	currentSnippet = curSn.clone()
	currentSnippet.view = v
	currentSnippet.startPos = xy.Move(-len(name), v.Buf)
	currentSnippet.modText = true
	c.SetSelectionStart(currentSnippet.startPos)
	c.SetSelectionEnd(xy)
	c.DeleteSelection()
	c.ResetSelection()
	currentSnippet.modText = false
	currentSnippet.insert()
	if len(currentSnippet.placeholders) == 0 {
		pos := currentSnippet.startPos.Move(len(currentSnippet.str()), v.Buf)
		for v.Cursor.LessThan(pos) {
			v.Cursor.Right()
		}
		for v.Cursor.GreaterThan(pos) {
			v.Cursor.Left()
		}
	} else {
		currentSnippet.focusNext()
	}
	return true
}

func (v *View) SnippetNext(usePlugin bool) bool {
	if currentSnippet != nil {
		currentSnippet.focusNext()
	}
	return true
}

func (v *View) SnippetAccept(usePlugin bool) bool {
	v.Buf.Settings["autoclose"] = snAutoclose
	currentSnippet = nil
	return true
}

func (v *View) SnippetCancel(usePlugin bool) bool {
	if currentSnippet != nil {
		currentSnippet.remove()
	}
	return true
}
