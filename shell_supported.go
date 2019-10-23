// +build linux darwin dragonfly openbsd_amd64 freebsd

package main

import (
	"github.com/hanspr/shellwords"
)

// TermEmuSupported terminal support
const TermEmuSupported = true

// RunTermEmulator start terminal
func RunTermEmulator(input string, wait bool, getOutput bool, callback string) error {
	args, err := shellwords.Split(input)
	if err != nil {
		return err
	}
	err = CurView().StartTerminal(args, wait, getOutput, callback)
	return err
}
