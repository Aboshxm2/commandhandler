package main

import (
	"flag"
	"fmt"
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
			Name:        "rules",
			Description: "A command that accepts 2 options with some rules",
			Options: []commandhandler.Option{
				{
					Name:        "option1",
					Description: "First Option",
					Type:        commandhandler.IntegerOptionType,
					Required:    true,
					Rules: []commandhandler.Rule{
						commandhandler.MinInt{Min: 2},
						commandhandler.MaxInt{Max: 9},
					},
				},
				{
					Name:        "option2",
					Description: "Second Option",
					Type:        commandhandler.StringOptionType,
					Rules: []commandhandler.Rule{
						commandhandler.MaxString{Max: 10},
					},
				},
			},
			Run: func(ctx commandhandler.Context, opts map[string]any) {
				opt1 := opts["option1"]
				opt2, ok := opts["option2"]

				if ok {
					ctx.Reply(fmt.Sprintf("Option 1: %v, Option 2: %v", opt1, opt2))
				} else {
					ctx.Reply(fmt.Sprintf("Option 1: %v, Option 2 was not provided", opt1))
				}
			},
		},
	}

	resolver := commandhandler.NewResolver()
	handler := commandhandler.NewHandler(prefix, cmds, resolver)

	s.AddHandler(handler.OnMessageCreate)
	s.AddHandler(handler.OnInteractionCreate)

	builder := commandhandler.NewBuilder()
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
