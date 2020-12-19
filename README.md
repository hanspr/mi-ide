# ![mi-ide](./assets/logo.png)

__This project is personal project.__

mi-ide (_**mee ide**_ in spanish, _**my ide**_ in English. As in: this is _**my** editor_) is a spin-off version of micro editor project at https://github.com/zyedidia/micro.

This is the list of the most important existing features :

* :+1: Short learning curve
* :+1: Great syntax highlighting and easily customizable
* :star: Good cursor positioning after paste, cuts, duplication of lines, searching, moving on the windows edges. So you always have a clear view of the surrounding code
* :star: Easy translation to other languages
* :star: Auto detect file encoding. Open (decode) / Save (encode) in the original encoding of a file: UTF8, ISO8859, WINDOWS, etc.
* :star: Ease the need to learn too many key combinations and commands by the use (and abuse) of good mouse support with: icons, buttons, dialog boxes.
* :star: Use and abuse color to easily find the cursor, selected text, etc.
* :+1::star: Good and powerful plugin system to hack the editor to your personal needs.
    - Without the need to compile or setup a complicated environment
    - Resilient to editor's new versions.
    - Plugins that will enhance the user coding experience for each particular language
    - So you can say at the end of all your hacking: "finally!, I got **my** IDE"
* :star: Save editor settings for:
    - A particular coding language (c, php, python, perl,..)
    - A single file
    - Or just for the current opened session
* :star: A powerful auto indent
* :star: Internal copy paste between terminals in the same server without the need of _**shift-key and mouse dragging**_

![Screenshot](./assets/features.gif)

# Table of Contents
- [Installation](#installation)
- [Usage](#usage)
- [Documentation and Help](#documentation-and-help)
- [Plugin Development](#plugin)
- [Contributing](#contributing)
    - [Translate](#translage)
- [Support the project & Services](#support-the-project)

- - -

# Installation

* Download a [prebuilt binary](https://github.com/hanspr/mi-ide/releases).
    - Unzip the release
        - `unzip release##.zip`
    - Place the binary in any location on your home directory, for example
        - `mv mi-ide ~/.local`
    - Create an alias in your `.bashrc` or `.bash_aliases` to the location of your executable
        - Change to your home directory : `cd`
        - Edit : `nano .bashrc` or `nano .bash_aliases`
        - Add : `alias mi-ide='~/.local/mi-ide'`
        - Reload changes: `. .bashrc` or `. .bash_aliases`
    - Execute mi-ide.
        - The first time you run the editor will download the configurations from github.
        - And place them in `~/.config/mi-ide/`

* You can also build mi-ide from source, by cloning this repo and install all dependencies.

```bash
go get "github.com/blang/semver"
go get "github.com/dustin/go-humanize"
go get "github.com/flynn/json5"
go get "github.com/go-errors/errors"
go get "github.com/hanspr/clipboard"
go get "github.com/hanspr/glob"
go get "github.com/hanspr/highlight"
go get "github.com/hanspr/ioencoder"
go get "github.com/hanspr/lang"
go get "github.com/hanspr/shellwords"
go get "github.com/hanspr/tcell"
go get "github.com/hanspr/terminal"
go get "github.com/hanspr/terminfo"
go get "github.com/mattn/go-isatty"
go get "github.com/mattn/go-runewidth"
go get "github.com/mitchellh/go-homedir"
go get "github.com/phayes/permbits"
go get "github.com/sergi/go-diff/diffmatchpatch"
go get "github.com/yuin/gopher-lua"
go get "layeh.com/gopher-luar"
```

### Colors and syntax highlighting

If your terminal does not support 256 color. Try changing the colorscheme to `default16`, by going to Menu > System Settings : Select "colorscheme".

If your terminal supports 256 colors but you do not see the full colors available you may try these commands:

```
export TERM=xterm-256color

#Add this line to your .bashrc
export TERM=xterm-256color
```

# Usage

Once you have built/installed the editor, start it by running `mi-ide path/to/file.txt` or simply `mi-ide` to open an empty buffer.

The very first time will open a welcome.md file to give you the basic information about key bindings and how to access the help documentation.

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to select text, double click to enable a word, and triple click to enable line selection.

[For a full introduction you may watch this video](https://youtu.be/grHzfIvC6_I) (22 minutes)

# Documentation and Help

Mi-ide has a built-in help system which you can access by pressing `Alt-?` or `Ctrl-E` and typing `help`.

You will also find more detailed information on the [Wiki pages](https://github.com/hanspr/mi-ide/wiki)

# Plugin

The plugin system has been modified in this project, you should be able to run any plugin you develop for micro editor, by doing some minor adjustments.

Please visit the [developers page](https://github.com/hanspr/mi-sources/wiki/plugins) with full instructions and videos about the plugin framework

## Translate

You will find the translation file in your config directory under : langs

- Open the file : en_US.lang
- Change the name to your ISO code location
- Translate each sentence after the | (pipe)
- Save the final file with the new name.
- Switch language by going to the menu > Global Settings
- Change to your new language
