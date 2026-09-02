package jsengine

import (
	"io"
)

type console struct {
	w io.Writer
}

func (c *console) Log(s string) {
	_, _ = c.w.Write([]byte(s + "\n"))
}
