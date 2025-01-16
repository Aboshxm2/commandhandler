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
		return errors.New("value is required but was not provided")
	}

	return nil
}

type Max struct {
	Max int64
}

func (r Max) Test(value any) error {
	switch v := value.(type) {
	case int64:
		if v <= r.Max {
			return nil
		}
	case float64:
		if int64(v) < r.Max {
			return nil
		}
	case string:
		if len(v) < int(r.Max) {
			return nil
		}
	default:
		panic("Value type should be int, float64 or string")
	}
	return fmt.Errorf("value exceeds the maximum allowed of %d", r.Max)
}

type Min struct {
	Min int64
}

func (r Min) Test(value any) error {
	switch v := value.(type) {
	case int64:
		if v >= r.Min {
			return nil
		}
	case float64:
		if int64(v) > r.Min {
			return nil
		}
	case string:
		if len(v) > int(r.Min) {
			return nil
		}
	default:
    fmt.Printf("%T\n", v)
		panic("Value type should be int, float64 or string")
	}
	return fmt.Errorf("value is less than the minimum allowed of %d", r.Min)
}

type Uppercase struct{}

func (Uppercase) Test(value any) error {
	if value.(string) == strings.ToUpper(value.(string)) {
		return nil
	}
	return errors.New("value must be uppercase")
}

type Lowercase struct{}

func (Lowercase) Test(value any) error {
	if value.(string) == strings.ToLower(value.(string)) {
		return nil
	}
	return errors.New("value must be lowercase")
}

type ChannelType struct {
	Types []discordgo.ChannelType
}

func (r ChannelType) Test(value any) error {
	if slices.Contains(r.Types, value.(*discordgo.Channel).Type) {
		return nil
	}
	return fmt.Errorf("channel type '%v' is not allowed", value.(*discordgo.Channel).Type)
}

func Validate(opts []Option, values map[string]any) (errors []struct {
	Opt Option
	Err error
}) {
	for _, opt := range opts {
		for _, rule := range opt.Rules {
			if values[opt.Name] == nil {
				continue
			}
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
