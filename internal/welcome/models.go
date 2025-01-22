package welcome

import (
	"strconv"
	"time"
)

const (
	DefaultEnabled bool        = false
	DefaultMessage string      = "Seja Bem Vind@ ao servidor {{USER}}"
	DefaultKind    WelcomeType = WelcomeTypeMessage
)

type WelcomeType string

const (
	WelcomeTypeMessage WelcomeType = "Message"
	WelcomeTypeImage   WelcomeType = "Image"
	WelcomeTypeEmbed   WelcomeType = "Embed"
)

func (w WelcomeType) StringPtBr() (s string) {
	switch w {
	case WelcomeTypeMessage:
		s = "Mensagem"
	case WelcomeTypeImage:
		s = "Imagem"
	case WelcomeTypeEmbed:
		s = "Embed"
	default:
		s = "Inválido"
	}
	return
}

func WelcomeTypeFromString(s string) (wt WelcomeType, ok bool) {
	wt = WelcomeType(s)

	switch wt {
	case WelcomeTypeMessage, WelcomeTypeImage, WelcomeTypeEmbed:
		ok = true
	default:
		ok = false
	}
	return
}

type Welcome struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Enabled   bool
	ChannelId *string
	// The message template
	Message string
	Kind    WelcomeType
}

func atoi(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func itoa(i int64) string {
	return strconv.FormatInt(i, 10)
}
