package absmachine

import (
	"strings"

	"github.com/jorgensigvardsson/gomud/logging"
)

type SimpleVerbMobAction struct {
	Verb        string
	Preposition string
}

func (action *SimpleVerbMobAction) Run(mob *Mob, logger logging.Logger) (output string, err error) {
	outputBuilder := strings.Builder{}

	outputBuilder.WriteString(mob.Name)
	outputBuilder.WriteString(" ")
	if strings.HasSuffix(action.Verb, "ing") {
		outputBuilder.WriteString("is ")
	}

	outputBuilder.WriteString(action.Verb)

	if action.Preposition != "" {
		outputBuilder.WriteString(" ")
		outputBuilder.WriteString(action.Preposition)
	}

	outputBuilder.WriteString(".")

	return outputBuilder.String(), nil
}
