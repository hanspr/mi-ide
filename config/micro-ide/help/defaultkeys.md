# Default Keys

Below are simple charts of the default hotkeys and their functions. For more
information about binding custom hotkeys or changing default bindings, please
run `> help keybindings`

Please remember that *all* keys here are rebindable! If you don't like it, you
can change it!

### Function keys, and alternatives

| Key        | Description of function                  |
|------------|------------------------------------------|
| F1 , Alt-1 | Open file                                |
| F2 , Alt-2 | Save                                     |
| F3 , Alt-3 | Save As ...                              |
| F4 , Alt-4 | Close focused window (Tab if empty)      |
| F5         | Move to previous window                  |
| F6         | Move to next window                      |
| F7         | Move to previous Tab                     |
| F8         | Move to next Tab                         |
| F9-F12     | Keys reserved for plug-ins               |
| Alt-[5-0]  | Keys reserved for plug-ins               |

### Navigation

| Key                       | Description of function                              |
|-------------------------- |----------------------------------------------------- |
| Arrows                    | Move the cursor around                               |
| Shift+arrows              | Move and select text                                 |
| Home or CtrlLeftArrow     | Move to the beginning of the current line            |
| End or CtrlRightArrow     | Move to the end of the current line                  |
| AltLeftArrow              | Move cursor one word left                            |
| AltRightArrow             | Move cursor one word right                           |
| Alt+{                     | Move cursor to previous empty line                   |
| Alt+}                     | Move cursor to next empty line, or end of document   |
| PageUp                    | Move cursor up one page                              |
| PageDown                  | Move cursor down one page                            |
| CtrlHome or CtrlUpArrow   | Move cursor to start of document                     |
| CtrlEnd or CtrlDownArrow  | Move cursor to end of document                       |
| Ctrl+G                    | Jump to a line in the file (prompts with line #)     |
| Ctrl+L                    | Center line vertically in view                       |

### Tabs

| Key     | Description of function   |
|-------- |-------------------------  |
| Ctrl+T  | Open a new tab            |
| F7      | Previous tab              |
| F8      | Next tab                  |

### Windows (Views)

| Key     | Description of function           |
|-------- |-----------------------------------|
| Alt-V   | Split current window vertically   |
| Alt-H   | Split current window horizontally |

### Find Operations

| Key              | Description of function                   |
|------------------|-------------------------------------------|
| Ctrl+F           | Find (opens Search Dialog)                |
| Ctrl+R           | Replace (opens Search / Replace Dialog)   |
| F5 , Backspace   | Find previous instance of current search  |
| F6 , Enter       | Find next instance of current search      |

### File Operations

| Key    | Description of function                                               |
|--------|-----------------------------------------------------------------------|
| Ctrl+Q | Close All Windows, Tabs, and Exit (asks to save if a buffer is dirty) |
| F1     | Open a file (prompts for filename)                                    |
| F2     | Save current file                                                     |
| F3     | Save As .. current file                                               |
| Ctrl+S | Save current file                                                     |

### Text operations

| Key                               | Description of function                   |
|-----------------------------------|------------------------------------------ |
| AltShiftRightArrow                | Select word right                         |
| AltShiftLeftArrow                 | Select word left                          |
| ShiftHome or CtrlShiftLeftArrow   | Select to start of current line           |
| ShiftEnd or CtrlShiftRightArrow   | Select to end of current line             |
| CtrlShiftUpArrow                  | Select to start of file                   |
| CtrlShiftDownArrow                | Select to end of file                     |
| Ctrl+K                            | Cut selected line and append to clipboard |
| Ctrl+X                            | Cut selected text                         |
| Ctrl+C                            | Copy selected text                        |
| Ctrl+V                            | Paste                                     |
| Ctrl+K                            | Cut current line                          |
| Ctrl+D                            | Duplicate current line                    |
| Ctrl+Z                            | Undo                                      |
| Ctrl+Y                            | Redo                                      |
| AltUpArrow                        | Move current line or selected lines up    |
| AltDownArrow                      | Move current line of selected lines down  |
| AltBackspace or AltCtrl+H         | Delete word left                          |
| Alt+L                             | Delete line                               |
| Ctrl+A                            | Select all                                |

### Power user

| Key    | Description of function                                                   |
|--------|---------------------------------------------------------------------------|
| Ctrl+E | Open a command prompt for running commands                                |
| Tab    | In command prompt, it will auto complete if available                     |
| Ctrl+B | Run a shell command (this will hide micro-ide while your command executes)|

### Multiple cursors

| Key            | Description of function                                               |
|----------------|---------------------------------------------------------------------- |
| Alt+N          | Create new multiple cursor from selection                             |
| Alt+P          | Remove latest multiple cursor                                         |
| Alt+C          | Remove all multiple cursors (cancel)                                  |
| Alt+X          | Skip multiple cursor selection                                        |
| Alt+M          | Spawn a new cursor at the beginning of every line in current selection|
| Ctrl-MouseLeft | Place a multiple cursor at any location                               |

### Other

| Key    | Description of function                    |
|--------|--------------------------------------------|
| Alt+G  | Open help file                             |
| Ctrl+H | Backspace                                  |
| Ctrl+R | Toggle the line number ruler               |
| Alt+!  | Toggle soft wrap                           |

### Macros

| Key    | Description of function                                              |
|--------|----------------------------------------------------------------------|
| Ctrl+U | Toggle macro recording  (on / off)                                   |
| Ctrl+J | Run latest recorded macro                                            |

### Emacs style actions

| Key       | Description of function   |
|-----------|-------------------------  |
| Alt+F     | Next word                 |
| Alt+B     | Previous word             |
| Alt+A     | Move to start of line     |
| Alt+E     | Move to end of line       |
