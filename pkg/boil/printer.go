package boil

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
)

// Printer is a boil preset TabWriter.
type Printer struct {
	w *tabwriter.Writer
}

// NewPrinter returns a new *Printer that writes to w.
// If w is nil Printer prints to stdout.
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

func (self *Printer) Write(p []byte) (n int, err error) { 
	if _, err = self.w.Write(p) ; err != nil {
		return
	}
	return 0, self.w.Flush()
}