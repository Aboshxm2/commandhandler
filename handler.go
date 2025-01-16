package commandhandler

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Handler interface {
	Register(cmd Command)
	OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate)
	OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate)
}

func NewHandler(prefix string, cmds []Command, resolver Resolver) Handler {
	return SimpleHandler{
		resolver, prefix, cmds,
	}
}

var errNotFound = errors.New("command not found")

type SimpleHandler struct {
	resolver Resolver
	prefix   string
	cmds     []Command
}

func (h SimpleHandler) Register(cmd Command) {
	h.cmds = append(h.cmds, cmd)
}

func (h SimpleHandler) OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if !strings.HasPrefix(m.Content, h.prefix) {
		return
	}

	args := getArgs(m.Content, h.prefix)

	cmd, args, err := findCommand(h.cmds, args, 1)

	ctx := MessageToContext(s, m.Message)

	if err != nil {
		if err != errNotFound {
			ctx.Reply("Error: " + err.Error())
		}
		return
	}

	opts, err := h.resolver.ResolveMessageOptions(cmd, ctx, args)

	if err != nil {
		ctx.Reply("Error: " + err.Error())
		return
	}

	errors := Validate(cmd.Options, opts)

	if len(errors) > 0 {
		ctx.Reply(fmt.Sprintf("An error has occurred in option '%s'. Error: %s", errors[0].Opt.Name, errors[0].Err))
		return
	}

	cmd.Run(ctx, opts)
}

func (h SimpleHandler) OnInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	cmd, err := findSlashCommandSubCommand(h.cmds, i)

	if err != nil {
		return
	}

	ctx := SlashCommandToContext(s, i)

	opts, err := h.resolver.ResolveSlashCommandOptions(cmd, ctx, i.ApplicationCommandData())
	if err != nil {
		ctx.Reply("Error: " + err.Error())
		return
	}

	errors := Validate(cmd.Options, opts)

	if len(errors) > 0 {
		ctx.Reply(fmt.Sprintf("An error has occurred in option '%s'. Error: %s", errors[0].Opt.Name, errors[0].Err))
		return
	}

	cmd.Run(ctx, opts)
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

func findCommand(cmds []Command, args []string, depth int) (Command, []string, error) {
	for _, cmd := range cmds {
		if cmd.Name == args[0] || slices.Contains(cmd.Aliases, args[0]) {
			if len(cmd.Subs) > 0 {
				if len(args) > 1 {
					depth++
					findCommand(cmd.Subs, args[1:], depth)
				} else {
					return Command{}, nil, errors.New("subcommand expected but not provided")
				}
			} else {
				return cmd, args[1:], nil
			}
		}
	}

	if depth != 1 {
		return Command{}, nil, fmt.Errorf("subcommand '%s' does not exist", args[0])
	}

	return Command{}, nil, errNotFound
}

func findSlashCommandSubCommand(cmds []Command, i *discordgo.InteractionCreate) (Command, error) {
	d := i.ApplicationCommandData()
	var name string
	if len(d.Options) == 0 || (d.Options[0].Type != discordgo.ApplicationCommandOptionSubCommandGroup && d.Options[0].Type != discordgo.ApplicationCommandOptionSubCommand) {
		name = d.Name
	} else if d.Options[0].Type == discordgo.ApplicationCommandOptionSubCommandGroup {
		name = d.Options[0].Options[0].Name
	} else {
		name = d.Options[0].Name
	}
	for _, cmd := range cmds {
		if cmd.Name == name {
			return cmd, nil
		}
	}

	return Command{}, errNotFound
}
