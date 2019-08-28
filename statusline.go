package main

import (
	"fmt"
	"path"
	"strconv"

	"github.com/hanspr/tcell"
)

// Statusline represents the information line at the bottom
// of each view
// It gives information such as filename, whether the file has been
// modified, filetype, cursor location
type Statusline struct {
	view    *View
	hotspot map[string]Loc
}

func (sline *Statusline) EncodingSelected(values map[string]string) {
	b := sline.view.Buf
	b.encoding = true
	b.encoder = values["encoding"]
	b.ReOpen()
	if values["encoding"] == "UTF-8" {
		b.encoding = false
	} else {
		b.encoding = true
	}
	b.IsModified = true
}

func (sline *Statusline) MouseEvent(e *tcell.EventMouse, rx, ry int) {
	if e.Buttons() != tcell.Button1 || e.HasMotion() == true {
		return
	}
	for action, hs := range sline.hotspot {
		if rx >= hs.X && rx <= hs.Y {
			if action == "ENCODER" {
				micromenu.SelEncoding(sline.view.Buf.encoder, sline.EncodingSelected)
			} else if action == "BUFERSET" {
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
	return
}

// Display draws the statusline to the screen
func (sline *Statusline) Display() {
	var size int
	var columnNum, lineNum, totColumn string

	active := true
	offset := 0
	w := sline.view.Width

	if sline.view.Num != CurView().Num {
		active = false
	}

	// We'll draw the line at the lowest line in the view
	y := sline.view.Height + sline.view.y

	file := sline.view.Buf.GetName()
	if sline.view.Buf.Settings["basename"].(bool) {
		file = path.Base(file)
	}

	// If the buffer is dirty (has been modified) write a little '*' and no space
	// Is clearer to the eye
	if sline.view.Buf.Modified() && sline.view.Type.Kind == 0 {
		file += "*"
	}

	// Fix length of file name to help eyes find information in the same place
	if 30 > w/4 {
		size = w / 4
		if len(file) > size {
			file = ".. " + file[len(file)-size+1:]
		}
	} else {
		size = 30
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

	// Fix size of cursor position so it stays in the same place all the time
	file += fmt.Sprintf(" %6s/%s,%-4s ", lineNum, totColumn, columnNum)

	sline.hotspot["BUFERSET"] = Loc{Count(file) + offset, Count(file) + offset + 2}
	file += " ⎈ "

	// bellow 66 columns begin hidding information
	if w > 55 {
		var ff string

		if sline.view.x != 0 {
			offset++
		}

		// Create hotspots for status line events
		if sline.view.Buf.Settings["indentchar"] == "\t" {
			ff = "Tab"
		} else {
			ff = "Spc"
		}
		ff = fmt.Sprintf("%-4s%-2d▲", ff, int(sline.view.Buf.Settings["tabsize"].(float64)))
		sline.hotspot["TABSPACE"] = Loc{Count(file) + offset, Count(file) + offset + Count(ff) - 1}
		file += ff + " "

		sline.hotspot["FILETYPE"] = Loc{Count(file) + offset - 1, Count(file) + offset + 1 + Count(sline.view.Buf.FileType())}
		file += sline.view.Buf.FileType() + " ▲"

		if size > 12 {
			ff = fmt.Sprintf("%-4s", sline.view.Buf.Settings["fileformat"].(string))
			sline.hotspot["FILEFORMAT"] = Loc{Count(file) + offset, Count(file) + offset + 5}
			file += " " + ff + " "

			sline.hotspot["ENCODER"] = Loc{Count(file) + offset - 1, Count(file) + offset + Count(sline.view.Buf.encoder) + 1}
			file += " " + sline.view.Buf.encoder + " "
		}

		if size > 12 {
			if sline.view.isOverwriteMode {
				file += " Over"
			} else {
				file += " Ins "
			}
		}
		if size > 12 {
			if sline.view.Type.Readonly == true {
				file += " ro"
			}
		}
	}
	rightText := ""

	statusLineStyle := defStyle.Reverse(true)
	if style, ok := colorscheme["statusline"]; ok {
		statusLineStyle = style
	}
	if active == false {
		if style, ok := colorscheme["unfocused"]; ok {
			statusLineStyle = style
		}
	}

	if active && LastOverwriteStatus != sline.view.isOverwriteMode {
		sline.view.SetCursorColorShape()
		LastOverwriteStatus = sline.view.isOverwriteMode
	}

	// Maybe there is a unicode filename?
	fileRunes := []rune(file)

	if sline.view.Type == vtTerm {
		fileRunes = []rune(sline.view.term.title)
		rightText = ""
	}

	viewX := sline.view.x
	if viewX != 0 {
		screen.SetContent(viewX, y, '█', nil, statusLineStyle)
		viewX++
	}
	for x := 0; x < w; x++ {
		tStyle := statusLineStyle
		if w > 65 && sline.hotspot["BUFERSET"].X-offset <= x && x <= sline.hotspot["BUFERSET"].Y-offset && active {
			tStyle = StringToStyle("#ffffff,#4e4e4e")
		} else if w > 65 && sline.hotspot["TABSPACE"].X-offset <= x && x <= sline.hotspot["TABSPACE"].Y-offset && active {
			tStyle = StringToStyle("#ffffff,#444444")
		} else if w > 65 && sline.hotspot["FILETYPE"].X-offset <= x && x <= sline.hotspot["FILETYPE"].Y-offset && active {
			tStyle = StringToStyle("#ffffff,#4e4e4e")
		} else if w > 65 && sline.hotspot["FILEFORMAT"].X-offset <= x && x <= sline.hotspot["FILEFORMAT"].Y-offset && active {
			tStyle = StringToStyle("#ffffff,#444444")
		} else if w > 65 && sline.hotspot["ENCODER"].X-offset <= x && x <= sline.hotspot["ENCODER"].Y-offset && active {
			tStyle = StringToStyle("#ffffff,#4e4e4e")
		}
		if x < len(fileRunes) {
			screen.SetContent(viewX+x, y, fileRunes[x], nil, tStyle)
		} else if x >= w-len(rightText) && x < len(rightText)+w-len(rightText) {
			screen.SetContent(viewX+x, y, []rune(rightText)[x-w+len(rightText)], nil, tStyle)
		} else {
			screen.SetContent(viewX+x, y, ' ', nil, tStyle)
		}
	}
}
