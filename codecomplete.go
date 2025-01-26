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

// GetCodeComplete: Adds a new codecomplete function if is not already loaded
// if already exists, reuse same process
// Should be called by at buffer creation and stored in the buffer struct
// nil: could not create a codecomplete process
// missing information or setup problems
func GetCodeCompletePointer(buf *Buffer) *codecomplete {
	ftype := buf.FileType()
	dir := GetProjectDir(filepath.Dir(buf.AbsPath), true)
	if p, ok := cc[ftype+":"+dir]; ok {
		return p
	}
	p := NewCodeComplete(buf)
	cc[ftype+":"+dir] = p
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
	// load rules
	if cc.LoadRules() {
		// We have rules we can scan the buffer
		go cc.LoadWords()
		return cc
	}
	return nil
}

// LoadRules: loads rules from current language
func (cc *codecomplete) LoadRules() bool {
	if _, err := os.Stat(cc.rulesFile); os.IsNotExist(err) {
		messenger.AddLog("No rules file: ", err.Error())
		return false
	}
	file, err := os.Open(cc.rulesFile)
	if err != nil {
		messenger.AddLog("Rules file inaccessible: ", err.Error())
		return false
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		rule, err := regexp.Compile(scanner.Text())
		if err == nil {
			cc.rules = append(cc.rules, rule)
		} else {
			messenger.AddLog("rules have errors: ", scanner.Text())
			messenger.AddLog("error: ", err.Error())
			return false
		}
	}
	return true
}

// LoadWords: load words from current project
func (cc *codecomplete) LoadWords() {
	if cc.ready {
		return
	}
	dir := GetProjectDir(filepath.Dir(cc.b.AbsPath), true)
	cc.workdir = dir
	cc.wordsFile = dir + "/.miide/" + cc.b.FileType() + ".words"
	file, err := os.Open(cc.wordsFile)
	if err != nil {
		// no file exists yet, fill info with current buffer
		cc.ScanBuffer(true)
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
	if cc.bufferScanned || !cc.ready {
		return
	}
	cc.bufferScanned = true
	// todo : scan buffer for words to store
	// test each line for a match, and save in cc.words array
	if write {
		// new words array save result
		cc.WriteWords()
	}
	cc.ready = true
}

// WriteWords : Save all scanned words to file
func (cc *codecomplete) WriteWords() {
	if !cc.ready {
		return
	}
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

// title :Running process

// ScanLine: scan for new words on the current buffer and filetype
func (cc *codecomplete) ScanLine(word string) {
	if !cc.ready {
		return
	}
}

// FindWord : take tokens, form a searchable word, return matches
func (cc *codecomplete) FindWord(token string) []string {
	words := []string{}
	if !cc.ready {
		return words
	}
	// search for codeline pointer based on : filetype , path
	// not found return
	// test cc.ready
	// todo : find words in current filetype and return array
	return words
}
