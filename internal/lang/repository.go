package lang

type Translator interface {
	Translate(srcl, dstl Language, source string) (string, error)
}
