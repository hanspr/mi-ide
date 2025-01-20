package main

import "strings"

type gitstatus struct {
	enabled bool
	status  string
	fgcolor map[string]string
	bgcolor string
}

func NewGitStatus() *gitstatus {
	g := new(gitstatus)
	g.enabled = false
	g.status = " "
	g.fgcolor = make(map[string]string)
	g.bgcolor = "#000000"
	g.CheckGit()
	return g
}

func (g *gitstatus) CheckGit() {
	_, err := RunShellCommand("git status --porcelain")
	if err == nil {
		g.status = " "
		g.enabled = true
		g.GitSetStatus()
		return
	}
	g.enabled = false
}

func (g *gitstatus) GitSetStatus() {
	if !g.enabled {
		return
	}
	status, err := RunShellCommand("git status --porcelain")
	if err != nil {
		g.enabled = false
		g.status = " "
		return
	}
	branch, _ := RunShellCommand("git branch --show-current")
	branch = branch[:len(branch)-1]
	g.status = "[ " + branch + "{"
	g.enabled = true
	lines := strings.Split(status, "\n")
	m := false
	u := false
	s := false
	for _, l := range lines {
		if (!m || !s) && strings.Contains(l, "M") {
			if !s && (strings.Contains(l, "MM ") || strings.Contains(l, "M  ")) {
				// staged
				g.status = g.status + "+"
				g.fgcolor["+"] = "gold"
				s = true
			}
			if !m && (strings.Contains(l, "MM ") || strings.Contains(l, " M ")) {
				m = true
				g.status = g.status + "m"
				g.fgcolor["m"] = "red"
			}
		}
		if !u && strings.Contains(l, "?? ") {
			u = true
			g.status = g.status + "u"
			g.fgcolor["u"] = "#5fd7ff"
		}
		if u && m && s {
			break
		}
	}
	g.status = g.status + "}"
}
