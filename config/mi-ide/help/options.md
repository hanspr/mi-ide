# Options

Micro-ide stores all of the user configuration in its configuration directory.

Micro-ide uses the `$XDG_CONFIG_HOME/mi-ide` as the configuration directory. As per
the XDG spec, if `$XDG_CONFIG_HOME` is not set, `~/.config/mi-ide` is used as
the config directory.

Here are the options that you can set:

* `autoindent`: when creating a new line use the same indentation as the
   previous line.

	default value: `true`

* `autosave`: mi-ide will save the buffer every 8 seconds automatically. mi-ide
   also will automatically save and quit when you exit without asking. Be
   careful when using this feature, because you might accidentally save a file,
   overwriting what was there before.

	default value: `false`

* `basename`: in the infobar, show only the basename of the file being edited
   rather than the full path.

    default value: `false`

* `colorcolumn`: if this is not set to 0, it will display a column at the
  specified column. This is useful if you want column 80 to be highlighted
  special for example.

	default value: `0`

* `colorscheme`: loads the colorscheme stored in
   $(configDir)/colorschemes/`option`.mi-ide, This setting is `global only`.

	default value: `default`

	Note that the default colorschemes (default, solarized, and solarized-tc)
	are not located in configDir, because they are embedded in the mi-ide binary.

	The colorscheme can be selected from all the files in the
	~/.config/mi-ide/colorschemes/ directory. mi-ide comes by default with three
	colorschemes:

	You can read more about mi-ide's colorschemes in the `colors` help topic
	(`help colors`).

* `cursorline`: highlight the line that the cursor is on in a different color
   (the color is defined by the colorscheme you are using).

	default value: `true`

* `eofnewline`: mi-ide will automatically add a newline to the file.

	default value: `false`

* `fastdirty`: this determines what kind of algorithm mi-ide uses to determine if
   a buffer is modified or not. When `fastdirty` is on, mi-ide just uses a
   boolean `modified` that is set to `true` as soon as the user makes an edit.
   This is fast, but can be inaccurate. If `fastdirty` is off, then mi-ide will
   hash the current buffer against a hash of the original file (created when the
   buffer was loaded). This is more accurate but obviously more resource
   intensive. This option is only for people who really care about having
   accurate modified status.

	default value: `true`

* `fileformat`: this determines what kind of line endings mi-ide will use for the
   file. UNIX line endings are just `\n` (lf) whereas dos line endings are
   `\r\n` (crlf). The two possible values for this option are `unix` and `dos`.
   The fileformat will be automatically detected and displayed on the statusline
   but this option is useful if you would like to change the line endings or if
   you are starting a new file.

	default value: `unix`

* `filetype`: sets the filetype for the current buffer. This setting is
   `local only`.

	default value: this will be automatically set depending on the file you have
	open

* `indentchar`: sets the indentation character.

	default value: ` `

* `keepautoindent`: when using autoindent, whitespace is added for you. This
   option determines if when you move to the next line without any insertions
   the whitespace that was added should be deleted. By default the autoindent
   whitespace is deleted if the line was left empty.

	default value: `false`

* `mouse`: whether to enable mouse support. When mouse support is disabled,
   usually the terminal will be able to access mouse events which can be useful
   if you want to copy from the terminal instead of from mi-ide (if over ssh for
   example, because the terminal has access to the local clipboard and mi-ide
   does not).

	default value: `true`

* `pluginchannels`: contains all the channels mi-ide's plugin manager will search
   for plugins in. A channel is simply a list of 'repository' json files which
   contain metadata about the given plugin. See the `Plugin Manager` section of
   the `plugins` help topic for more information.

   default value: `https://raw.githubusercontent.com/hanspr/mi-channel/master/plugins.json`

* `pluginrepos`: contains all the 'repositories' mi-ide's plugin manager will
   search for plugins in. A repository consists of a `repo.json` file which
   contains metadata for a single plugin.

	default value: ` `

* `rmtrailingws`: mi-ide will automatically trim trailing whitespaces at eol.

	default value: `false`

* `ruler`: display line numbers.

	default value: `true`

* `savecursor`: remember where the cursor was last time the file was opened and
   put it there when you open the file again.

	default value: `false`

* `savehistory`: remember command history between closing and re-opening
   mi-ide.

    default value: `true`

* `saveundo`: when this option is on, undo is saved even after you close a file
   so if you close and reopen a file, you can keep undoing.

	default value: `false`

* `scrollmargin`: amount of lines you would like to see above and below the
   cursor.

	default value: `3`

* `smartindent`: It works only on languages that use braces to define blocks "C" like languages
   When you type the Tab key mi-ide will attempt to indent the line to the correct depth of
   indentation.

   default value: `false`

* `smartpaste`: should mi-ide add leading whitespace when pasting multiple lines?
   This will attempt to preserve the current indentation level when pasting an
   unindented block.

	default value: `true`

* `softwrap`: should mi-ide wrap lines that are too long to fit on the screen.

	default value: `false`

* `splitbottom`: when a horizontal split is created, should it be created below
   the current split?

	default value: `true`

* `splitfalse`: If splits should open empty or the current buffer on the new view
   default value: `false`

* `splitright`: when a vertical split is created, should it be created to the
   right of the current split?

	default value: `true`

* `matchbrace`: highlight matching braces for '()', '{}', '[]'

    default value: `false`

* `matchbraceleft`: when matching a closing brace, should matching match the
   brace directly under the cursor, or the character to the left? only matters
   if `matchbrace` is true

    default value: `false`

* `syntax`: turns syntax on or off.

	default value: `true`

* `tabmovement`: navigate spaces at the beginning of lines as if they are tabs
   (e.g. move over 4 spaces at once). This option only does anything if
   `tabstospaces` is on.

	default value: `false`

* `tabsize`: sets the tab size to `option`

	default value: `4`

* `tabstospaces`: use spaces instead of tabs

	default value: `false`

* `termtitle`: defines whether or not your terminal's title will be set by mi-ide
   when opened.

	default value: `false`

