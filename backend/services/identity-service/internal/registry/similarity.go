package registry

import (
	"strings"
	"unicode"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
)

const (
	MinimumNameSimilarity = 0.85
	trigramSize           = 3
)

func normalise(name string) string {
	var builder strings.Builder
	builder.Grow(len(name))

	previousWasSpace := true
	for _, character := range strings.ToLower(name) {
		switch {
		case unicode.IsLetter(character) || unicode.IsDigit(character):
			builder.WriteRune(character)
			previousWasSpace = false
		case !previousWasSpace:
			builder.WriteRune(' ')
			previousWasSpace = true
		}
	}

	return strings.TrimSpace(builder.String())
}

func NameSimilarity(declared, registered string) float64 {
	left, right := normalise(declared), normalise(registered)
	if left == "" || right == "" {
		return 0
	}
	if left == right {
		return 1
	}

	dice := metrics.NewSorensenDice()
	dice.NgramSize = trigramSize

	return strutil.Similarity(left, right, dice)
}
