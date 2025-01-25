package commandhandler

import (
	"regexp"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Handler interface {
	OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate)
	OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func NewHandler(prefix string, cmds []Command, resolver Resolver) Handler {
	return SimpleHandler{
		resolver, prefix, cmds,
	}
}

type SimpleHandler struct {
	resolver Resolver
	prefix   string
	cmds     []Command
}

func (h SimpleHandler) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !strings.HasPrefix(m.Content, h.prefix) {
		return
	}

	args := getArgs(m.Content, h.prefix)

	cmd, cmdHierarchy, args, cmdErr := parseArgs(h.cmds, args)

	ctx := MessageToContext(s, m.Message)

	if cmdErr.Err != nil {
		ctx.Reply(FormatCommandError(cmdHierarchy, cmdErr.Cmd, cmdErr.Err))
		return
	}

	opts, optErr := h.resolver.ResolveMessageOptions(cmd, ctx, args)

	if optErr.Err != nil {
		ctx.Reply(FormatOptionError(cmdHierarchy, args, optErr.Opt, optErr.Err))
		return
	}

	optErr = Validate(cmd.Options, opts)

	if optErr.Err != nil {
		ctx.Reply(FormatOptionError(cmdHierarchy, args, optErr.Opt, optErr.Err))
		return
	}

	cmd.Run(ctx, opts)
}

func (h SimpleHandler) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := SlashCommandToContext(s, i)

	cmd, cmdHierarchy, cmdErr := parseSlashCommandArgs(h.cmds, i)
	if cmdErr.Err != nil {
		ctx.Reply(FormatCommandError(cmdHierarchy, cmdErr.Cmd, cmdErr.Err))
		return
	}

	var opts map[string]any
	var optErr OptionError
// TODO maybe make resolver genric and parse args here just like OnMessageCreate
	switch len(cmdHierarchy) {
	case 1:
		opts, optErr = h.resolver.ResolveSlashCommandOptions(cmd, ctx, i.ApplicationCommandData().Options)
	case 2:
		opts, optErr = h.resolver.ResolveSlashCommandOptions(cmd, ctx, i.ApplicationCommandData().Options[0].Options)
	case 3:
		opts, optErr = h.resolver.ResolveSlashCommandOptions(cmd, ctx, i.ApplicationCommandData().Options[0].Options[0].Options)
	}

	if optErr.Err != nil {
		opts := []string{}
		for _, opt := range cmd.Options {
			opts = append(opts, opt.Name)
		}
		ctx.Reply(FormatOptionError(cmdHierarchy, opts, optErr.Opt, optErr.Err))
		return
	}

	optErr = Validate(cmd.Options, opts)

	if optErr.Err != nil {
		opts := []string{}
		for _, opt := range cmd.Options {
			opts = append(opts, opt.Name)
		}
		ctx.Reply(FormatOptionError(cmdHierarchy, opts, optErr.Opt, optErr.Err))
		return
	}

	cmd.Run(ctx, opts)
}

func findCommand(cmds []Command, name string) (Command, bool) {
	for _, cmd := range cmds {
		if cmd.Name == name || slices.Contains(cmd.Aliases, name) {
			return cmd, true
		}
	}

	return Command{}, false
}

func getArgs(message string, prefix string) []string {
	args := []string{}

	regex := regexp.MustCompile(`"(.*)"|([^"\s]*)`)
	for _, match := range regex.FindAllStringSubmatch(strings.TrimPrefix(message, prefix), -1) {
		if match[1] != "" {
			args = append(args, match[1])
		} else {
			args = append(args, match[2])
		}
	}

	return args
}

func parseArgs(cmds []Command, args []string) (lastCmd Command, cmdHierarchy []string, newArgs []string, err CommandError) {
	for _, arg := range args {
		cmd, ok := findCommand(cmds, arg)
		if !ok {
			if len(lastCmd.Subs) > 0 {
				err = CommandError{arg, InvalidSubCommandError}
				return
			}
			break
		}

		lastCmd = cmd
		cmdHierarchy = append(cmdHierarchy, arg)
		cmds = cmd.Subs

		if len(lastCmd.Subs) == 0 {
			break
		}
	}

	if len(lastCmd.Subs) > 0 {
		err = CommandError{"", RequiredSubCommandError}
		return
	}

	newArgs = args[len(cmdHierarchy):]

	return
}

func parseSlashCommandArgs(cmds []Command, i *discordgo.InteractionCreate) (cmd Command, cmdHierarchy []string, err CommandError) {
	d := i.ApplicationCommandData()

	if len(d.Options) == 0 || (d.Options[0].Type != discordgo.ApplicationCommandOptionSubCommandGroup && d.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand) {
		c, ok := findCommand(cmds, d.Name)
		if !ok {
			err.Err = CommandNotFoundError
			return
		}
		cmdHierarchy = append(cmdHierarchy, c.Name)
		cmd = c
	} else if d.Options[0].Type == discordgo.ApplicationCommandOptionSubCommandGroup {
		c1, ok1 := findCommand(cmds, d.Name)
		c2, ok2 := findCommand(c1.Subs, d.Options[0].Name)
		c3, ok3 := findCommand(c2.Subs, d.Options[0].Options[0].Name)
		if !ok1 || !ok2 || !ok3 {
			err.Err = CommandNotFoundError
			return
		}
		cmdHierarchy = append(cmdHierarchy, c1.Name, c2.Name, c3.Name)
		cmd = c3
	} else {
		c1, ok1 := findCommand(cmds, d.Name)
		c2, ok2 := findCommand(c1.Subs, d.Options[0].Name)
		if !ok1 || !ok2 {
			err.Err = CommandNotFoundError
			return
		}
		cmdHierarchy = append(cmdHierarchy, c1.Name, c2.Name)
		cmd = c2
	}

	return
}
