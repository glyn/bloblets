package resmatch

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/glyn/bloblets/servutil"
)

func ResourceMatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		servutil.Fail(w, "invalid method: %s", r.Method)
		return
	}

	requestBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		servutil.Fail(w, "read body failed: %s", err)
		return
	}

	requestFieldsColl := []IntegrityFields{}
	err = json.Unmarshal(requestBytes, &requestFieldsColl)
	if err != nil {
		servutil.Fail(w, "unmarshal body failed: %s", err)
		return
	}

	responseFields, err := integrityFieldsJSON()
	if err != nil {
		servutil.Fail(w, "cannot compute response: %s", err)
		return
	}

	_, err = w.Write(responseFields)
	if err != nil {
		servutil.Fail(w, "cannot write response: %s", err)
		return
	}
}
