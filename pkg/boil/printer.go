package boil

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

type Printer struct {
	w *tabwriter.Writer
}

func NewPrinter(w io.Writer) *Printer {
	if w == nil {
		w = os.Stdout
	}
	return &Printer{
		w: tabwriter.NewWriter(w, 2, 2, 2, 32, 0),
	}
}

func (self *Printer) Printf(format string, args ...any) {
	fmt.Fprintf(self.w, format, args...)
	self.w.Flush()
}

func (self *Printer) Write(p []byte) (n int, err error) { return self.w.Write(p) }

func (self *Printer) Flush() { self.w.Flush() }
