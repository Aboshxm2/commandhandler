package commandhandler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type MessageResolver func(ctx Context, arg string) (any, error)
type SlashCommandResolver func(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error)

type Resolver interface {
	ResolveMessageOptions(cmd Command, ctx Context, args []string) (map[string]any, error)
	ResolveSlashCommandOptions(cmd Command, ctx Context, args discordgo.ApplicationCommandInteractionData) (map[string]any, error)
}

func NewResolver() Resolver {
	return SimpleResolver{
		MessageResolvers:      DefaultMessageResolvers(),
		SlashCommandResolvers: DefaultSlashCommandResolvers(),
	}
}

type SimpleResolver struct {
	MessageResolvers      map[OptionType]MessageResolver
	SlashCommandResolvers map[OptionType]SlashCommandResolver
}

func (r SimpleResolver) ResolveMessageOptions(cmd Command, ctx Context, args []string) (map[string]any, error) {
	opts := map[string]any{}
	for i, opt := range cmd.Options {
		if len(args)-1 < i {
			opts[opt.Name] = nil
			continue
		}

		arg := args[i]

		if len(opt.Choices) > 0 {
			var err error
			arg, err = resolveMessageOptionChoices(opt, arg)
			if err != nil {
				return nil, err
			}
		}

		if resolver, ok := r.MessageResolvers[opt.Type]; ok {
			v, err := resolver(ctx, arg)
			if err != nil {
				return nil, fmt.Errorf("failed to resolve option '%s': %w", opt.Name, err)
			}

			opts[opt.Name] = v
		} else {
			return nil, fmt.Errorf("no resolver found for option type '%v'", opt.Type)
		}
	}
	return opts, nil
}

func resolveMessageOptionChoices(opt Option, arg string) (string, error) {
	for _, c := range opt.Choices {
		if c.Name == arg {
			switch v := c.Value.(type) {
			case int64:
				arg = strconv.Itoa(v)
			case float64:
				arg = strconv.FormatFloat(v, 'f', -1, 64)
			case string:
				arg = v
			default:
				panic(fmt.Sprintf("choice value cannot be of type %T", v))
			}
			return arg, nil
		}
	}

	return "", fmt.Errorf("value '%v' is not a valid choice", arg)
}

func (r SimpleResolver) ResolveSlashCommandOptions(cmd Command, ctx Context, args discordgo.ApplicationCommandInteractionData) (map[string]any, error) {
	opts := map[string]any{}
	for _, opt := range cmd.Options {
		var found *discordgo.ApplicationCommandInteractionDataOption
		for _, arg := range args.Options {
			if arg.Name == opt.Name {
				found = arg
			}
		}
		if found != nil {
			if resolver, ok := r.SlashCommandResolvers[opt.Type]; ok {
				v, err := resolver(ctx, *found)
				if err != nil {
					return nil, fmt.Errorf("failed to resolve option '%s': %w", opt.Name, err)
				}
				opts[opt.Name] = v
			} else {
				return nil, fmt.Errorf("no resolver found for option type '%v'", opt.Type)
			}
		} else {
			opts[opt.Name] = nil
		}
	}

	return opts, nil
}

func DefaultMessageResolvers() map[OptionType]MessageResolver {
	return map[OptionType]MessageResolver{
		StringOptionType:  stringResolver,
		IntegerOptionType: integerResolver,
		FloatOptionType:   floatResolver,
		BooleanOptionType: booleanResolver,
		UserOptionType:    userResolver,
		MemberOptionType:  memberResolver,
		ChannelOptionType: channelResolver,
		RoleOptionType:    roleResolver,
	}
}

func DefaultSlashCommandResolvers() map[OptionType]SlashCommandResolver {
	return map[OptionType]SlashCommandResolver{
		StringOptionType:  slashCommandStringResolver,
		IntegerOptionType: slashCommandIntegerResolver,
		FloatOptionType:   slashCommandFloatResolver,
		BooleanOptionType: slashCommandBooleanResolver,
		UserOptionType:    slashCommandUserResolver,
		MemberOptionType:  slashCommandMemberResolver,
		ChannelOptionType: slashCommandChannelResolver,
		RoleOptionType:    slashCommandRoleResolver,
	}
}

func integerResolver(ctx Context, arg string) (any, error) {
	return strconv.ParseInt(arg, 10, 64)
}

func floatResolver(ctx Context, arg string) (any, error) {
	return strconv.ParseFloat(arg, 64)
}

func stringResolver(ctx Context, arg string) (any, error) {
	return arg, nil
}

func booleanResolver(ctx Context, arg string) (any, error) {
	return strconv.ParseBool(arg)
}

func userResolver(ctx Context, arg string) (any, error) {
	if strings.HasPrefix(arg, "<@") && strings.HasSuffix(arg, ">") {
		arg = arg[2 : len(arg)-1]
	}
	v, err := ctx.Session().State.Member(ctx.GuildId(), arg)
	if err != nil {
		return ctx.Session().User(arg)
	}
	return v.User, nil
}

func memberResolver(ctx Context, arg string) (any, error) {
	if strings.HasPrefix(arg, "<@") && strings.HasSuffix(arg, ">") {
		arg = arg[2 : len(arg)-1]
	}
	v, err := ctx.Session().State.Member(ctx.GuildId(), arg)
	if err != nil {
		return ctx.Session().GuildMember(ctx.GuildId(), arg)
	}
	return v, nil
}

func channelResolver(ctx Context, arg string) (any, error) {
	if strings.HasPrefix(arg, "<#") && strings.HasSuffix(arg, ">") {
		arg = arg[2 : len(arg)-1]
	}
	v, err := ctx.Session().State.Channel(arg)
	if err != nil {
		return ctx.Session().Channel(arg)
	}
	return v, nil
}

func roleResolver(ctx Context, arg string) (any, error) {
	if strings.HasPrefix(arg, "<@&") && strings.HasSuffix(arg, ">") {
		arg = arg[3 : len(arg)-1]
	}
	v, err := ctx.Session().State.Role(ctx.GuildId(), arg)
	if err != nil {
		roles, err := ctx.Session().GuildRoles(ctx.GuildId())
		if err != nil {
			return nil, err
		}

		for _, role := range roles {
			if role.ID == arg {
				return role, nil
			}
		}
		return nil, fmt.Errorf("role with ID '%s' not found in guild", arg)
	}
	return v, nil
}

func slashCommandIntegerResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return arg.IntValue(), nil
}

func slashCommandFloatResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return arg.FloatValue(), nil
}

func slashCommandStringResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return arg.StringValue(), nil
}

func slashCommandBooleanResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return arg.BoolValue(), nil
}


func slashCommandUserResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return userResolver(ctx, arg.Value.(string))
}

func slashCommandMemberResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return memberResolver(ctx, arg.Value.(string))
}

func slashCommandChannelResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return channelResolver(ctx, arg.Value.(string))
}

func slashCommandRoleResolver(ctx Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return roleResolver(ctx, arg.Value.(string))
}
