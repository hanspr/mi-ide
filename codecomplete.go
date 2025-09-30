package main

import (
	"bufio"
	"os"
	"strings"
	"sync"
)

type codecomplete struct {
	lang  map[string]map[string][]string
	ready map[string]bool
	count map[string]int
	mtx   sync.Mutex
}

func newCodeComplete() *codecomplete {
	c := &codecomplete{
		lang:  make(map[string]map[string][]string),
		ready: make(map[string]bool),
		count: make(map[string]int),
	}
	return c
}

func (c *codecomplete) closeCodeComplete(lang string) {
	_, ok := c.lang[lang]
	if !ok {
		//messenger.AddLog("Nothing to unload for",lang)
		return
	}
	c.mtx.Lock()
	// messenger.AddLog("A", c.count[lang])
	c.count[lang]--
	// messenger.AddLog("D", c.count[lang])
	if c.count[lang] > 0 {
		c.mtx.Unlock()
		return
	}
	// messenger.AddLog("remove ", lang)
	delete(c.lang, lang)
	delete(c.ready, lang)
	delete(c.count, lang)
	c.mtx.Unlock()
}

func (c *codecomplete) loadCompletions(lang string) {
	c.mtx.Lock()
	c.loadCoreCompletions(lang)
	c.mtx.Unlock()
}

func (c *codecomplete) loadCoreCompletions(lang string) {
	_, ok := c.lang[lang]
	if ok {
		c.count[lang]++
		return
	}
	c.ready[lang] = false
	c.count[lang] = 0
	path := configDir + "/codecomplete/" + lang + "/core.txt"
	file, err := os.Open(path)
	if err != nil {
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
	c.count[lang]++
	c.ready[lang] = true
}

//todo: load current buffer completes

//todo: create autocomplete search and paint
