package commandhandler

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Rule interface {
	Test(value any) error
}

type Required struct{}

func (Required) Test(value any) error {
	if value == nil {
		return errors.New("")
	}

	return nil
}

type Choice struct {
	Name  string
	Value string
}

type Choices struct {
	Choices []Choice
}

func (r Choices) Test(value any) error {
	for _, c := range r.Choices {
		if c.Value == value {
			return nil
		}
	}
	return errors.New("")
}

type Max struct {
	Max int
}

func (r Max) Test(value any) error {
	switch v := value.(type) {
	case int:
		if v < r.Max {
			return nil
		}
	case float64:
		if int(v) < r.Max {
			return nil
		}
	case string:
		if len(v) < r.Max {
			return nil
		}
	default:
		panic("Value type should be int, float64 or string")
	}
	return fmt.Errorf("Value is greater than %d", r.Max)
}

type Min struct {
	Min int
}

func (r Min) Test(value any) error {
	switch v := value.(type) {
	case int:
		if v > r.Min {
			return nil
		}
	case float64:
		if int(v) > r.Min {
			return nil
		}
	case string:
		if len(v) > r.Min {
			return nil
		}
	default:
		panic("Value type should be int, float64 or string")
	}
	return fmt.Errorf("Value is smaller than %d", r.Min)
}

type Uppercase struct{}

func (Uppercase) Test(value any) bool {
	return value.(string) == strings.ToUpper(value.(string))
}

type Lowercase struct{}

func (Lowercase) Test(value any) bool {
	return value.(string) == strings.ToLower(value.(string))
}

type ChannelType struct {
	Types []discordgo.ChannelType
}

func (r ChannelType) Test(value any) bool {
	return slices.Contains(r.Types, value.(*discordgo.Channel).Type)
}

func Validate(opts []Option, values map[string]any) (errors []struct {
	Opt Option
	Err error
}) {
	for _, opt := range opts {
		for _, rule := range opt.Rules {
			if err := rule.Test(values[opt.Name]); err != nil {
				errors = append(errors, struct {
					Opt Option
					Err error
				}{opt, err})
			}
		}
	}
	return
}
