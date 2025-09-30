package main

import (
	"bufio"
	"os"
	"strings"
)

type codecomplete struct {
	lang  map[string]map[string][]string
	ready map[string]bool
}

func newCodeComplete() *codecomplete {
	c := &codecomplete{
		lang:  make(map[string]map[string][]string),
		ready: make(map[string]bool),
	}
	return c
}

func (c *codecomplete) closeCodeComplete(lang string) {
	// on close
	// validar si existen más ventanas con el mismo lenguaje
	// SI: no hacer nada
	// NO: descargar todos los objetos para liberar espacio
	// c.lang[lang] = nil
	// delete(c.lang[lang])
	// delete(c.ready[lang]=false)
}

func (c *codecomplete) loadCompletions(lang string) {
	// if CurView() == nil {
	// 	TermMessage("Cargar completions")
	// } else {
	// 	messenger.AddLog("Cargar completions")
	// }
	c.loadCoreCompletions(lang)
	//TermMessage("Regresar")
}

func (c *codecomplete) loadCoreCompletions(lang string) {
	_, ok := c.lang[lang]
	if ok {
		// if CurView() == nil {
		// 	TermMessage("Ya hemos cargado completions para ", lang, " anteriormente")
		// } else {
		// 	messenger.AddLog("Ya hemos cargado completions para ", lang, " anteriormente")
		// }
		return
	}
	c.ready[lang] = false
	path := configDir + "/codecomplete/" + lang + "/core.txt"
	// if CurView() == nil {
	// 	TermMessage("Cargar ruta:", path)
	// } else {
	// 	messenger.AddLog("Cargar ruta:", path)
	// }
	file, err := os.Open(path)
	if err != nil {
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
	// if CurView() == nil {
	// 	TermMessage("estructura cargada!!")
	// } else {
	// 	messenger.AddLog("estructura cargada!!")
	// }
	c.ready[lang] = true
}
