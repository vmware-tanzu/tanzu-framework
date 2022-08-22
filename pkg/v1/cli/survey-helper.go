package cli

import (
	"github.com/AlecAivazis/survey/v2"
	"os"
)

func SurveyOptions() (surveyOptions survey.AskOpt) {
	surveyOptions = survey.WithIcons(func(icons *survey.IconSet) {
		icons.Question.Format = "cyan+b"
	})
	return
}

func WithStdioOptions() (surveyOptions survey.AskOpt) {
	surveyOptions = survey.WithStdio(os.Stdin, os.Stderr, os.Stderr)
	return
}
