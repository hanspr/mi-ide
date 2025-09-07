# ![mi-ide](./assets/logo.png)

mi-ide (___mi ide___ in Spanish, ___my ide___ in English. As in: this is ___my editor___) is a spin-off version of micro editor project.

This has become a personal project, but if you are still looking for something different you may try this editor.

What is this editor's philosophy:

* Is a remote terminal oriented editor
  * I'm not interested in developing another editor for the desktop environment
  * This editor is meant to run in small cloud servers
  * No space for language servers, not enough cpu to support a heavy load
* I do want to have as many amenities from well developed ides, like
  * __Syntax error checking__ inside the editor
  * __git status on the status bar__, and access to git status and git diff within the editor
  * __Search for functions__ within the project
  * __Snippets__
  * Good set of commands and key actions to help writing code fluently
  * A new approach to keybindins to __maintain the hands on the keyboard__ as much as possible
  * A clipboard that works across editors in diferente terminals and servers (__cloud clipboard__)
  * A plugin system to enhance the experience of developing in a particular language
  * __Syntax highlight__
  * A beautiful interface (why not)
* My particular setup is : tmux, lazygit and mi-ide. With this combination I have a very usefull development environment on my remote servers.

The original motivation came from the usual experience of searching for an editor when I began coding. Vim is too hard for me, nano way too simple.

I eventually landed with jed editor. It had the basic features that I liked, but again too limited and very difficult to enhance it. I did learn slang and did some hacking to it, but the framework had many limitations for me to grow it.

And one day I discover micro. Written in go. I was very interested in learning go. So I downloaded, installed, and run it. And at that time it was very promissing, but it was full of bugs that made it impractical (have to say that since then, I have't followed the project so I'm sure many issues were resolved).

But I took this opportunity to grab a book and started learning go. Jumped into the code, and after 3 months, I managed to remove most of the bugs, and began using it daily and little by little grow it to fit my needs. Which drove me very far away from the original project.

So this is the result of it, check the manual and see if it would fill your requirements. And if not, grab it, change it and customize it to your needs.

![mi ide view](assets/mi-ide.gif)

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Default keys](https://github.com/hanspr/mi-ide/wiki/defaultkeys#simplified-key-map)
- [Documentation and Help](#documentation-and-help)
- [Plugin Development](#plugin)
- [Translate](#translate)

- - -

## Installation

* mi-ide comes on a single binary, there are no installation scripts.
* Place it anywhere you want.
* Download a [prebuilt binary](https://github.com/hanspr/mi-ide/releases).
    - Unzip the release
        - `unzip release##.zip`
    - Place the binary in any location on your home directory, for example
        - `mv mi-ide ~/.local/bin`
        - `chmod 770 ~/.local/bin/mi-ide`
    - Create an alias in your `.bashrc` or `.bash_aliases` to the location of your executable
        - Change to your home directory : `cd`
        - Edit : `nano .bashrc` or `nano .bash_aliases`
        - Add : `alias mi-ide='~/.local/bin/mi-ide'`
        - Reload changes: `. .bashrc` or `. .bash_aliases`
    - Execute mi-ide.
        - The first time you run the editor will download the configurations from github.
        - And place them in `~/.config/mi-ide/`

* You can also build mi-ide from source. Clone the repo.

```bash
go get
go build .
```

## Usage

Once you have built/installed the editor, start it by running `mi-ide path/to/file.txt` or simply `mi-ide` to open an empty buffer.

The very first time will open a welcome.md file to give you the basic information about key bindings and how to access the help documentation.

You can move the cursor around with the arrow keys and mouse.

You can also use the mouse to select text, double click to enable a word, and triple click to enable line selection.

Mouse should be enabled by default, if is not, you can enable on the general settings or toggle by typing : `Ctrl-k p`

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
