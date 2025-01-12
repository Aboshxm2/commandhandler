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

	registerCommands(dg)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	dg.Close()
}

func registerCommands(s *discordgo.Session) {
	const prefix = "!"

	cmds := []commandhandler.Command{
		{
			Name:        "ping",
			Description: "Simple pingpong command",
			Run: func(ctx commandhandler.Context, opts map[string]any) {
				err := ctx.Reply("pong")

				if err != nil {
					fmt.Println("Cannot send message. Error: ", err)
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
		s.ApplicationCommandCreate(s.State.Application.ID, *guildId, builder.Build(cmd))
	}
}
