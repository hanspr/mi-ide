package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

var (
	snipFileType = ""
	snippets     = map[string]*snippet{}
	snAutoclose  = false
)

type snippet struct {
	code string
}

func (s *snippet) addCodeLine(line string) {

}

func (s *snippet) prepare() {

}

func (s *snippet) clone() *snippet {
	return nil
}

func (s *snippet) str() string {
	return ""
}

func (s *snippet) findLocation(loc Loc) Loc {
	return Loc{-1, -1}
}

func (s *snippet) remove(loc Loc) []Loc {
	return nil
}

func (s *snippet) insert() {
}

func (s *snippet) focusNext() {

}

func (v *View) InsertSnippet(usePlugin bool) bool {
	c := v.Cursor
	buf := v.Buf
	name := ""
	ok := false
	snCode := &snippet{}
	snAutoclose = false
	if v.SelectWordLeft(false) {
		name = c.GetSelection()
	} else {
		messenger.AddLog("No word selected")
		return false
	}

	messenger.AddLog(name)

	loadSnippets(buf.FileType())
	if snCode, ok = snippets[name]; !ok {
		return false
	}
	if buf.Settings["autoclose"].(bool) {
		snAutoclose = true
		buf.Settings["autoclose"] = false
	}
	// Insert snippet
	c.DeleteSelection()
	c.ResetSelection()
	snCode.code = ""
	return true
}

func snippetAccept() {
	CurView().Buf.Settings["autoclose"] = snAutoclose
}

func loadSnippets(filetype string) {
	if filetype != snipFileType {
		snippets = ReadSnippets(filetype)
		snipFileType = filetype
	}
}

func ReadSnippets(filetype string) map[string]*snippet {
	lineNum := 0
	curName := ""
	rgxComment, _ := regexp.Compile(`^#`)
	rgxSnip, _ := regexp.Compile(`^snippet \w`)
	rgxCode, _ := regexp.Compile(`^\t`)
	snippets := map[string]*snippet{}
	filename := configDir + "settings/snippets/" + filetype + ".snippets"
	file, err := os.Open(filename)
	if err != nil {
		messenger.Error("No snippets file for " + filetype)
		return snippets
	}
	defer file.Close()
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
			snippets[name].code = ""
			curName = name
		} else if rgxSnip.Match(line) {
			// snippet code
			cline := strings.Replace(string(line), "\t", "", 1)
			snippets[curName].code = snippets[curName].code + string(cline)
		} else {
			messenger.AddLog("ReadSnippets error ", lineNum, " : ", string(line))
		}
	}
	return snippets
}
