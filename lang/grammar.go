package lang

import (
	"fmt"
	"unicode/utf8"

	"github.com/jorgensigvardsson/gomud/absmachine"
)

func IsVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	default:
		return false
	}
}

func IndefiniteArticleFor(noun string) string {
	if noun == "" {
		// All bets are off...
		return "a"
	}

	r, _ := utf8.DecodeRuneInString(noun)

	if IsVowel(r) {
		return "an"
	}

	return "a"
}

func DirectionName(direction absmachine.Direction) string {
	switch direction {
	case absmachine.DIR_NORTH:
		return "North"
	case absmachine.DIR_SOUTH:
		return "South"
	case absmachine.DIR_UP:
		return "Up"
	case absmachine.DIR_DOWN:
		return "Down"
	case absmachine.DIR_EAST:
		return "East"
	case absmachine.DIR_WEST:
		return "West"
	default:
		panic(fmt.Sprintf("Unknown direction %v", direction))
	}
}
