# ![Micro](./assets/logo.png)

Micro ide is a spin-off version of the great micro editor project at https://github.com/zyedidia/micro.

This version is a highly customized and modified versión with many of the features that I always wanted in a ssh terminal editor and that are some how missing in all the editors I tried (or some has all of them, but there learning curve has always been to steep for me).

Many of those features were already provided by micro editor :+1:. But some others were still missing :star:, and there inclusion are personal and not necessarily useful to everyone. This are some of these features:

* Low learning curve :+1:
* Great syntax highlighting and easily customizable :+1:
* Good cursor positioning after paste, cuts, duplicate a line, moving on the edges of the window, reposition on search, etc. :star:
* Multiple language support :star:
* Auto detect file encoding, open encode and decode: UTF8, ISO8859, WINDOWS, etc :star:
* Replace the need to learn too many key combinations and commands by the use (and abuse) of good mouse support with: icons, buttons, dialog boxes. Similar to windowed editors. :star:
* Use and abuse color to easily find the cursor, selected text, etc. :star:
* Good and powerful plugin system to hack the editor to my personal needs (:+1::star:)
    - Without the need to compile or setup a complicated environment
    - Resilient to editor new versions
* Save settings for :star:
    - a coding language (c,php,python,perl,..)
    - or single file
* A more powerful auto indent :star:
* Auto complete :star:
* Internal copy paste between terminals in the same server without the need of shift-key and mouse dragging :star:
* Features that require an external server to work :star:
    - Internal copy paste between terminals in different servers without the need of shift-key and mouse dragging
    - Package my full editor to install on new servers
    - Package and retrieve my settings to quickly sync my editor's choices and hacks across different servers I work on
    - Transfer a single script, text file quickly from one server to another
    - See [Support the project & Services](#support-the-project) for these features

![Screenshot](./assets/features.gif)


# Table of Contents
- [Features](#features)
- [New](#new)
- [Installation](#installation)
- [Usage](#usage)
- [Documentation and Help](#documentation-and-help)
- [Plugin Development](#plugin)
- [Contributing](#contributing)
- [Support the project & Services](#support-the-project)

- - -

# New

* Polished looks
* Many small bugs fixed
* Improved status line, icons, and buttons to set buffer settings
* Added a toolbar
* Added a menu
* Multi language support
* Encoding translation from/to UTF8 to: ISO-8859,WINDOWS,SHIFT-JIS,BIG5,etc.
* Added dialogs and configuration screens for:
    - Settings, keybindings, search, replace.
* Plugins
    - The interface was redesigned to avoid conflicts between plugins
    - Settings are now independent for each plugin
    - Access to create dialogs and add icons to the toolbar
    - Access to add submenus the the main menu

# Installation

To install micro, you can download a [prebuilt binary](https://github.com/hanspr/micro-ide/releases), or you can build it from source.

### Colors and syntax highlighting

If you open micro and it doesn't seem like syntax highlighting is working, this is probably because you are using a terminal which does not support 256 color. Try changing the colorscheme to `simple` by pressing CtrlE in micro and typing `set colorscheme simple`.

If you are using the default Ubuntu terminal, to enable 256 make sure your `TERM` variable is set to `xterm-256color`.

Many of the Windows terminals don't support more than 16 colors, which means that micro's default colorscheme won't look very good. You can either set the colorscheme to `simple`, or download a better terminal emulator, like mintty.

# Usage

Once you have built/installed the editor, start it by running `micro-ide path/to/file.txt` or simply `micro-de` to open an empty buffer.

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to manipulate the text. Simply clicking and dragging will select text. You can also double click to enable word selection, and triple click to enable line selection.

For a brief introduction you may watch this video

# Documentation and Help

Micro-ide has a built-in help system which you can access by pressing `Alt-?` or `Ctrl-E` and typing `help`.

# Plugin

The plugin system has been modified in this version, you should be able to make any previous plugin you develop by doing some minor adjustments.

Please visit the developers page with full instructions at:

# Contributing

If you find any bugs, please report them! I am also happy to accept pull requests from anyone.

You can use the [GitHub issue tracker](https://github.com/hanspr/micro-ide/issues) to report bugs, ask questions, or suggest new features.

# Support the project

If you find this project useful and want to support it, please use the following page to process your contribution.

This page will allow you to receive a key code that you can use as a permanent access to additional features of the editor that require cloud support to work, like:

* Internal copy paste between editors running in different servers
* File transfer between editors in different servers
* Save and retrieve your settings from any server. To sync all your micro-ide editors runngin in different machines
* Create your own full packed micro-ide with all your personal settings, plugins, hacks. So you may download and deploy on other servers without having to reconfigure and reinstall everything
