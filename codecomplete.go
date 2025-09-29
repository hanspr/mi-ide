package main

import (
	"bufio"
	"os"
	"strings"
)

type codecomplete struct {
	lang  map[string]map[string][]string
	ready bool
}

func newCodeComplete() *codecomplete {
	c := &codecomplete{
		lang:  make(map[string]map[string][]string),
		ready: false,
	}
	return c
}

func closeCodeComplete(c *codecomplete) {
	c = nil
}

func (c *codecomplete) loadCompletions(lang string) {
	TermMessage("Cargar completions")
	// if CurView() == nil {
	// 	TermMessage("Cargar completions")
	// } else {
	// 	messenger.AddLog("Cargar completions")
	// }
	c.loadCoreCompletions(lang)
	TermMessage("Regresar")
}

func (c *codecomplete) loadCoreCompletions(lang string) {
	_, ok := c.lang[lang]
	if ok {
		TermMessage("Ya hemos cargado completions para ", lang, " anteriormente")
		// if CurView() == nil {
		// 	TermMessage("Ya hemos cargado completions para ", lang, " anteriormente")
		// } else {
		// 	messenger.AddLog("Ya hemos cargado completions para ", lang, " anteriormente")
		// }
		return
	}
	path := configDir + "/codecomplete/" + lang + "/core.txt"
	TermMessage("Cargar ruta:", path)
	// if CurView() == nil {
	// 	TermMessage("Cargar ruta:", path)
	// } else {
	// 	messenger.AddLog("Cargar ruta:", path)
	// }
	file, err := os.Open(path)
	if err != nil {
		TermMessage("Error:", err.Error())
		// if CurView() == nil {
		// 	TermMessage("Error:", err.Error())
		// } else {
		// 	messenger.AddLog("Error:", err.Error())
		// }
		return
	}
	defer file.Close()
	c.lang[lang] = make(map[string][]string)
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		data := strings.Split(scanner.Text(), "|")
		c.lang[lang][data[0]] = data[1:]
	}
	TermMessage("estructura cargada!!")
	// if CurView() == nil {
	// 	TermMessage("estructura cargada!!")
	// } else {
	// 	messenger.AddLog("estructura cargada!!")
	// }
	c.ready = true
}
