package main

import (
	"regexp"
	"strings"
)

// Gitstatus struc to hold git status state
type Gitstatus struct {
	enabled bool
	status  string
	fgcolor map[string]string
	bgcolor string
}

// NewGitStatus create git status object
func NewGitStatus() *Gitstatus {
	g := new(Gitstatus)
	g.enabled = false
	g.status = " "
	g.fgcolor = make(map[string]string)
	g.bgcolor = "#437e0c"
	g.CheckGit()
	return g
}

// CheckGit validate if dir is a git repo
func (g *Gitstatus) CheckGit() bool {
	if g.enabled {
		return true
	}
	_, err := ExecCommand("git", "status", "--porcelain")
	if err == nil {
		g.enabled = true
		return true
	}
	g.enabled = false
	return false
}

// GitSetStatus validate, status, branch
// create string to display in statusbar
func (g *Gitstatus) GitSetStatus() {
	if !g.enabled {
		return
	}
	status, err := ExecCommand("git", "status", "--porcelain")
	if err != nil {
		g.enabled = false
		g.status = " "
		return
	}
	// git 2.22+
	branch, err := ExecCommand("git", "branch", "--show-current")
	if err != nil {
		// git -2.22
		branch = ""
		text, err := ExecCommand("git", "branch")
		if err != nil {
			TermMessage(text)
			g.enabled = false
			g.status = " "
			return
		}
		re := regexp.MustCompile(`\* (\w+)`)
		matches := re.FindStringSubmatch(text)
		if matches != nil {
			branch = matches[1]
		}
	} else {
		branch = branch[:len(branch)-1]
	}
	g.status = ""
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
	if m || s || u {
		g.status = "] " + branch + "{" + g.status + "}"
	} else {
		g.status = "[ " + branch + "}"
	}
	if CurView() != nil {
		CurView().sline.Display()
	}
}
