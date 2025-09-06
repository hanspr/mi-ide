# Possible commands

You can execute an editor command by pressing `Ctrl-e` followed by the command.
Here are the possible commands that you can use.

Tab key lets you see the list of availalbe commands at a given label, and autocompletes

* `quit`: Quits mi-ide.

* `cd path`: Change the working directory to the given `path`.

* `config`: Submenu for config commands, possible options
  `buffersettings` : buffer settings for current file
  `settings` : global settings
  `keybindings` : reconfigure bindings
  `plugins` : open the plugin manager

* `edit`: Submenu to edit config files directly
  `settings` : global settings
  `snippets` : snippets for current buffer language

* `git`: Submenu to execute some git commands
  `status` : git status
  `diff` : open new tab with the git diff
  `diffstaged` : open new tab with the git diff --staged

* `help` : Access to help topics

* `log`: opens a log of all messages and debug statements.

* `reload`: reloads all runtime files. Only needed if you edit configuration files : colors, syntax, etc.

* `open filename`: Open a file in the current buffer.

* `save <filename>`: Saves the current buffer. If the filename is provided it will 'save as' the filename.

* `pwd`: Print the current working directory.

* `show`: Show codeing help information

---

# Command Parsing

When running a command, you can use extra syntax that mi-ide will expand before
running the command. To use an argument with a space in it, simply put it in
quotes. You can also use environment variables in the command bar and they
will be expanded to their value. Finally, you can put an expression in backticks
and it will be evaluated by the shell beforehand.
