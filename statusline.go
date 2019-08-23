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
			} else if action == "FILETYPE" {
				x, _ := e.Position()
				diff := (hs.X + (hs.Y-hs.X+1)/2) - rx + 2
				micromenu.SelFileType(x + diff)
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

	fvstyle := true
	offset := 0
	_, h := screen.Size()
	w := sline.view.Width

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

	// Fix length of strings to help eyes find information in the same place
	if 40 > w/4 {
		size = w / 4
		if len(file) > size {
			file = ".. " + file[len(file)-size+1:]
		}
	} else {
		size = 40
	}

	if w > 80 || dirview.Open {
		// If the screen is wide enough we can offer the treeview
		if y >= h-3 && sline.view.x < 2 {
			if dirview.Open {
				file = fmt.Sprintf("<< %-"+strconv.Itoa(size)+"s", file)
				fvstyle = true
			} else {
				file = fmt.Sprintf(">> %-"+strconv.Itoa(size)+"s", file)
				fvstyle = true
			}
		} else {
			file = fmt.Sprintf("   %-"+strconv.Itoa(size)+"s", file)
			fvstyle = false
			dirview.active = false
		}
		dirview.active = true
	} else {
		// Below 80 columns, treeview fails there is not enough room
		file = fmt.Sprintf("   %-"+strconv.Itoa(size)+"s", file)
		fvstyle = false
		dirview.active = false
	}

	// Add one to cursor.x and cursor.y because (0,0) is the top left,
	// but users will be used to (1,1) (first line,first column)
	// We use GetVisualX() here because otherwise we get the column number in runes
	// so a '\t' is only 1, when it should be tabSize
	columnNum := strconv.Itoa(sline.view.Cursor.GetVisualX() + 1)
	lineNum := strconv.Itoa(sline.view.Cursor.Y + 1)

	// Find buffer total lines and add to status
	totColumn := strconv.Itoa(sline.view.Buf.End().Y + 1)

	// Fix size of cursor position so it stays in the same place all the time
	file += fmt.Sprintf(" %6s/%s,%-4s ", lineNum, totColumn, columnNum)

	// bellow 66 colomns begin hidding information even more
	if w > 65 {
		if sline.view.x != 0 {
			offset++
		}

		// Create hotspots for status line events
		sline.hotspot["FILETYPE"] = Loc{Count(file) + 1 + offset, Count(file) + offset + 1 + Count(sline.view.Buf.FileType())}
		file += " " + sline.view.Buf.FileType() + "▲"

		if size > 12 {
			sline.hotspot["FILEFORMAT"] = Loc{Count(file) + 2 + offset, Count(file) + offset + 3 + Count(sline.view.Buf.Settings["fileformat"].(string))}
			file += "  (" + sline.view.Buf.Settings["fileformat"].(string) + ")  "
		}
		sline.hotspot["ENCODER"] = Loc{Count(file) + 1 + offset, Count(file) + offset + Count(sline.view.Buf.encoder)}
		file += " " + sline.view.Buf.encoder

		if size > 12 {
			if sline.view.isOverwriteMode {
				file += "  Over"
			} else {
				file += "  Inst"
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
	if style, ok := colorscheme["unfocused"]; ok {
		if sline.view.Num != CurView().Num {
			statusLineStyle = style
		}
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
		if x < 3 && fvstyle {
			tStyle = StringToStyle("#ffd700,#5f87af")
		} else if w > 65 && sline.hotspot["ENCODER"].X-offset <= x && x <= sline.hotspot["ENCODER"].Y-offset && sline.view.Num == CurView().Num {
			tStyle = StringToStyle("#ffd700,#585858")
		} else if w > 65 && sline.hotspot["FILETYPE"].X-offset <= x && x <= sline.hotspot["FILETYPE"].Y-offset && sline.view.Num == CurView().Num {
			tStyle = StringToStyle("#ffd700,#585858")
		} else if w > 65 && sline.hotspot["FILEFORMAT"].X-offset <= x && x <= sline.hotspot["FILEFORMAT"].Y-offset && sline.view.Num == CurView().Num {
			tStyle = StringToStyle("#ffd700,#585858")
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
