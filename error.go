package commandhandler

import (
	"errors"
)

var (
	CommandNotFoundError    = errors.New("command not found")
	RequiredSubCommandError = errors.New("subcommand required but not found")
	InvalidSubCommandError  = errors.New("unknown subcommand")
	RequiredOptionError     = errors.New("option required but not given")
	InvalidOptionError      = errors.New("unknown option")
)

type OptionError struct {
	Opt string
	Err error
}

type CommandError struct {
	Cmd string
	Err error
}

func FormatOptionError(cmdHierarchy []string, opts []string, opt string, err error) string {
	message := "Command: "

	for _, cmd := range cmdHierarchy {
		message += cmd + " "
	}

	for _, o := range opts {
		if o == opt {
			message += "**" + o + "**"
			break
		} else {
			message += o + " "
		}
	}

	message += "\nError: " + err.Error()
	return message
}

func FormatCommandError(cmdHierarchy []string, cmd string, err error) string {
	message := "Command: "

	for _, c := range cmdHierarchy {
		if c == cmd {
			message += "**" + c + "**"
			break
		} else {
			message += c + " "
		}
	}

	message += "\nError: " + err.Error()
	return message
}
