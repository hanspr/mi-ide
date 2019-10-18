package main

import "github.com/hanspr/highlight"

var syntaxFiles []*highlight.File

//LoadSyntaxFiles load syntax files
func LoadSyntaxFiles() {
	InitColorscheme()
	for _, f := range ListRuntimeFiles(RTSyntax) {
		data, err := f.Data()
		if err != nil {
			TermMessage("Error loading syntax file " + f.Name() + ": " + err.Error())
		} else {
			LoadSyntaxFile(data, f.Name())
		}
	}
}

// LoadSyntaxFile load a single file
func LoadSyntaxFile(text []byte, filename string) {
	f, err := highlight.ParseFile(text)

	if err != nil {
		TermMessage("Syntax file error: " + filename + ": " + err.Error())
		return
	}

	syntaxFiles = append(syntaxFiles, f)
}
