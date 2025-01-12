package commandhandler

type Command struct {
	Name        string
	Description string
	Aliases     []string
	Subs        []Command
	Options     []Option
	Run         func(ctx Context, opts map[string]any)
}
