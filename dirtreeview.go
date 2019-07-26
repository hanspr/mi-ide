// dirtreeview.go

package main

import (
	"io/ioutil"
	"strings"
	"time"
)

type DirTreeView struct {
	// Open : true, false
	Open      bool
	LastPath  string
	tree_view *View
	pos       Loc
	curView   int
	active    bool
}

func (d *DirTreeView) opeFile(line string) {
	if len(line) < 5 {
		return
	}
	fname := string([]rune(line)[4:])
	if strings.Contains(fname, "../") {
		if d.curView == 1 {
			return
		}
		d.pos.X = 0
		d.pos.Y = 0
		// Remove las "/" from Path
		x := strings.LastIndex(d.LastPath, "/")
		fname := string([]rune(d.LastPath)[:x])
		// Find the new last "/" now : PATH/DIR
		x = strings.LastIndex(fname, "/") + 1
		// Now find the directory name above[x:] DIR, and path to show [:x] PATH
		search := "\\s" + string([]rune(d.LastPath)[x:]) + "$"
		fname = string([]rune(d.LastPath)[:x])
		d.LastPath = fname
		messenger.Information(d.LastPath)
		time.Sleep(200 * time.Millisecond)
		d.viewDir()
		Search(search, d.tree_view, true)
	} else if strings.Contains(fname, "/") {
		if d.curView == 1 {
			return
		}
		d.pos.X = 0
		d.pos.Y = 0
		d.LastPath = d.LastPath + fname
		messenger.Information(d.LastPath)
		d.viewDir()
	} else {
		d.pos = d.tree_view.Cursor.Loc
		if d.curView == 0 {
			for i, t := range tabs {
				for _, v := range t.Views {
					if strings.Contains(v.Buf.Path, fname) {
						d.dirviewClose()
						messenger.Information(fname, " : Focused")
						curTab = i
						return
					}
				}
			}
			dirview.dirviewClose()
			NewTab([]string{d.LastPath + fname})
			CurView().Buf.name = fname
			dirview.dirviewOpen()
			messenger.Information(fname, " "+Language.Translate("opened in new Tab"))
		} else {
			d.curView = dirview.tree_view.splitNode.GetNextPrevView(1)
			if tabs[curTab].Views[d.curView].Buf.Modified() {
				messenger.Error(Language.Translate("You need to save view first"))
				return
			}
			tabs[curTab].Views[d.curView].Open(d.LastPath + fname)
			tabs[curTab].Views[d.curView].Buf.name = fname
			messenger.Information(fname, " "+Language.Translate("opened in View"))
			d.curView = 0
		}
		//curTab = 0
	}
}

func (d *DirTreeView) viewDir() {
	var dir string
	var s, p string

	if d.tree_view == nil {
		return
	}
	dir = "🖿   ../"
	d.tree_view.Buf.remove(d.tree_view.Buf.Start(), d.tree_view.Buf.End())
	pos := d.tree_view.Buf.Start()
	files, _ := ioutil.ReadDir(d.LastPath)
	for _, f := range files {
		if f.IsDir() {
			p = "🖿   "
			s = "/"
		} else {
			p = "  🗎 "
			s = ""
		}
		dir = dir + "\n" + p + f.Name() + s
	}
	d.tree_view.Buf.Insert(pos, dir)
	d.tree_view.Cursor.GotoLoc(d.pos)
	//d.tree_view.Cursor.GotoLoc(Loc{0, 0})
	d.tree_view.Relocate()
	d.LineSelect()
}

func (d *DirTreeView) dirviewOpen() {
	if d.tree_view != nil {
		return
	}
	if CurView().Type.Kind != 0 {
		return
	}
	d.curView = 0
	if CurView().splitNode.parent.kind == true {
		CurView().VSplitIndex(NewBufferFromString("", "fileviewer"), 0)
	} else {
		n := CurView().splitNode.GetNextPrevView(0)
		CurView().VSplitIndex(NewBufferFromString("", "fileviewer"), n)
	}
	d.tree_view = CurView()
	d.tree_view.Width = 30
	d.tree_view.LockWidth = true
	d.tree_view.Type.Kind = 2
	d.tree_view.Type.Readonly = true
	d.tree_view.Type.Scratch = true
	SetLocalOption("softwrap", "true", d.tree_view)
	SetLocalOption("ruler", "false", d.tree_view)
	SetLocalOption("autosave", "false", d.tree_view)
	tabs[curTab].Resize()
	d.Open = true
	d.tree_view.Cursor.GotoLoc(Loc{0, 0})
	d.viewDir()
	messenger.Information(d.LastPath)
}

func (d *DirTreeView) dirviewClose() {
	if d.tree_view == nil {
		return
	}
	d.pos = d.tree_view.Cursor.Loc
	d.tree_view.Quit(false)
	messenger.Information("")
}

func (d *DirTreeView) toggleView() {
	if d.active == false {
		return
	}
	if d.Open == true {
		d.dirviewClose()
		messenger.Message("")
	} else {
		d.dirviewOpen()
	}
}

func (d *DirTreeView) onIconClick() {
	d.toggleView()
}

func (d *DirTreeView) onTabChange() {
	d.dirviewClose()
}

func (d *DirTreeView) LineSelect() {
	c := d.tree_view.Cursor
	c.Start()
	c.SetSelectionStart(c.Loc)
	c.End()
	c.SetSelectionEnd(c.Loc)
}

func (d *DirTreeView) onMouseClick(v *View) int {
	if dirview.Open == false {
		// Is not for me
		return 0
	}
	if d.tree_view != CurView() {
		return 0
	}
	if v.tripleClick == false && v.doubleClick == false {
		d.LineSelect()
		return 1
	} else if v.doubleClick == true || v.tripleClick == true {
		d.opeFile(d.tree_view.Buf.Line(v.Cursor.Y))
		return 2
	}
	return 0
}

func (d *DirTreeView) onExecuteAction(action string, v *View) bool {
	if dirview.Open == false {
		// Is not for me
		return false
	}
	if d.tree_view != CurView() {
		d.dirviewClose()
		return false
	}
	if action == "InsertNewline" {
		d.opeFile(d.tree_view.Buf.Line(v.Cursor.Y))
		return true
	} else if action == "openonView" {
		d.curView = 1
		d.LineSelect()

		d.opeFile(d.tree_view.Buf.Line(v.Cursor.Y))
		d.curView = 0
		return true
	} else if strings.Contains(action, "Cursor") {
		d.LineSelect()
		return true
	}
	return false
}
