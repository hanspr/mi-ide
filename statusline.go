package main

import (
	"fmt"
	"path"
	"strconv"

	"github.com/hanspr/tcell/v2"
)

// Statusline represents the information line at the bottom
// of each view
// It gives information such as filename, whether the file has been
// modified, filetype, cursor location
type Statusline struct {
	view    *View
	hotspot map[string]Loc
}

// EncodingSelected Change encoding settings for current buffer
func (sline *Statusline) EncodingSelected(values map[string]string) {
	b := sline.view.Buf
	b.encoding = true
	b.encoder = values["encoding"]
	b.ReOpen()
	if values["encoding"] == "UTF8" {
		b.encoding = false
	} else {
		b.encoding = true
	}
	b.IsModified = true
}

// MouseEvent handle mouse events on the status line
func (sline *Statusline) MouseEvent(e *tcell.EventMouse, rx, ry int) {
	if e.Buttons() != tcell.Button1 {
		return
	}
	for action, hs := range sline.hotspot {
		if rx >= hs.X && rx <= hs.Y {
			if action == "ENCODER" {
				x, _ := e.Position()
				diff := (hs.X + (hs.Y-hs.X+1)/2) - rx
				micromenu.SelEncoding(x+diff, sline.EncodingSelected)
			} else if action == "BUFFERSET" {
				micromenu.SelLocalSettings(sline.view.Buf)
			} else if action == "FILETYPE" {
				x, _ := e.Position()
				diff := (hs.X + (hs.Y-hs.X+1)/2) - rx
				micromenu.SelFileType(x + diff)
			} else if action == "TABSPACE" {
				_, y := e.Position()
				x, _ := e.Position()
				diff := (hs.X + (hs.Y-hs.X+1)/2) - rx
				micromenu.SelTabSpace(x+diff, y)
			} else if action == "FILEFORMAT" {
				if sline.view.Buf.Settings["fileformat"].(string) == "unix" {
					sline.view.Buf.Settings["fileformat"] = "dos"
				} else {
					sline.view.Buf.Settings["fileformat"] = "unix"
				}
			}
		}
	}
}

const fwidth = 60

// Display draws the statusline to the screen
func (sline *Statusline) Display() {
	var size int
	var columnNum, lineNum, totColumn string

	active := true
	offset := 0
	w := sline.view.Width
	showbuttons := true

	if sline.view.Num != CurView().Num {
		active = false
	}

	if sline.view.x != 0 {
		offset++
	}

	// We'll draw the line at the lowest line in the view
	y := sline.view.Height + sline.view.y

	file := sline.view.Buf.AbsPath
	if sline.view.Buf.Settings["basename"].(bool) {
		file = path.Base(file)
	}

	// If the buffer is dirty (has been modified) write a little '*' and no space
	// Is clearer to the eye
	if sline.view.Buf.Modified() && sline.view.Type.Kind == 0 {
		file += "*"
	}

	if fwidth > w/4 {
		size = w / 4
	} else {
		size = fwidth
	}
	if len(file) > size {
		file = ".. " + file[len(file)-size+1:]
	}

	file = fmt.Sprintf("%-"+strconv.Itoa(size)+"s", file)

	// Add one to cursor.x and cursor.y because (0,0) is the top left,
	// but users will be used to (1,1) (first line,first column)
	// We use GetVisualX() here because otherwise we get the column number in runes
	// so a '\t' is only 1, when it should be tabSize
	// Find buffer total lines and add to status

	if active {
		columnNum = strconv.Itoa(sline.view.Cursor.GetVisualX() + 1)
		lineNum = strconv.Itoa(sline.view.Cursor.Y + 1)
	} else {
		columnNum = strconv.Itoa(sline.view.savedLoc.X + 1)
		lineNum = strconv.Itoa(sline.view.savedLoc.Y + 1)
	}
	totColumn = strconv.Itoa(sline.view.Buf.End().Y + 1)

	file += fmt.Sprintf(" %6s/%s,%-4s ", lineNum, totColumn, columnNum)

	// bellow 69 columns begin hidding information
	if w > 69 && mouseEnabled {
		var ff string

		sline.hotspot["BUFFERSET"] = Loc{Count(file) + offset, Count(file) + offset + 2}
		file += " ⎈ "

		// Create hotspots for status line events
		if sline.view.Buf.Settings["indentchar"] == "\t" {
			ff = "Tab "
		} else {
			ff = "Spc "
		}
		ff = fmt.Sprintf("%-4s%-2d▴", ff, int(sline.view.Buf.Settings["tabsize"].(float64)))
		sline.hotspot["TABSPACE"] = Loc{Count(file) + offset, Count(file) + offset + Count(ff) - 1}
		file += ff + " "

		sline.hotspot["FILETYPE"] = Loc{Count(file) + offset - 1, Count(file) + offset + 1 + Count(sline.view.Buf.FileType())}
		file += sline.view.Buf.FileType() + " ▴"

		ff = fmt.Sprintf("%-4s", sline.view.Buf.Settings["fileformat"].(string))
		sline.hotspot["FILEFORMAT"] = Loc{Count(file) + offset, Count(file) + offset + 6}
		file += " " + ff + "   "

		sline.hotspot["ENCODER"] = Loc{Count(file) + offset - 2, Count(file) + offset + 1 + Count(sline.view.Buf.encoder)}
		file += sline.view.Buf.encoder + " ▴"
	} else if !mouseEnabled {
		var ff string

		file += " ⎈ "
		if sline.view.Buf.Settings["indentchar"] == "\t" {
			ff = "tab "
		} else {
			ff = "spc "
		}
		file += fmt.Sprintf("%-4s%-2d", ff, int(sline.view.Buf.Settings["tabsize"].(float64))) + "  "
		file += sline.view.Buf.FileType() + "   "
		file += fmt.Sprintf("%-4s", sline.view.Buf.Settings["fileformat"].(string)) + "   "
		file += sline.view.Buf.encoder + "  "
		showbuttons = false
	} else {
		showbuttons = false
	}

	if size > 10 {
		if sline.view.isOverwriteMode {
			file += " Over"
		} else {
			file += " Ins "
		}
		if sline.view.Type.Readonly || sline.view.Buf.RO {
			file += " (ro) "
		}
	}

	rightText := Version

	statusLineStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["statusline"]; ok {
		statusLineStyle = style
	}
	if !active {
		if style, ok := colorscheme["unfocused"]; ok {
			statusLineStyle = style
		}
	}

	if active && lastOverwriteStatus != sline.view.isOverwriteMode {
		sline.view.SetCursorColorShape()
		lastOverwriteStatus = sline.view.isOverwriteMode
	}

	//Git status
	if git.enabled {
		file = file + "   " + git.status
	}

	// Maybe there is a unicode filename?
	fileRunes := []rune(file)

	viewX := sline.view.x
	if viewX != 0 {
		screen.SetContent(viewX, y, '█', nil, statusLineStyle)
		viewX++
	}
	gstart := 0
	gitbg := git.bgcolor
	gitfg := "#000000"
	for x := 0; x < w; x++ {
		tStyle := statusLineStyle
		if navigationMode {
			tStyle = StringToStyle("#ffffff,#A90000")
		}
		if showbuttons {
			if sline.hotspot["BUFFERSET"].X-offset <= x && x <= sline.hotspot["BUFFERSET"].Y-offset && active {
				tStyle = StringToStyle("#ffffff,#585858")
			} else if sline.hotspot["TABSPACE"].X-offset <= x && x <= sline.hotspot["TABSPACE"].Y-offset && active {
				tStyle = StringToStyle("#ffffff,#444444")
			} else if sline.hotspot["FILETYPE"].X-offset <= x && x <= sline.hotspot["FILETYPE"].Y-offset && active {
				tStyle = StringToStyle("#ffffff,#585858")
			} else if sline.hotspot["FILEFORMAT"].X-offset <= x && x <= sline.hotspot["FILEFORMAT"].Y-offset && active {
				tStyle = StringToStyle("#ffffff,#444444")
			} else if sline.hotspot["ENCODER"].X-offset <= x && x <= sline.hotspot["ENCODER"].Y-offset && active {
				tStyle = StringToStyle("#ffffff,#585858")
			}
		}
		if x < len(fileRunes) {
			if gstart > 0 {
				if fileRunes[x] == '}' {
					gstart = 99
					fileRunes[x] = ' '
					tStyle = StringToStyle(gitfg + "," + gitbg)
				} else if gstart == 1 {
					if fileRunes[x] == '{' {
						gstart = 2
						fileRunes[x] = ' '
						gitbg = "#000000"
						gitfg = "#000000"
					}
					tStyle = StringToStyle(gitfg + "," + gitbg)
				} else if fileRunes[x] != ' ' {
					tStyle = StringToStyle("bold " + git.fgcolor[string(fileRunes[x])] + "," + "#000000")
				} else {
					tStyle = StringToStyle(gitfg + "," + gitbg)
				}
			} else if fileRunes[x] == '[' || fileRunes[x] == ']' {
				gstart = 1
				if fileRunes[x] == ']' {
					gitfg = "#000000"
					gitbg = "gold"
				}
				fileRunes[x] = '↳'
				tStyle = StringToStyle(gitfg + "," + gitbg)
			}
			screen.SetContent(viewX+x, y, fileRunes[x], nil, tStyle)
		} else if x >= w-len(rightText) && x < len(rightText)+w-len(rightText) {
			screen.SetContent(viewX+x, y, []rune(rightText)[x-w+len(rightText)], nil, tStyle)
		} else {
			screen.SetContent(viewX+x, y, ' ', nil, tStyle)
		}
	}
}
