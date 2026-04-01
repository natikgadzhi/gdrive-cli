package main

import (
	clioutput "github.com/natikgadzhi/cli-kit/output"
	"github.com/natikgadzhi/gdrive-cli/internal/table"
)

// borderedRenderer is implemented by types that can render themselves
// into a bordered table.
type borderedRenderer interface {
	RenderBorderedTable(t *table.Table)
}

// printResult writes data in the specified format. For JSON it marshals to
// stdout via cli-kit. For table mode it uses the bordered table package.
func printResult(format string, data any, renderer borderedRenderer) error {
	if clioutput.IsJSON(format) {
		return clioutput.PrintJSON(data)
	}
	t := table.New()
	renderer.RenderBorderedTable(t)
	return t.Flush()
}
