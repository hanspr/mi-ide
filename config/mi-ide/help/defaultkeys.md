# Default Keys

Below are simple charts of the default hotkeys and their functions. For more
information about binding custom hotkeys or changing default bindings, please
run `> help keybindings`

Please remember that *all* keys here are rebindable! If you don't like it, you
can change it!

Be aware that Alt-A and Alt-a are considered different key combinations, but
Ctrl-A and Ctrl-a are the same key combination, the terminal looks at it as
Ctrl-A

## Function keys, and alternatives

| Key       :| Description of function                  |
|------------|------------------------------------------|
| F1         | Open file                                |
| F2         | Save                                     |
| F3         | Save As ...                              |
| F4         | Close focused window (Tab if empty)      |
| Shift F1   | Open file viewer                         |
| Shift F2   | Save All                                 |
| Shift F4   | Close other windows                      |
| F9-F12     | Keys reserved for plug-ins               |
| Alt-[1-0]  | Keys reserved for plug-ins               |
| Ctrl+P     | Toggle Mouse mode enabled,disabled       |

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
| Ctrl+G                    | Jump to a line in the file (prompts with line #)     |
| Alt+g                     | Find function from word under cursor                 |
| Ctrl+L                    | Center current line vertically in view               |
| Alt+j                     | Move cursor left                                     |
| Alt+k                     | Move cursor down                                     |
| Alt+i                     | Move cursor up                                       |
| Alt+l                     | Move cursor right                                    |
| Alt+u                     | Move cursor to start of line                         |
| Alt+o                     | Move cursor to end of line                           |
| Alt+y                     | Move cursor page up                                  |
| Alt+h                     | Move cursor page down                                |
| Shift-Alt-[ilouyh]        | Select text in corresponding direction               |
| Ctrl-K                    | Select line                                          |
| Ctrl-J                    | Delete line                                          |

## Tabs

| Key   :| Description of function   |
|--------|---------------------------|
| Ctr-T  | Open a new tab            |
| Alt-a  | Previous tab              |
| Alt-s  | Next tab                  |

## Windows (Views)

| Key     :| Description of function             |
|----------|-------------------------------------|
| Alt-:    | Split current window vertically     |
| Alt-_    | Split current window horizontally   |
| Alt-q    | Move cursor to previous view        |
| Alt-w    | Move cursor to next view            |

## Find Operations

| Key             :| Description of function                   |
|------------------|-------------------------------------------|
| Ctrl+F           | Find (open Search Dialog)                 |
| Ctrl+R           | Replace (open Search / Replace Dialog)    |
| Backspace        | Find previous instance of current search  |
| Enter            | Find next instance of current search      |

## File Operations

| Key   :| Description of function                                               |
|--------|-----------------------------------------------------------------------|
| Ctrl+Q | Close All Windows, Tabs, and Exit (asks to save if a buffer is dirty) |
| F1     | Open a file (prompts for filename)                                    |
| F2     | Save current file                                                     |
| F3     | Save As .. current file                                               |
| Ctrl+S | Save current file                                                     |
| Ctrl+W | Close current tab or window                                           |
| Alt-F  | Open file viewer                                                      |

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
| AltUpArrow or Alt+Y               | Move current line or selected lines up    |
| AltDownArrow or Alt+H             | Move current line or selected lines down  |
| AltBackspace or AltCtrl+H         | Delete word left                          |
| Ctrl+A                            | Select all                                |
| Ctrl+K                            | Cut selected line and append to clipboard |
| Ctrl+X                            | Cut selected text                         |
| Ctrl+C                            | Copy selected text                        |
| Ctrl+V                            | Paste                                     |
| Ctrl+D                            | Duplicate current line                    |
| Ctrl+Z                            | Undo                                      |
| Ctrl+Y                            | Redo                                      |
| Ctrl+J                            | Delete line                               |
| Ctrl+T                            | Toggle selection case                     |

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
| Ctrl+E | Open a command prompt for running commands                                |
| Tab    | In command prompt, it will auto complete if available                     |
| Ctrl+B | Run a shell command (this will hide micro-ide while your command executes)|

## Other

| Key   :| Description of function                    |
|--------|--------------------------------------------|
| Alt-?  | Open help file                             |
| Alt-#  | Toggle the line number ruler               |
| Alt-!  | Toggle soft wrap                           |
| Alt-<  | Clear message status line                  |

## Cloud Services

| Key   :| Description of function                   |
|--------|-------------------------------------------|
| Alt-x  | Cut selected text to the cloud            |
| Alt-c  | Copy selected text to the cloud           |
| Alt-v  | Paste from the cloud                      |
| Alt-z  | Transfer current buffer to the cloud      |
| Alt-b  | Download last uploaded buffer from cloud  |
| Alt-+  | Upload or download my saved settings      |
