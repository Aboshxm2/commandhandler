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

type MaxInt struct {
	Max int64
}

func (r MaxInt) Test(value any) error {
	if value.(int64) <= r.Max {
		return nil
	}
	return fmt.Errorf("value exceeds the maximum allowed of %d", r.Max)
}

type MaxFloat struct {
	Max float64
}

func (r MaxFloat) Test(value any) error {
	if value.(float64) <= r.Max {
		return nil
	}
	return fmt.Errorf("value exceeds the maximum allowed of %.2f", r.Max)
}

type MaxString struct {
	Max int
}

func (r MaxString) Test(value any) error {
	if len(value.(string)) <= r.Max {
		return nil
	}
	return fmt.Errorf("string length exceeds the maximum allowed of %d", r.Max)
}

type MinInt struct {
	Min int64
}

func (r MinInt) Test(value any) error {
	if value.(int64) >= r.Min {
		return nil
	}
	return fmt.Errorf("value is less than the minimum allowed of %d", r.Min)
}

type MinFloat struct {
	Min float64
}

func (r MinFloat) Test(value any) error {
	if value.(float64) >= r.Min {
		return nil
	}
	return fmt.Errorf("value is less than the minimum allowed of %.2f", r.Min)
}

type MinString struct {
	Min int
}

func (r MinString) Test(value any) error {
	if len(value.(string)) >= r.Min {
		return nil
	}
	return fmt.Errorf("string length is less than the minimum allowed of %d", r.Min)
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
