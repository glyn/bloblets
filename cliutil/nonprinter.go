package cliutil

type NonPrinter struct {
}

func (t *NonPrinter) Print(values ...interface{}) (n int, err error) {
	return 0, nil
}

func (t *NonPrinter) Printf(format string, values ...interface{}) (n int, err error) {
	return 0, nil
}

func (t *NonPrinter) Println(values ...interface{}) (n int, err error) {
	return 0, nil
}

func (t *NonPrinter) ForcePrint(values ...interface{}) (n int, err error) {
	return 0, nil
}

func (t *NonPrinter) ForcePrintf(format string, values ...interface{}) (n int, err error) {
	return 0, nil
}

func (t *NonPrinter) ForcePrintln(values ...interface{}) (n int, err error) {
	return 0, nil
}
