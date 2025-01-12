package commandhandler

type OptionType uint8

const (
	StringOptionType  OptionType = 0
	IntegerOptionType OptionType = 1
	FloatOptionType   OptionType = 2
	BooleanOptionType OptionType = 3
	UserOptionType    OptionType = 4
	MemberOptionType  OptionType = 5
	ChannelOptionType OptionType = 6
	RoleOptionType    OptionType = 7
)

type Option struct {
	Name        string
	Type        OptionType
	Description string
	Rules       []Rule
}
