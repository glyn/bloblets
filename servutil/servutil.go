package servutil

import (
	"fmt"
	"net/http"
)

func Fail(w http.ResponseWriter, format string, a ...interface{}) {
	http.Error(w, fmt.Sprintf(format, a...), 500)
}
