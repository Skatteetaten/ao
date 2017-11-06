package command

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func DefaultTablePrinter(lines []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
	w.Flush()
}
