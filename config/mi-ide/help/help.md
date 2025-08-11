# mi-ide help text

Thank you for downloading and using mi-ide.

mi-ide is a terminal-based text editor that aims to be easy to use and intuitive,
while also taking advantage of the full capabilities of modern terminals.

For a list of the default keybindings press CtrlE and type `help defaultkeys`.

See the next section for more information about documentation and help.

## Quick-start

Press CtrlQ to quit, and CtrlS to save. Press CtrlE to start typing commands and
you can see which commands are available by pressing tab, or by viewing the help
topic `> help commands`. When I write `> ...` I mean press CtrlE and then type
whatever is there.

Move the cursor around with the mouse or the arrow keys, or the predefiened alternate
keys : Alt-(jkli) [left,down,right,up]. Type `> help defaultkeys` to  get a quick,
easy overview of the default hotkeys and what they do.

`In Navigation` mode the window has a red frame, and you can use the navigation keys
without pressing Alt, in this case just use :
* jkli (left,down,right,up)
* up (page up, page down)

All read buffers are set to Navigation mode, to change from Navitation Mode to normal
mode, type : Ctrl-N

If the color theme doesn't look good, you can change it with
`> set colorscheme ...`. You can press tab to see the available colorschemes, or
see more information with `> help colors`.

Press Alt-Shift-(qw) to move between splits, and type `> vsplit filename` or
`> hsplit filename` to open a new split.

Press Alt-Shift-(as) to move between tabs.

## Accessing more help

To use it, press Alt-? or CtrlE to access command mode and type in `help` followed by a
topic. Typing `help` followed by nothing will open this page.

Here are the possible help topics that you can read:

* defaultkeys: Gives a more straight-forward list of the hotkey commands and what
  they do.
* commands: Gives a list of all the commands and what they do
* options: Gives a list of all the options you can customize
* plugins: Explains how mi-ide's plugin system works and how to create your own
  plugins
* colors: Explains mi-ide's colorscheme and syntax highlighting engine and how to
  create your own colorschemes or add new languages to the engine

For example, to open the help page on plugins you would press CtrlE and type
`help plugins`.
