package parser

import (
	"io"
	"sync/atomic"
)

type countingReader struct {
	r io.Reader
	n atomic.Int64
}

func (c *countingReader) Read(b []byte) (int, error) {
	n, err := c.r.Read(b)
	c.n.Add(int64(n))

	return n, err
}
