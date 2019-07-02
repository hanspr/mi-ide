# Tutorial

This is a brief intro to micro's configuration system that will give some simple
examples showing how to configure settings, rebind keys, and use `init.lua` to
configure micro to your liking.

Hopefully you'll find this useful.

See `> help defaultkeys` for a list an explanation of the default keybindings.

### Plugins

Micro has a plugin manager which can automatically download plugins for you. To
see the 'official' plugins, go to github.com/micro-editor/plugin-channel.

For example, if you'd like to install the snippets plugin (listed in that repo),
just press `CtrlE` to execute a command, and type `plugin install snippets`.

For more information about the plugin manager, see the end of the `plugins` help
topic.

### Settings

In micro, your settings are stored in `~/.config/micro/settings.json`, a file
that is created the first time you run micro. It is a json file which holds all
the settings and their values. To change an option, you can either change the
value in the `settings.json` file, or you can type it in directly while using
micro.

Simply press CtrlE to go to command mode, and type `set option value` (in the
future, I will use `> set option value` to indicate pressing CtrlE). The change
will take effect immediately and will also be saved to the `settings.json` file
so that the setting will stick even after you close micro.

You can also set options locally which means that the setting will only have the
value you give it in the buffer you set it in. For example, if you have two
splits open, and you type `> setlocal tabsize 2`, the tabsize will only be 2 in
the current buffer. Also micro will not save this local change to the
`settings.json` file. However, you can still set options locally in the
`settings.json` file. For example, if you want the `tabsize` to be 2 only in
Ruby files, and 4 otherwise, you could put the following in `settings.json`:

```json
{
    "*.rb": {
        "tabsize": 2
    },
    "tabsize": 4
}
```

Micro will set the `tabsize` to 2 only in files which match the glob `*.rb`.

If you would like to know more about all the available options, see the
`options` topic (`> help options`).

### Keybindings

Keybindings work in much the same way as options. You configure them using the
`~/.config/micro/bindings.json` file.

For example if you would like to bind `CtrlR` to redo you could put the
following in `bindings.json`:

```json
{
    "CtrlR": "redo"
}
```

Very simple.

You can also bind keys while in micro by using the `> bind key action` command,
but the bindings you make with the command won't be saved to the `bindings.json`
file.

For more information about keybindings, like which keys can be bound, and
what actions are available, see the `keybindings` help topic (`> help keybindings`).

### Configuration with Lua

If you need more power than the json files provide, you can use the `init.lua`
file. Create it in `~/.config/micro`. This file is a lua file that is run when
micro starts and is essentially a one-file plugin.

I'll show you how to use the `init.lua` file by giving an example of how to
create a binding to `CtrlR` which will execute `go run` on the current file,
given that the current file is a Go file.

You can do that by putting the following in `init.lua`:

```lua
function gorun()
    local buf = CurView().Buf -- The current buffer
    if buf:FileType() == "go" then
        HandleShellCommand("go run " .. buf.Path, true, true) -- the first true means don't run it in the background
    end
end

BindKey("CtrlR", "init.gorun")
```

Alternatively, you could get rid of the `BindKey` line, and put this line in the
`bindings.json` file:

```json
{
    "CtrlR": "init.gorun"
}
```

For more information about plugins and the lua system that micro uses, see the
`plugins` help topic (`> help plugins`).
