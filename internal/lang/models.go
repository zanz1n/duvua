package lang

import "fmt"

type Language uint8

const (
	UnknownLanguage = iota
	English
	Portuguese
	Spanish
	Japanese
	Chinese
)

var (
	testLanguage = Language(0)

	_ fmt.Stringer = &testLanguage
)

// String implements fmt.Stringer.
func (l Language) String() string {
	switch l {
	case English:
		return "en"
	case Portuguese:
		return "pt"
	case Spanish:
		return "es"
	case Japanese:
		return "ja"
	case Chinese:
		return "zh-cn"
	default:
		return "auto"
	}
}
