package main

import (
	"strconv"
)

// SplitType specifies whether a split is horizontal or vertical
type SplitType bool

const (
	// VerticalSplit type
	VerticalSplit = false
	// HorizontalSplit type
	HorizontalSplit = true
)

// A Node on the split tree
type Node interface {
	VSplit(buf *Buffer, splitIndex int)
	HSplit(buf *Buffer, splitIndex int)
	String() string
}

// A LeafNode is an actual split so it contains a view
type LeafNode struct {
	view *View

	parent *SplitTree
}

// NewLeafNode returns a new leaf node containing the given view
func NewLeafNode(v *View, parent *SplitTree) *LeafNode {
	n := new(LeafNode)
	n.view = v
	n.view.splitNode = n
	n.parent = parent
	return n
}

// A SplitTree is a Node itself and it contains other nodes
type SplitTree struct {
	kind SplitType

	parent   *SplitTree
	children []Node

	x int
	y int

	width      int
	height     int
	lockWidth  bool
	lockHeight bool

	tabNum int
}

var viewIndex = make(map[int]int)
var viewPos = make(map[string]int)
var posView = make(map[string]int)

// VSplit creates a vertical split
func (l *LeafNode) VSplit(buf *Buffer, splitIndex int) {
	w, _ := screen.Size()
	if w < 80 || l.view.Width < 80 {
		messenger.Error("There is not enough Window Width to split properlly")
		return
	}
	if splitIndex < 0 {
		splitIndex = 0
	}

	tab := tabs[l.parent.tabNum]
	if l.parent.kind == VerticalSplit {
		if splitIndex > len(l.parent.children) {
			splitIndex = len(l.parent.children)
		}

		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum

		l.parent.children = append(l.parent.children, nil)
		copy(l.parent.children[splitIndex+1:], l.parent.children[splitIndex:])
		l.parent.children[splitIndex] = NewLeafNode(newView, l.parent)

		tab.Views = append(tab.Views, nil)
		copy(tab.Views[splitIndex+1:], tab.Views[splitIndex:])
		tab.Views[splitIndex] = newView

		tab.CurView = splitIndex
		newView.savedLoc = Loc{0, 0}
		newView.savedLine = buf.Line(0)
	} else {
		if splitIndex > 1 {
			splitIndex = 1
		}

		s := new(SplitTree)
		s.kind = VerticalSplit
		s.parent = l.parent
		s.tabNum = l.parent.tabNum
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum

		if splitIndex == 1 {
			s.children = []Node{l, NewLeafNode(newView, s)}
		} else {
			s.children = []Node{NewLeafNode(newView, s), l}
		}
		l.parent.children[search(l.parent.children, l)] = s
		l.parent = s

		tab.Views = append(tab.Views, nil)
		copy(tab.Views[splitIndex+1:], tab.Views[splitIndex:])
		tab.Views[splitIndex] = newView

		tab.CurView = splitIndex
		newView.savedLoc = Loc{0, 0}
		newView.savedLine = buf.Line(0)
	}

	tab.Resize()
}

// HSplit creates a horizontal split
func (l *LeafNode) HSplit(buf *Buffer, splitIndex int) {
	h, _ := screen.Size()
	if h < 40 || l.view.Height < 40 {
		messenger.Error("There is not enough Window Height to split properlly")
		return
	}

	if splitIndex < 0 {
		splitIndex = 0
	}

	tab := tabs[l.parent.tabNum]
	if l.parent.kind == HorizontalSplit {
		if splitIndex > len(l.parent.children) {
			splitIndex = len(l.parent.children)
		}

		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum

		l.parent.children = append(l.parent.children, nil)
		copy(l.parent.children[splitIndex+1:], l.parent.children[splitIndex:])
		l.parent.children[splitIndex] = NewLeafNode(newView, l.parent)

		tab.Views = append(tab.Views, nil)
		copy(tab.Views[splitIndex+1:], tab.Views[splitIndex:])
		tab.Views[splitIndex] = newView

		tab.CurView = splitIndex
		newView.savedLoc = Loc{0, 0}
		newView.savedLine = buf.Line(0)
	} else {
		if splitIndex > 1 {
			splitIndex = 1
		}

		s := new(SplitTree)
		s.kind = HorizontalSplit
		s.tabNum = l.parent.tabNum
		s.parent = l.parent
		newView := NewView(buf)
		newView.TabNum = l.parent.tabNum
		newView.Num = len(tab.Views)
		if splitIndex == 1 {
			s.children = []Node{l, NewLeafNode(newView, s)}
		} else {
			s.children = []Node{NewLeafNode(newView, s), l}
		}
		l.parent.children[search(l.parent.children, l)] = s
		l.parent = s

		tab.Views = append(tab.Views, nil)
		copy(tab.Views[splitIndex+1:], tab.Views[splitIndex:])
		tab.Views[splitIndex] = newView

		tab.CurView = splitIndex
		newView.savedLoc = Loc{0, 0}
		newView.savedLine = buf.Line(0)
	}

	tab.Resize()
}

// Delete deletes a split
func (l *LeafNode) Delete() {
	i := search(l.parent.children, l)

	copy(l.parent.children[i:], l.parent.children[i+1:])
	l.parent.children[len(l.parent.children)-1] = nil
	l.parent.children = l.parent.children[:len(l.parent.children)-1]

	// Changed to use curTab, instead of l.parent.tabNum.
	// curTab garanties that the Tab that is viewed is the one that we are working on
	tab := tabs[curTab]
	j := findView(tab.Views, l.view)
	copy(tab.Views[j:], tab.Views[j+1:])
	tab.Views[len(tab.Views)-1] = nil // or the zero value of T
	tab.Views = tab.Views[:len(tab.Views)-1]

	for i, v := range tab.Views {
		v.Num = i
	}
	//	if tab.CurView > 0 {
	//		tab.CurView--
	//	}
}

// Cleanup rearranges all the parents after a split has been deleted
func (s *SplitTree) Cleanup() {
	for i, node := range s.children {
		if n, ok := node.(*SplitTree); ok {
			if len(n.children) == 1 {
				if child, ok := n.children[0].(*LeafNode); ok {
					s.children[i] = child
					child.parent = s
					continue
				}
			}
			n.Cleanup()
		}
	}
}

// ResizeSplits resizes all the splits correctly
func (s *SplitTree) ResizeSplits() {
	reminder := 0
	extrarow := 0
	lockedWidth := 0
	lockedHeight := 0
	lockedChildren := 0
	for _, node := range s.children {
		if n, ok := node.(*LeafNode); ok {
			if s.kind == VerticalSplit {
				if n.view.LockWidth {
					lockedWidth += n.view.Width
					lockedChildren++
				}
			} else {
				if n.view.LockHeight {
					lockedHeight += n.view.Height
					lockedChildren++
				}
			}
		} else if n, ok := node.(*SplitTree); ok {
			if s.kind == VerticalSplit {
				if n.lockWidth {
					lockedWidth += n.width
					lockedChildren++
				}
			} else {
				if n.lockHeight {
					lockedHeight += n.height
					lockedChildren++
				}
			}
		}
	}
	x, y := 0, 0
	for K, node := range s.children {
		if n, ok := node.(*LeafNode); ok {
			if s.kind == VerticalSplit {
				if !n.view.LockWidth {
					n.view.Width = (s.width - lockedWidth) / (len(s.children) - lockedChildren)
				}
				n.view.Height = s.height

				n.view.x = s.x + x
				n.view.y = s.y
				x += n.view.Width
			} else {
				if !n.view.LockHeight {
					if K == 0 {
						reminder = (s.height - lockedHeight) % (len(s.children) - lockedChildren)
					}
					if reminder > 0 {
						extrarow = 1
						reminder--
					} else {
						extrarow = 0
					}
					n.view.Height = (s.height-lockedHeight)/(len(s.children)-lockedChildren) + extrarow
				}
				n.view.Width = s.width

				n.view.y = s.y + y
				n.view.x = s.x
				y += n.view.Height
			}
			if n.view.Buf.Settings["statusline"].(bool) {
				n.view.Height--
			}
			n.view.ToggleTabbar()
		} else if n, ok := node.(*SplitTree); ok {
			if s.kind == VerticalSplit {
				if !n.lockWidth {
					n.width = (s.width - lockedWidth) / (len(s.children) - lockedChildren)
				}
				n.height = s.height

				n.x = s.x + x
				n.y = s.y
				x += n.width
			} else {
				if !n.lockHeight {
					if K == 0 {
						reminder = (s.height - lockedHeight) % (len(s.children) - lockedChildren)
					}
					if reminder > 0 {
						extrarow = 1
						reminder--
					} else {
						extrarow = 0
					}
					n.height = (s.height-lockedHeight)/(len(s.children)-lockedChildren) + extrarow
				}
				n.width = s.width

				n.y = s.y + y
				n.x = s.x
				y += n.height
			}
			n.ResizeSplits()
		}
	}
}

func (l *LeafNode) String() string {
	return l.view.Buf.GetName()
}

func search(haystack []Node, needle Node) int {
	for i, x := range haystack {
		if x == needle {
			return i
		}
	}
	return 0
}

func findView(haystack []*View, needle *View) int {
	for i, x := range haystack {
		if x == needle {
			return i
		}
	}
	return 0
}

// VSplit is here just to make SplitTree fit the Node interface
func (s *SplitTree) VSplit(buf *Buffer, splitIndex int) {}

// HSplit is here just to make SplitTree fit the Node interface
func (s *SplitTree) HSplit(buf *Buffer, splitIndex int) {}

func (s *SplitTree) String() string {
	str := "["
	for _, child := range s.children {
		str += child.String() + ", "
	}
	return str + "]"
}

// New routines 2019-05-21

func getTopSplitTree(s *SplitTree) *SplitTree {
	// Find top most parent
	if s.parent == nil {
		return s
	}
	return getTopSplitTree(s.parent)
}

func viewNum_Map(s *SplitTree) {
	for _, node := range s.children {
		if n, ok := node.(*LeafNode); ok {
			// Set the position of this view in the tree
			// Key is a string : currenttab-index
			// To avoid creating a map of maps
			posView[strconv.Itoa(curTab)+"-"+strconv.Itoa(viewIndex[curTab])] = n.view.Num
			// Save where this view is stored in the hash
			// Key is a string : currenttab-view.Num
			// To avoid creating a map of maps
			viewPos[strconv.Itoa(curTab)+"-"+strconv.Itoa(n.view.Num)] = viewIndex[curTab]
			// Increase the index
			viewIndex[curTab]++
		} else if n, ok := node.(*SplitTree); ok {
			viewNum_Map(n)
		}
	}
}

func loadNumMap(l *LeafNode) {
	for k := range posView {
		delete(posView, k)
	}
	for k := range viewPos {
		delete(viewPos, k)
	}
	pp := getTopSplitTree(l.parent)
	viewIndex[curTab] = 0
	viewNum_Map(pp)
}

func (l *LeafNode) GetNextPrevView(direction int) int {
	// Reload only if the ammount fo views has changed
	if viewIndex[curTab] != len(tabs[curTab].Views) {
		loadNumMap(l)
	}
	x := l.view.Num
	vp := viewPos[strconv.Itoa(curTab)+"-"+strconv.Itoa(x)]
	// check current position - direction is not negative
	if vp+direction < 0 {
		// We are at the begining, do not move. Return same view number
		return x
		// check current position + direction is not after last index
	} else if vp+direction >= viewIndex[curTab] {
		// We are at the end, do not move. Return same view number
		return x
	}
	// Return the view number, for the previous or next position in map
	return posView[strconv.Itoa(curTab)+"-"+strconv.Itoa(vp+direction)]
}

func (l *LeafNode) GetViewNumPosition(x int) int {
	loadNumMap(l)
	return viewPos[strconv.Itoa(curTab)+"-"+strconv.Itoa(x)]
}

func (l *LeafNode) GetPositionViewNum(vp int) int {
	loadNumMap(l)
	if vp >= viewIndex[curTab] {
		vp = viewIndex[curTab] - 1
	}
	return posView[strconv.Itoa(curTab)+"-"+strconv.Itoa(vp)]
}
