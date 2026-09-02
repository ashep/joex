package executor

type logger struct {
	off int
	buf []byte
}

func newLogger(bufSize int) *logger {
	if bufSize <= 0 {
		panic("buffer size must be greater than zero")
	}

	return &logger{
		buf: make([]byte, bufSize),
	}
}

func (l *logger) Write(msg []byte) (int, error) {
	// Overwrite entire buf
	if len(msg) >= len(l.buf) {
		return copy(l.buf, msg[len(msg)-len(l.buf):]), nil
	}

	// Shift left
	if needSpace := len(l.buf) - l.off - len(msg); needSpace < 0 {
		copy(l.buf, l.buf[-needSpace:])
		l.off = len(l.buf) - len(msg)
	}

	n := copy(l.buf[l.off:], msg)
	l.off += n

	return n, nil
}

func (l *logger) String() string {
	return string(l.buf[:l.off])
}

func (l *logger) Close() error {
	return nil
}
