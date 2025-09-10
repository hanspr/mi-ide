package main

import (
	"os"
	"sort"
	"strings"
)

var pluginCompletions []func(string) []string

// This file is meant (for now) for autocompletion in command mode, not
// while coding. This helps mi-ide autocomplete commands and then filenames
// for example with `vsplit filename`.

// FileComplete autocompletes filenames
func FileComplete(input string) (string, []string) {
	var sep string = string(os.PathSeparator)
	dirs := strings.Split(input, sep)

	var files []os.DirEntry
	var err error
	if len(dirs) > 1 {
		directories := strings.Join(dirs[:len(dirs)-1], sep) + sep

		directories = ReplaceHome(directories)
		files, err = os.ReadDir(directories)
	} else {
		files, err = os.ReadDir(".")
	}

	var suggestions []string
	if err != nil {
		return "", suggestions
	}
	for _, f := range files {
		name := f.Name()
		if f.IsDir() {
			name += sep
		}
		if strings.HasPrefix(name, dirs[len(dirs)-1]) {
			suggestions = append(suggestions, name)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		if len(dirs) > 1 {
			chosen = strings.Join(dirs[:len(dirs)-1], sep) + sep + suggestions[0]
		} else {
			chosen = suggestions[0]
		}
	} else {
		if len(dirs) > 1 {
			chosen = strings.Join(dirs[:len(dirs)-1], sep) + sep
		}
	}

	return chosen, suggestions
}

// CommandComplete autocompletes commands
func CommandComplete(input string) (string, []string) {
	var suggestions []string
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, cmd := range keys {
		if strings.HasPrefix(cmd, input) {
			suggestions = append(suggestions, cmd)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

func GroupComplete(group, input string) (string, []string) {
	var suggestions []string
	var options []string
	var chosen = ""
	i := strings.Index(group, ":")
	group = group[0:i]
	switch group {
	case "edit":
		options = []string{"settings", "snippets"}
	case "config":
		options = []string{"buffersettings", "cloudsettings", "keybindings", "plugins", "settings"}
	case "git":
		options = []string{"diff", "diffstaged", "status"}
	case "show":
		options = []string{"snippets"}
	}
	for _, cmd := range options {
		if strings.HasPrefix(cmd, input) {
			suggestions = append(suggestions, cmd)
		}
	}
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

// HelpComplete autocompletes help topics
func HelpComplete(input string) (string, []string) {
	var suggestions []string

	for _, file := range ListRuntimeFiles(RTHelp) {
		topic := file.Name()
		if strings.HasPrefix(topic, input) {
			suggestions = append(suggestions, topic)
		}
	}

	var chosen string
	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

// MakeCompletion registers a function from a plugin for autocomplete commands
func MakeCompletion(function string) Completion {
	pluginCompletions = append(pluginCompletions, LuaFunctionComplete(function))
	return Completion(-len(pluginCompletions))
}

// PluginComplete autocompletes from plugin function
func PluginComplete(complete Completion, input string) (chosen string, suggestions []string) {
	idx := int(-complete) - 1

	if len(pluginCompletions) <= idx {
		return "", nil
	}
	suggestions = pluginCompletions[idx](input)

	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return
}

// PluginCmdComplete completes with possible choices for the `> plugin` command
func PluginCmdComplete(input string) (chosen string, suggestions []string) {
	for _, cmd := range []string{"install", "remove", "search", "update", "list"} {
		if strings.HasPrefix(cmd, input) {
			suggestions = append(suggestions, cmd)
		}
	}

	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}

// PluginNameComplete completes with the names of loaded plugins
func PluginNameComplete(input string) (chosen string, suggestions []string) {
	for _, pp := range GetAllPluginPackages() {
		if strings.HasPrefix(pp.Name, input) {
			suggestions = append(suggestions, pp.Name)
		}
	}

	if len(suggestions) == 1 {
		chosen = suggestions[0]
	}
	return chosen, suggestions
}
