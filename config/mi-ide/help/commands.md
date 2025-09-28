# Possible commands

You can execute an editor command by pressing `Ctrl-e` followed by the command.
Here are the possible commands that you can use.

Tab key lets you see the list of availalbe commands at a given label, and autocompletes

|Command |               |Action                                                                                       |
|--------|---------------|---------------------------------------------------------------------------------------------|
|quit    |               |Quits mi-ide.                                                                                |
|cd      |`path`         |Change the working directory to the given `path`.                                            |
|config  |               |Submenu for config commands, possible options                                                |
|        |buffersettings |buffer settings for current file                                                             |
|        |cloudsettings  |configure cloud settings window                                                              |
|        |keybindings    |open bindings keys window                                                                    |
|        |plugins        |open the plugin manager                                                                      |
|        |settings       |show global settings window                                                                  |
|edit    |               |Submenu to edit config files directly                                                        |
|        |settings       |edit global settings json                                                                    |
|        |snippets       |edit snippets for current buffer file type                                                   |
|gemini  |               |ask question to gemini service api                                                           |
|        |               |you must have a valid api key `export GEMINI_API_KEY=<your key>`                             |
|git     |               |Submenu to execute some git commands                                                         |
|        |status         |git status                                                                                   |
|        |diff           |open new tab with the `git diff`                                                             |
|        |diffstaged     |open new tab with the git `diff --staged`                                                    |
|help    |               |Access to help topics (Tab to see available topics)                                          |
|log     |               |opens a log of all messages and debug statements.                                            |
|reload  |               |reloads all runtime files. Only needed if you edit configuration files: colors, syntax, etc. |
|save    |`filename`     |Saves the current buffer. If the filename is provided it will `save as` the filename.        |
|show    |               |Show coding help information                                                                 |
|        |snippets       |show available snippet names for current filetype buffer                                     |
|pwd     |               |Print the current working directory.                                                         |
|open    |`filename`     |Open a file in the current buffer.                                                           |

---

# Command Parsing

When running a command, you can use extra syntax that mi-ide will expand before
running the command. To use an argument with a space in it, simply put it in
quotes. You can also use environment variables in the command bar and they
will be expanded to their value. Finally, you can put an expression in backticks
and it will be evaluated by the shell beforehand.
