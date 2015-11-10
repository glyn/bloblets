package resmatch

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/glyn/bloblets/servutil"
)

func (m matcher) ResourceMatchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		servutil.Fail(w, "invalid method: %s", r.Method)
		return
	}

	requestBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		servutil.Fail(w, "reading body failed: %s", err)
		return
	}

	requestFields := []IntegrityFields{}
	err = json.Unmarshal(requestBytes, &requestFields)
	if err != nil {
		servutil.Fail(w, "unmarshalling body failed: %s", err)
		return
	}

	responseFields, err := m.matchResources(requestFields)
	if err != nil {
		servutil.Fail(w, "computing response failed: %s", err)
		return
	}

	responseFieldsBytes, err := json.Marshal(responseFields)
	if err != nil {
		servutil.Fail(w, "marshalling response failed: %s", err)
		return
	}

	_, err = w.Write(responseFieldsBytes)
	if err != nil {
		servutil.Fail(w, "writing response failed: %s", err)
		return
	}
}
