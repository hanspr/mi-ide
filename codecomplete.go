package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
)

// Init codecomplete env
func codeCompleteInit() {
	dir := configDir + "/codecomplete"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0700)
		if err != nil {
			TermMessage("Error creating code complete directory: " + err.Error())
		}
	}
}

// map of codecomplete pointers
var cc map[string]*codecomplete

// Code autocomplete data for a particular language
type codecomplete struct {
	b             *Buffer
	bufferScanned bool
	lang          string
	workdir       string
	rulesFile     string
	wordsFile     string
	words         []string
	ready         bool
	scanning      bool
	rules         []*regexp.Regexp
}

// AddCodeComplete: Adds a new codecomplete function if not already loaded
func AddCodeComplete(buf *Buffer) *codecomplete {
	ftype := buf.FileType()
	if p, ok := cc[ftype]; ok {
		dir := GetProjectDir(filepath.Dir(buf.AbsPath))
		if dir == p.workdir {
			// Same language , same project, share words
			return p
		}
		// Same language, different project, different words
	}
	p := NewCodeComplete(buf)
	cc[ftype] = p
	return p
}

// NewCodeComplete: Create code complete function
func NewCodeComplete(buf *Buffer) *codecomplete {
	cc := &codecomplete{
		b:             buf,
		bufferScanned: false,
		lang:          buf.FileType(),
		rulesFile:     configDir + "/codecomplete/" + buf.FileType() + ".rules",
		words:         []string{},
		ready:         false,
		scanning:      false,
	}
	// load rules and words if necessary
	go cc.LoadRules()
	return cc
}

// LoadRules: loads rules from current language
func (cc *codecomplete) LoadRules() {
	// we have done this
	if cc.ready {
		return
	}
	if _, err := os.Stat(cc.rulesFile); os.IsNotExist(err) {
		return
	}
	file, err := os.Open(cc.rulesFile)
	if err != nil {
		return
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		rule, err := regexp.Compile(scanner.Text())
		if err == nil {
			cc.rules = append(cc.rules, rule)
		} else {
			messenger.AddLog("rule ignored:", scanner.Text())
			messenger.AddLog("error:", err.Error())
		}
	}
	// We have rules we can scan the buffer
	go cc.LoadWords()
}

// LoadWords: load words from current project
func (cc *codecomplete) LoadWords() {
	if cc.ready {
		return
	}
	dir := GetProjectDir(filepath.Dir(cc.b.AbsPath))
	cc.workdir = dir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.Mkdir(dir+"/.miide", os.ModePerm)
		// no words, scan buffer, create new file for the next time
		go cc.ScanBuffer(true)
		// todo: scan full project dir
		return
	}
	// recover previous work, no need to scan de buffer again
	cc.wordsFile = dir + "/.miide/" + cc.b.FileType() + ".words"
	file, err := os.Open(cc.wordsFile)
	if err != nil {
		return
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		cc.words = append(cc.words, scanner.Text())
	}
	cc.ready = true
}

// ScanBuffer: scan this buffer
func (cc *codecomplete) ScanBuffer(write bool) {
	if cc.bufferScanned {
		return
	}
	cc.bufferScanned = true
	// todo : scan buffer for words
	// test each line for a match, and save in cc.words array
	if write {
		// new words array save result
		cc.WriteWords()
	}
	cc.ready = true
}

// WriteWords : Save all scanned words to file
func (cc *codecomplete) WriteWords() {
	f, err := os.Create(cc.wordsFile)
	if err != nil {
		return
	}
	defer f.Close()
	for _, word := range cc.words {
		_, err = f.WriteString(word + "\n")
		if err != nil {
			continue
		}
	}
}

// Running process

// ScanLine: scan for new words on the current buffer and filetype
func ScanLine(b *Buffer, char string) {
}

// FindWord : take tokens, form a searchable word, return matches
func FindWord(b *Buffer, token string) []string {
	words := []string{}
	// todo : find words in current filetype and return array
	return words
}
