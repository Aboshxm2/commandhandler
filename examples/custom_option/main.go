package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/Aboshxm2/commandhandler"
	"github.com/bwmarrin/discordgo"
)

func initCommands(s *discordgo.Session) {
	const prefix = "!"

	cmds := []commandhandler.Command{
		{
			Name:        "custom_option",
			Description: "A command that accepts a custom option",
			Options: []commandhandler.Option{
				{
					Name:        "url",
					Description: "My custom option",
					Required:    true,
					Type:        urlOptionType,
				},
			},
			Run: func(ctx commandhandler.Context, opts map[string]any) {
				url := opts["url"].(*url.URL)
				ctx.Reply(fmt.Sprintf("Scheme: %s\nHost: %s\nPath: %s", url.Scheme, url.Host, url.Path))
			},
		},
	}

	messageResolvers := commandhandler.DefaultMessageResolvers()
	slashCommandResolvers := commandhandler.DefaultSlashCommandResolvers()

	messageResolvers[urlOptionType] = urlResolver
	slashCommandResolvers[urlOptionType] = slashCommandUrlResolver

	resolver := commandhandler.SimpleResolver{
		MessageResolvers:      messageResolvers,
		SlashCommandResolvers: slashCommandResolvers,
	}
	handler := commandhandler.NewHandler(prefix, cmds, resolver)

	s.AddHandler(handler.OnMessageCreate)
	s.AddHandler(handler.OnInteractionCreate)

	optionsTypeMap := commandhandler.DefaultOptionsTypeMap()
	optionsTypeMap[urlOptionType] = discordgo.ApplicationCommandOptionString

	builder := commandhandler.SimpleBuilder{
		OptionsTypeMap: optionsTypeMap,
	}

	for _, cmd := range cmds {
		_, err := s.ApplicationCommandCreate(s.State.Application.ID, *guildId, builder.Build(cmd))
		if err != nil {
			fmt.Println("error creating discord command,", err)
		}
	}
}

var (
	guildId = flag.String("guild", "", "Register commands in specific guild. If not passed register globally")
	token   = flag.String("token", "", "Bot token")
)

func init() {
	flag.Parse()
}

func main() {
	dg, err := discordgo.New("Bot " + *token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.Identify.Intents = discordgo.IntentsGuildMessages

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	initCommands(dg)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

const urlOptionType commandhandler.OptionType = 20

func urlResolver(ctx commandhandler.Context, arg string) (any, error) {
	return url.Parse(arg)
}

func slashCommandUrlResolver(ctx commandhandler.Context, arg discordgo.ApplicationCommandInteractionDataOption) (any, error) {
	return url.Parse(arg.StringValue())
}
