# Default Keys

For a detailed layout visit: https://github.com/hanspr/mi-ide/wiki/defaultkeys

## Function keys, and alternatives

| Key       :| Description of function                  |
|------------|------------------------------------------|
| F1         | Open file                                |
| F2         | Save                                     |
| F3         | Save As ...                              |
| F4         | Close focused window (Tab if empty)      |
| F9-F12     | Keys reserved for plug-ins               |
| Alt-[1-0]  | Keys reserved for plug-ins               |
| Ctrl+k     | Double combination command               |
| Ctrl+k p   | Toggle Mouse mode enabled,disabled       |

## Navigation

| Key                      :| Description of function                              |
|---------------------------|------------------------------------------------------|
| Arrows                    | Move the cursor around                               |
| Shift+arrows              | Move and select text                                 |
| Home or CtrlLeftArrow     | Move to the beginning of the current line            |
| End or CtrlRightArrow     | Move to the end of the current line                  |
| AltLeftArrow              | Move cursor one word left                            |
| AltRightArrow             | Move cursor one word right                           |
| Alt-{                     | Move cursor to end of document                       |
| Alt-}                     | Move cursor to beginning of document                 |
| PageUp                    | Move cursor up one page                              |
| PageDown                  | Move cursor down one page                            |
| CtrlHome or CtrlUpArrow   | Move cursor to start of document                     |
| CtrlEnd or CtrlDownArrow  | Move cursor to end of document                       |
| Ctrl+g                    | Jump to a line in the file (prompts with line #)     |
| Alt+g                     | Find function from word under cursor                 |
| Alt+j                     | Move cursor left                                     |
| Alt+k                     | Move cursor down                                     |
| Alt+i                     | Move cursor up                                       |
| Alt+l                     | Move cursor right                                    |
| Alt+u                     | Move cursor to start of line                         |
| Alt+o                     | Move cursor to end of line                           |
| Alt+y                     | Move cursor page up                                  |
| Alt+h                     | Move cursor page down                                |
| Shift-Alt-[ilouyh]        | Select text in corresponding direction               |
| Ctrl+l                    | Center current line vertically in view               |
| Ctrl-u                    | Delete char under cursor (same as Del)               |
| Ctrl-j                    | Delete line                                          |
| Ctrl-n                    | Enter / Exit Navigation mode (see below)             |

## Navigation mode

Allows you to navigate the buffer in readonly mode, using the same alt keys but without
the need to hold the Alt key.

This is you can navigate the text with the keys : `j k l i`
Using `Shift` with the navigation keys will select text.
All Ctrl keys are availabe for : copy, paste, cut, add line, remove lines, etc.

## Tabs

| Key   :| Description of function   |
|--------|---------------------------|
| Ctr-t  | Open a new tab            |
| Alt-a  | Previous tab              |
| Alt-s  | Next tab                  |

## Views (split windos)

| Key     :| Description of function             |
|----------|-------------------------------------|
| Alt-:    | Split current window vertically     |
| Alt-_    | Split current window horizontally   |
| Alt-q    | Move cursor to previous view        |
| Alt-w    | Move cursor to next view            |

## Find Operations

| Key             :| Description of function                   |
|------------------|-------------------------------------------|
| Ctrl+f           | Find (open Search Dialog)                 |
| Ctrl+r           | Replace (open Search / Replace Dialog)    |
| Backspace        | Find previous instance of current search  |
| Enter            | Find next instance of current search      |

## File Operations

| Key     :| Description of function                                               |
|----------|-----------------------------------------------------------------------|
| Ctrl+q   | Close All Windows, Tabs, and Exit (asks to save if a buffer is dirty) |
| F1       | Open a file (prompts for filename)                                    |
| F2       | Save current file                                                     |
| F3       | Save As .. current file                                               |
| Ctrl+s   | Save current file                                                     |
| Ctrl+w   | Close current tab or window                                           |
| Ctrl+b f | Open file viewer                                                      |

## Text operations

| Key                              :| Description of function                   |
|-----------------------------------|-------------------------------------------|
| Shift+arrows                      | Move and select text                      |
| AltShiftRightArrow                | Select word right                         |
| AltShiftLeftArrow                 | Select word left                          |
| ShiftHome ,  CtrlShiftLeftArrow   | Select to start of current line           |
| ShiftEnd ,  CtrlShiftRightArrow   | Select to end of current line             |
| AltShiftUpArrow                   | Select one Page Up                        |
| AltShiftDownArrow                 | Select one Page Down                      |
| AltUpArrow   or Alt+e             | Move current line or selected lines up    |
| AltDownArrow or Alt+d             | Move current line or selected lines down  |
| Ctrl+b a                          | Select all                                |
| Ctrl+k                            | Cut selected line and append to clipboard |
| Ctrl+x                            | Cut selected text                         |
| Ctrl+c                            | Copy selected text                        |
| Ctrl+v                            | Paste                                     |
| Ctrl+d                            | Duplicate current line                    |
| Ctrl+z                            | Undo                                      |
| Alt+z                             | Redo                                      |
| Ctrl+j                            | Delete line                               |
| Ctrl+c                            | Toggle selection case                     |

## Multiple cursors

| Key           :| Description of function                                               |
|----------------|-----------------------------------------------------------------------|
| Alt-n          | Create new multiple cursor from selection                             |
| Alt-,          | Remove latest multiple cursor                                         |
| Alt-.          | Remove all multiple cursors (cancel)                                  |
| Alt-;          | Skip multiple cursor selection                                        |
| Alt-m          | Spawn a new cursor at the beginning of every line in current selection|
| Ctrl-MouseLeft | Place a multiple cursor at any location                               |

## Power user

| Key   :| Description of function                                                   |
|--------|---------------------------------------------------------------------------|
| Ctrl+e | Open a command prompt for running commands                                |
| Tab    | In command prompt, it will auto complete if available                     |

## Other

| Key   :| Description of function                    |
|--------|--------------------------------------------|
| Alt-?  | Open help file                             |
| Alt-#  | Toggle the line number ruler               |
| Alt-!  | Toggle soft wrap                           |
| Alt-<  | Clear message status line                  |

## Cloud Services

| Key     :| Description of function                   |
|----------|-------------------------------------------|
| Ctrl+k x | Cut selected text to the cloud            |
| Ctrl+k c | Copy selected text to the cloud           |
| Ctrl+k v | Paste from the cloud                      |
| Ctrl+k z | Transfer current buffer to the cloud      |
| Ctrl+k b | Download last uploaded buffer from cloud  |
| Ctrl+k + | Upload or download settings               |


