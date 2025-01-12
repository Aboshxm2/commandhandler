package commandhandler

import (
	"github.com/bwmarrin/discordgo"
)

type Context interface {
	Session() *discordgo.Session
	GuildId() string
	ChannelId() string
	Member() *discordgo.Member
	Reply(content string) error
}

type MessageContext struct {
	s *discordgo.Session
	m *discordgo.Message
}

func (ctx MessageContext) Session() *discordgo.Session { return ctx.s }

func (ctx MessageContext) GuildId() string { return ctx.m.GuildID }

func (ctx MessageContext) ChannelId() string { return ctx.m.ChannelID }

func (ctx MessageContext) Member() *discordgo.Member { return ctx.m.Member }

func (ctx MessageContext) Reply(content string) error {
	_, err := ctx.s.ChannelMessageSendReply(ctx.ChannelId(), content, ctx.m.Reference())
	return err
}

func (ctx MessageContext) Message() *discordgo.Message { return ctx.m }

type SlashCommandContext struct {
	s *discordgo.Session
	i *discordgo.Interaction
}

func (ctx SlashCommandContext) Session() *discordgo.Session { return ctx.s }

func (ctx SlashCommandContext) GuildId() string { return ctx.i.GuildID }

func (ctx SlashCommandContext) ChannelId() string { return ctx.i.ChannelID }

func (ctx SlashCommandContext) Member() *discordgo.Member { return ctx.i.Member }

func (ctx SlashCommandContext) Reply(content string) error {
	return ctx.s.InteractionRespond(ctx.i, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

func (ctx SlashCommandContext) Interaction() *discordgo.Interaction { return ctx.i }

func MessageToContext(s *discordgo.Session, m *discordgo.Message) Context {
	return &MessageContext{s, m}
}

func SlashCommandToContext(s *discordgo.Session, i *discordgo.InteractionCreate) Context {
	return &SlashCommandContext{s, i.Interaction}
}
