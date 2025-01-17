package commandhandler

import (
	"github.com/bwmarrin/discordgo"
)

func DefaultOptionsTypeMap() map[OptionType]discordgo.ApplicationCommandOptionType {
	return map[OptionType]discordgo.ApplicationCommandOptionType{
		StringOptionType:  discordgo.ApplicationCommandOptionString,
		IntegerOptionType: discordgo.ApplicationCommandOptionInteger,
		FloatOptionType:   discordgo.ApplicationCommandOptionNumber,
		BooleanOptionType: discordgo.ApplicationCommandOptionBoolean,
		UserOptionType:    discordgo.ApplicationCommandOptionUser,
		MemberOptionType:  discordgo.ApplicationCommandOptionUser,
		ChannelOptionType: discordgo.ApplicationCommandOptionChannel,
		RoleOptionType:    discordgo.ApplicationCommandOptionRole,
	}
}

type Builder interface {
	Build(cmd Command) *discordgo.ApplicationCommand
}

func NewBuilder() Builder {
	return SimpleBuilder{DefaultOptionsTypeMap()}
}

type SimpleBuilder struct {
	OptionsTypeMap map[OptionType]discordgo.ApplicationCommandOptionType
}

func (b SimpleBuilder) buildOption(opt Option) *discordgo.ApplicationCommandOption {
	o := discordgo.ApplicationCommandOption{
		Name:        opt.Name,
		Description: opt.Description,
		Type:        b.OptionsTypeMap[opt.Type],
		Required:    opt.Required,
	}

	for _, c := range opt.Choices {
		o.Choices = append(o.Choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  c.Name,
			Value: c.Value,
		})
	}

	for _, rule := range opt.Rules {
		switch r := rule.(type) {
		case MaxInt:
			o.MaxValue = float64(r.Max)
		case MaxFloat:
			o.MaxValue = r.Max
		case MaxString:
			o.MaxLength = r.Max
		case MinInt:
			v := float64(r.Min)
			o.MinValue = &v
		case MinFloat:
			v := r.Min
			o.MinValue = &v
		case MinString:
			v := r.Min
			o.MinLength = &v
		}
	}

	return &o
}

func (b SimpleBuilder) Build(cmd Command) *discordgo.ApplicationCommand {
	command := discordgo.ApplicationCommand{
		Name:        cmd.Name,
		Description: cmd.Description,
	}

	if len(cmd.Subs) > 0 {
		if len(cmd.Subs[0].Subs) > 0 {
			subs := []*discordgo.ApplicationCommandOption{}

			for _, sub := range cmd.Subs {
				subsOfSub := []*discordgo.ApplicationCommandOption{}

				for _, subOfSub := range sub.Subs {
					opts := []*discordgo.ApplicationCommandOption{}
					for _, opt := range subOfSub.Options {
						opts = append(opts, b.buildOption(opt))
					}

					subsOfSub = append(subsOfSub, &discordgo.ApplicationCommandOption{
						Type:        discordgo.ApplicationCommandOptionSubCommand,
						Name:        subOfSub.Name,
						Description: subOfSub.Description,
						Options:     opts,
					})
				}

				subs = append(subs, &discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionSubCommandGroup,
					Name:        sub.Name,
					Description: sub.Description,
					Options:     subsOfSub,
				})
			}

			command.Options = subs
		} else {
			subs := []*discordgo.ApplicationCommandOption{}

			for _, sub := range cmd.Subs {
				opts := []*discordgo.ApplicationCommandOption{}

				for _, opt := range sub.Options {
					opts = append(opts, b.buildOption(opt))
				}

				subs = append(subs, &discordgo.ApplicationCommandOption{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        sub.Name,
					Description: sub.Description,
					Options:     opts,
				})
			}

			command.Options = subs
		}
	} else {
		opts := []*discordgo.ApplicationCommandOption{}

		for _, opt := range cmd.Options {
			opts = append(opts, b.buildOption(opt))
		}

		command.Options = opts
	}

	return &command
}
