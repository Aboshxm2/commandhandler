package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Aboshxm2/commandhandler"
	"github.com/bwmarrin/discordgo"
)

func initCommands(s *discordgo.Session) {
	const prefix = "!"

	cmds := []commandhandler.Command{
		{
			Name:        "subcommands",
			Description: "A command that have 2 sub commands",
			Subs: []commandhandler.Command{
				{
					Name:        "sub_of_sub",
					Description: "ss",
					Subs: []commandhandler.Command{
						{
							Name:        "deep",
							Description: "deep",
							Options: []commandhandler.Option{
								{
									Name:        "first",
									Description: "first",
									Type:        commandhandler.IntegerOptionType,
								},
							},
							Run: func(ctx commandhandler.Context, opts map[string]any) {
								if v, ok := opts["first"]; ok {
									ctx.Reply(strconv.FormatInt(v.(int64), 10))
								} else {
									ctx.Reply("no opt given")
								}
							},
						},
					},
				},
				{
					Name:        "reverse",
					Description: "Reverse whatever text you pass",
					Options: []commandhandler.Option{
						{
							Name:        "text",
							Description: "x",
							Type:        commandhandler.StringOptionType,
							Required:    true,
						},
					},
					Run: func(ctx commandhandler.Context, opts map[string]any) {
						fmt.Println(opts)
						ctx.Reply("reverse " + opts["text"].(string))
					},
				},
				{
					Name:        "else",
					Description: "else description",
					Options: []commandhandler.Option{
						{
							Name:        "text",
							Description: "some desc",
							Type:        commandhandler.StringOptionType,
							Required:    true,
						},
					},
					Run: func(ctx commandhandler.Context, opts map[string]any) {
						ctx.Reply("else " + opts["text"].(string))
					},
				},
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
