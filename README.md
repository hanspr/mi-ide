# ![mi-ide](./assets/logo.png)

mi-ide (_mi ide_ in spanish, "my ide" as in: this is **my** editor) is a spin-off version of micro editor project at https://github.com/zyedidia/micro.

This project is mainly focused on **improving the experience of coding remotely over ssh on headless Linux servers** (no X Server support, no Graphical interface).

Many of features were already provided by micro editor :+1:. But some others were still missing :star:, and there inclusion are not necessarily useful to everyone.

This is the list of the most important features :

* :+1: Short learning curve
* :+1: Great syntax highlighting and easily customizable
* :star: Good cursor positioning after paste, cuts, duplicate a line, searching, moving on the edges of the window. So you always have a clear view of the surrounding code
* :star: Translate to other languages easily
* :star: Auto detect file encoding. Open (decode) / Save (encode) in the original encoding of the file: UTF8, ISO8859, WINDOWS, etc. (Limited to the encoders available in go libraries)
* :star: Replace the need to learn too many key combinations and commands by the use (and abuse) of good mouse support with: icons, buttons, dialog boxes. Similar to windowed editors
* :star: Use and abuse color to easily find the cursor, selected text, etc.
* :+1::star: Good and powerful plugin system to hack the editor to your personal needs.
    - Without the need to compile or setup a complicated environment
    - Resilient to editor new versions, so a plugin can last a long time without being recoded, and recoded.
    - Plugins that will enhance the user coding experience for each particular language
    - So you can say at the end of all your hacking: "finally!, I got **my** ide"
* :star: Save editor settings for
    - a particular coding language (c, php, python, perl,..)
    - a single file
    - or just for the current opened session
* :star: A powerful auto indent
* :star: Internal copy paste between terminals in the same server without the need of _**shift-key and mouse dragging**_
* :star: Features that require a server to work
    - Internal copy paste between terminals in different servers without the need of _**shift-key and mouse dragging**_
    - Package and retrieve your personal settings to quickly sync your editor's choices and hacks across different servers you work on, when the editor is already installed
    - Transfer a single script, text file quickly from one server to another
    - See [Support the project & Services](#support-the-project--services) for these features

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
        - add : `alias='mi-ide ~/.local/mi-ide'`
        - reload changes: `. .bashrc` or `. .bash_aliases`
    - Execute mi-ide.
        - The first time you run the editor will download the configurations from github. **(β)** please read note
        - And place them in `~/.config/mi-ide/`

>(β) **Important note**
>
>For convenience to most people, included in the download is a compiled version of this piece of software: uchardet ( https://www.freedesktop.org/wiki/Software/uchardet/ ). It is compiled for 64 bits (no 32 bit version distributed).
>
>uchardet is not installed in all distributions by default, and some distributions have a very old one. And there is no port for it in pure "go" yet.
>
>This compiled binary is used as an external application, to detect the encoding of the file about to be opened, and then set the correct decoder and read the file into the editor. It was not linked with CGO, to avoid portability or compiling problems.
>
>Some people will not like this convenience (downloading a binary file). If you feel uncomfortable, this are the alternatives:
>
>1. Start mi-ide (do not open any file so the uchardet code will not be executed) and let mi-ide download the configurations
>    1. exit the editor
>    2. go to your settings folder (~./config/mi-ide/libs)
>    3. delete the files: uchardet, libuchardet.so.0
>2. If you don't want to run mi-ide, then download config.zip manually from: https://github.com/hanspr/mi-channel
>    1. Extract the directory
>    2. Remove uchardet, libuchardet.so.0
>    3. Copy the config directory to your .config directory
>2. If after removing the binaries you still want uchardet support into mi-ide, your options are:
>    1. install uchardet for your distribution on your server (make sure that is version : 0.0.6 or above)
>    2. Or, download and compile your own version from https://www.freedesktop.org/wiki/Software/uchardet/ and install the compiled files on ~./config/mi-ide/libs or globally in your server

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

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to select text, double click to enable a word, and triple click to enable line selection.

[For a full introduction you may watch this video](https://youtu.be/6bgv3bQbxgM) (27 minutes)

# Documentation and Help

Mi-ide has a built-in help system which you can access by pressing `Alt-?` or `Ctrl-E` and typing `help`.

You will also find more detiled information on the [Wiki pages](https://github.com/hanspr/mi-ide/wiki)

# Plugin

The plugin system has been modified in this version, you should be able to run any plugin you develop for micro editor, by doing some minor adjustments.

Please visit the [developers page](https://github.com/hanspr/mi-ide/wiki) with full instructions and videos about the plugin framework

# Contributing

You can use the [GitHub issue tracker](https://github.com/hanspr/mi-ide/issues) to report bugs, ask questions, or suggest new features.

To create pull requests, please follow these recommendations:

* Document very well your modifications in the code, so I can understand the changes
* Test your changes for a few weeks (by using the editor on real work). To confirm that your modification does not create a side effects on the rest of the editor. I tell this out of personal experience.
    - At the beginning I used to change one line of code and thought that I had fixed or improved something. A few days later I realized I broke something else.

## Translate

You may find the translation file in your config directory under : langs

- Open the file : en_US.lang
- Change the name to your ISO code location
- Translate each sentence after the | (pipe)
- Save the final file with the new name.
- Switch language by going to the menu > Global Settings
- Change to your new language
- If you want to contribute your translation
    - Clone : https://github.com/hanspr/mi-sources
    - Copy your file or update the current translation in the langs directory
    - Create a pull request

**Note for English speaking people**

>English is not my first tongue, so If there are misspelled words or grammar errors, you may have to translate the mistake to the correct sentence (as explained above) and create push request with your en_US.md translation from bad English to correct English

# Support the project & Services

If you find this project useful and want to support it, you will receive a lifetime Cloud Key to activate Cloud Services.

|Kind       | Amount | Button|
|----------|--------|-------|
|Just Support, not interested in a Cloud Key | 5 USD | [![mi-ide-20](https://img.shields.io/badge/Just%20Support-PayPal-green)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=YDQ2HBASRXTCE)|
|Support & Cloud Key | 20 USD | [![mi-ide-20](https://img.shields.io/badge/Support%20%26%20Cloud%20Key-PayPal-green)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=K43LR98FUSSNG)|
|Cloud Key & More Support | 50 USD | [![mi-ide-20](https://img.shields.io/badge/Cloud%20Key%20%26%20More-PayPal-green)](https://www.paypal.com/cgi-bin/webscr?cmd=_s-xclick&hosted_button_id=SAPEG25ZKA67N)|


**A Cloud Key, gives you lifetime access to**

* Copy Paste between your editors running in different servers
* Transfer any code that is opened in the current Window to different server runngin mi-ide
* Save and retrieve your settings from any server. To sync all your : settings, plugins, hacks, colors, bindings, etc.
* All your information is completely confidential and secure.
    - Check the code at : ["github.com/hanspr/clipboard"](github.com/hanspr/clipboard)
    - It is very easy to read and follow
* How does it work?
    - All connections are over https
    - You set up 3 values in the editor
        - The key assigned to you
        - A password that you select
        - A passphrase that you select
    - The key and password gives you access to your clipboard, settings, and files.
        - The key and password are transmitted to the server on every transaction to validate the access to your account
    - The passphrase:
        - Is never sent to the server
    - On Copy, Upload File or Settings
        - mi-ide encrypts your clip, file or settings using your passphrase
        - The application, sends your key and password only
        - Transfers the data to the server
        - The server stores your encrypted : clipboard, file or settings
    - On Paste, File Download or Settings
        - The application, sends your key and password only
        - It receives the data : clipboard, file or settings
        - Decrypts the data using your passphrase
            - Inserts the clipboard in your current buffer, or
            - Opens the file in a new tab, or
            - Installs your settings

There is no way to access your data in the server.
