package downloader

type fixedBuffer struct {
	data []byte
	pos  int
}

type tooBigError struct{}

func (tooBigError) Error() string {
	return "too big"
}

func (buf *fixedBuffer) Write(p []byte) (n int, err error) {
	plen := len(p)
	if len(buf.data)-buf.pos < plen {
		return 0, tooBigError{}
	}
	copy(buf.data[buf.pos:], p)
	buf.pos += plen
	return plen, nil
}

func (buf fixedBuffer) bytes() []byte {
	return buf.data[:buf.pos]
}

func (buf *fixedBuffer) reset() {
	buf.pos = 0
}

func makeFixedBuffer(size int) fixedBuffer {
	return fixedBuffer{
		data: make([]byte, size),
		pos:  0,
	}
}
