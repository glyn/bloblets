package cliutil

import (
	"time"

	"github.com/cloudfoundry/cli/cf/terminal"
)

type LoggingTeePrinter struct {
	terminal.TeePrinter
}

func (t *LoggingTeePrinter) Print(values ...interface{}) (n int, err error) {
	now := time.Now()
	return t.TeePrinter.Print(append([]interface{}{now}, values...)...)
}

func (t *LoggingTeePrinter) Printf(format string, values ...interface{}) (n int, err error) {
	now := time.Now()
	return t.TeePrinter.Printf("%s: "+format, append([]interface{}{now}, values...)...)
}

func (t *LoggingTeePrinter) Println(values ...interface{}) (n int, err error) {
	now := time.Now()
	return t.TeePrinter.Println(append([]interface{}{now}, values...)...)
}

func (t *LoggingTeePrinter) ForcePrint(values ...interface{}) (n int, err error) {
	now := time.Now()
	return t.TeePrinter.ForcePrint(append([]interface{}{now}, values...)...)
}

func (t *LoggingTeePrinter) ForcePrintf(format string, values ...interface{}) (n int, err error) {
	now := time.Now()
	return t.TeePrinter.ForcePrintf("%s: "+format, append([]interface{}{now}, values...)...)
}

func (t *LoggingTeePrinter) ForcePrintln(values ...interface{}) (n int, err error) {
	now := time.Now()
	return t.TeePrinter.ForcePrintln(append([]interface{}{now}, values...)...)
}
