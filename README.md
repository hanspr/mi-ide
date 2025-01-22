# ![mi-ide](./assets/logo.png)

mi-ide (_**mee ide**_ in spanish, _**my ide**_ in English. As in: this is _**my** editor_) is a spin-off version of micro editor project at https://github.com/zyedidia/micro.

This has become a personal project, but if you are still looking for something different you may try this editor.

What is this editor philosophy:

* Is a remote terminal oriented editor
  * I'm not interested in developing another editor for the desktop environment
  * This editor is meant to run in small cloud servers
  * No space for language servers, not enough cpu to support a heavy load
* I do want to have as many amenities from well developed ides, like
  * Syntax error checking inside the editor
  * Check git status and git diff
  * Search for functions within the project
  * Snippets
  * Good set of commands and key actions to help writing code fluently
  * A new approach to keybindins to maintain the hands on the keyboard as much as possible
  * A clipboard that works across editors in diferente terminals and servers
  * A plugin system to enhance the experience of developing in a particular language
  * Syntax highlight
* My particular setup is : tmux, gitui and miide. With this combination I have a very usefull development environment on my remote servers.

The original motivation came from the usual experience of searching for an editor when began coding. Vim was to hard for me, nano way to simple.

I eventually landed with jed editor. It had the basic features that I liked, but again to limited and very difficult to enhance it. I did learn slang and did some hacking to it, but the framework had to many limitations to grow it.

And one day I discover micro. Written in go. I was very interested in learning go. So I downloaded, installed, and run it. And at that time it was very promissing but it was full of bugs that made it impractical (have to say that since then have not followed the project so I'm sure many issues were resolved).

But took this opportunity to grab a book and begin learning go. Jumped into the code, and after 3 long months managed to remove most of the bugs, and began using it daily and little by little grow it to fit my needs. Which drove me very far away from the original project.

So this is the result of it, check the manual and see if it would fill your requirements.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Documentation and Help](#documentation-and-help)
- [Plugin Development](#plugin)
- [Translate](#translate)

- - -

## Installation

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

```bash
export TERM=xterm-256color

#Add this line to your .bashrc
export TERM=xterm-256color
```

## Usage

Once you have built/installed the editor, start it by running `mi-ide path/to/file.txt` or simply `mi-ide` to open an empty buffer.

The very first time will open a welcome.md file to give you the basic information about key bindings and how to access the help documentation.

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to select text, double click to enable a word, and triple click to enable line selection.

[For a full introduction you may watch this video](https://youtu.be/grHzfIvC6_I) (22 minutes)

## Documentation and Help

Mi-ide has a built-in help system which you can access by pressing `Alt-?` or `Ctrl-E` and typing `help`.

You will also find more detailed information on the [Wiki pages](https://github.com/hanspr/mi-ide/wiki)

## Plugin

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
