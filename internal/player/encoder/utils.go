package encoder

import "io"

type pipeReadCloser struct {
	io.Reader
}

func NewPipeReadCloser(r io.Reader) io.ReadCloser {
	return &pipeReadCloser{r}
}

func (p *pipeReadCloser) Close() error {
	return nil
}
