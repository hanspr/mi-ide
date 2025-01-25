package main

import (
	"bufio"
	"os"
	"regexp"
)

func codeCompleteInit() {
	dir := configDir + "/codecomplete"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.Mkdir(dir, 0700)
		if err != nil {
			TermMessage("Error creating code complete directory: " + err.Error())
		}
	}
}

var cc map[string]*codecomplete

// Code autocomplete data for a particular language
type codecomplete struct {
	b        *Buffer
	lang     string
	file     string
	words    []string
	ready    bool
	scanning bool
	rules    []*regexp.Regexp
}

func AddCodeComplete(buf *Buffer) *codecomplete {
	ftype := buf.FileType()
	if p, ok := cc[ftype]; ok {
		return p
	}
	p := NewCodeComplete(buf)
	cc[ftype] = p
	return p
}

// NewCodeComplete: Create code complete function
func NewCodeComplete(buf *Buffer) *codecomplete {
	cc := &codecomplete{
		b:        buf,
		lang:     buf.FileType(),
		file:     configDir + "/codecomplete/" + buf.FileType() + ".rules",
		words:    []string{},
		ready:    false,
		scanning: false,
	}
	go cc.LoadRules()
	return cc
}

// LoadWords: load words from current language
func (cc *codecomplete) LoadRules() {
	if _, err := os.Stat(cc.file); os.IsNotExist(err) {
		return
	}
	file, err := os.Open(cc.file)
	if err != nil {
		return
	}

	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		rule := scanner.Text()
		cc.rules = append(cc.rules, regexp.MustCompile(rule))
	}
	cc.ready = true
}

// WriteWords : Save words to file
func (cc *codecomplete) WriteWords() {
	f, err := os.Create(cc.file)
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

// ScanWords : Scan words
func (cc *codecomplete) ScanWords() {
	if cc.scanning {
		return
	}
	cc.scanning = true
	go cc.ScanBuffer()
}

// ScanBuffer: internal, do not call directly
func (cc *codecomplete) ScanBuffer() {
	// todo : scan buffer for new words and add to cc.words
}

// FindWord : take a word an find all possible matches and return
func (cc *codecomplete) FindWord(word string) []string {
	words := []string{}
	// todo : find words in current filetype and return array
	return words
}
