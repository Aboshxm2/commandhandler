package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/Aboshxm2/commandhandler"
	"github.com/bwmarrin/discordgo"
)

type ContainsDigit struct{}

func (ContainsDigit) Test(value any) error {
	if regexp.MustCompile(`\d`).MatchString(value.(string)) {
		return nil
	}

	return errors.New("value must contain a digit")
}

func initCommands(s *discordgo.Session) {
	const prefix = "!"

	cmds := []commandhandler.Command{
		{
			Name:        "custom_rule",
			Description: "A command that accepts an option with custom rules",
			Options: []commandhandler.Option{
				{
					Name:        "username",
					Description: "This option must contain a digit",
					Type:        commandhandler.StringOptionType,
					Required:    true,
					Rules: []commandhandler.Rule{
						ContainsDigit{},
					},
				},
			},
			Run: func(ctx commandhandler.Context, opts map[string]any) {
				ctx.Reply("The username is " + opts["username"].(string))
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
