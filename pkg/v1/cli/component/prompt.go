package component

import (
	"github.com/AlecAivazis/survey/v2"
)

// PromptConfig is the configuration for a prompt.
type PromptConfig struct {
	// Message to display to user.
	Message string

	// Default option.
	Default string

	// Sensitive information.
	Sensitive bool

	// Help for the prompt.
	Help string
}

// Run the prompt.
func (p *PromptConfig) Run(response interface{}) error {
	return Prompt(p, response)
}

// Prompt from input.
func Prompt(p *PromptConfig, response interface{}) error {
	prompt := translatePromptConfig(p)
	return survey.AskOne(prompt, response)
}

func translatePromptConfig(p *PromptConfig) survey.Prompt {
	if p.Sensitive {
		return &survey.Password{
			Message: p.Message,
			Help:    p.Help,
		}
	}
	return &survey.Input{
		Message: p.Message,
		Default: p.Default,
		Help:    p.Help,
	}
}
