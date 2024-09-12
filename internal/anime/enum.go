package anime

import (
	"encoding/json"
	"fmt"

	"github.com/zanz1n/duvua/internal/errors"
)

type AnimeAgeRating uint8

const (
	AnimeAgeRatingUnknown AnimeAgeRating = iota
	AnimeAgeRatingGeneral
	AnimeAgeRatingParental
	AnimeAgeRatingRestricted
	AnimeAgeRatingExplicit
)

var (
	testAnimeAgeRating = AnimeAgeRating(0)

	_ json.Marshaler   = &testAnimeAgeRating
	_ json.Unmarshaler = &testAnimeAgeRating
	_ fmt.Stringer     = &testAnimeAgeRating
)

func (ar AnimeAgeRating) StringPtBr() string {
	switch ar {
	case AnimeAgeRatingGeneral:
		return "Livre"
	case AnimeAgeRatingParental:
		return "10"
	case AnimeAgeRatingRestricted:
		return "16"
	case AnimeAgeRatingExplicit:
		return "18"
	default:
		return "Unknown"
	}
}

// String implements fmt.Stringer.
func (ar *AnimeAgeRating) String() string {
	switch *ar {
	case AnimeAgeRatingGeneral:
		return "G"
	case AnimeAgeRatingParental:
		return "PG"
	case AnimeAgeRatingRestricted:
		return "R"
	case AnimeAgeRatingExplicit:
		return "R18"
	default:
		return "Unknown"
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (ar *AnimeAgeRating) UnmarshalJSON(b []byte) error {
	s := string(b)
	switch s {
	case `"G"`:
		*ar = AnimeAgeRatingGeneral
	case `"PG"`:
		*ar = AnimeAgeRatingParental
	case `"R"`:
		*ar = AnimeAgeRatingRestricted
	case `"R18"`:
		*ar = AnimeAgeRatingExplicit
	default:
		return errors.Unexpectedf("unknown age rating `%s`", s)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (ar *AnimeAgeRating) MarshalJSON() ([]byte, error) {
	return []byte("\"" + ar.String() + "\""), nil
}

type AnimeSubtype uint8

const (
	AnimeSubtypeUnknown AnimeSubtype = iota
	AnimeSubtypeONA
	AnimeSubtypeOVA
	AnimeSubtypeTV
	AnimeSubtypeMovie
	AnimeSubtypeMusic
	AnimeSubtypeSpecial
)

var (
	testAnimeSubtype = AnimeSubtype(0)

	_ json.Marshaler   = &testAnimeSubtype
	_ json.Unmarshaler = &testAnimeSubtype
	_ fmt.Stringer     = &testAnimeSubtype
)

func (st AnimeSubtype) StringPtBr() string {
	switch st {
	case AnimeSubtypeONA:
		return "ONA"
	case AnimeSubtypeOVA:
		return "OVA"
	case AnimeSubtypeTV:
		return "TV"
	case AnimeSubtypeMovie:
		return "Filme"
	case AnimeSubtypeMusic:
		return "Música"
	case AnimeSubtypeSpecial:
		return "Especial"
	default:
		return "Desconhecido"
	}
}

// String implements fmt.Stringer.
func (st *AnimeSubtype) String() string {
	switch *st {
	case AnimeSubtypeONA:
		return "ONA"
	case AnimeSubtypeOVA:
		return "OVA"
	case AnimeSubtypeTV:
		return "TV"
	case AnimeSubtypeMovie:
		return "movie"
	case AnimeSubtypeMusic:
		return "music"
	case AnimeSubtypeSpecial:
		return "special"
	default:
		return "unknown"
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (st *AnimeSubtype) UnmarshalJSON(b []byte) error {
	s := string(b)
	switch s {
	case `"ONA"`:
		*st = AnimeSubtypeONA
	case `"OVA"`:
		*st = AnimeSubtypeOVA
	case `"TV"`:
		*st = AnimeSubtypeTV
	case `"movie"`:
		*st = AnimeSubtypeMovie
	case `"music"`:
		*st = AnimeSubtypeMusic
	case `"special"`:
		*st = AnimeSubtypeSpecial
	default:
		return errors.Unexpectedf("unknown subtype `%s`", s)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (st *AnimeSubtype) MarshalJSON() ([]byte, error) {
	return []byte("\"" + st.String() + "\""), nil
}

type AnimeStatus uint8

const (
	AnimeStatusUnknown = iota
	AnimeStatusCurrent
	AnimeStatusFinished
	AnimeStatusTba
	AnimeStatusUnreleased
	AnimeStatusUpcoming
)

var (
	testAnimeStatus = AnimeStatus(0)

	_ json.Marshaler   = &testAnimeStatus
	_ json.Unmarshaler = &testAnimeStatus
	_ fmt.Stringer     = &testAnimeStatus
)

func (as AnimeStatus) StringPtBr() string {
	switch as {
	case AnimeStatusCurrent:
		return "Atual"
	case AnimeStatusFinished:
		return "Terminado"
	case AnimeStatusTba:
		return "Não anunciado"
	case AnimeStatusUnreleased:
		return "Não lançado"
	case AnimeStatusUpcoming:
		return "Por vir"
	default:
		return "Desconhecido"
	}
}

// String implements fmt.Stringer.
func (as *AnimeStatus) String() string {
	switch *as {
	case AnimeStatusCurrent:
		return "current"
	case AnimeStatusFinished:
		return "finished"
	case AnimeStatusTba:
		return "tba"
	case AnimeStatusUnreleased:
		return "unreleased"
	case AnimeStatusUpcoming:
		return "upcoming"
	default:
		return "unknown"
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (as *AnimeStatus) UnmarshalJSON(b []byte) error {
	s := string(b)
	switch s {
	case `"current"`:
		*as = AnimeStatusCurrent
	case `"finished"`:
		*as = AnimeStatusFinished
	case `"tba"`:
		*as = AnimeStatusTba
	case `"unreleased"`:
		*as = AnimeStatusUnreleased
	case `"upcoming"`:
		*as = AnimeStatusUpcoming
	default:
		return errors.Unexpectedf("unknown age rating `%s`", s)
	}
	return nil
}

// MarshalJSON implements json.Marshaler.
func (as *AnimeStatus) MarshalJSON() ([]byte, error) {
	return []byte("\"" + as.String() + "\""), nil
}
