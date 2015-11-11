package appbits

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cloudfoundry/cli/cf/api/resources"
	"github.com/glyn/bloblets/servutil"
)

func AppHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(100 * 1024 * 1024)
	mpForm := r.MultipartForm

	res := mpForm.Value["resources"]
	presentFiles := []resources.AppFileResource{}
	err := json.Unmarshal([]byte(res[0]), &presentFiles)
	if err != nil {
		servutil.Fail(w, "demarshalling resources failed: %s", err)
		return
	}

	application := mpForm.File["application"]

	fmt.Printf("appbits.AppHandler not implemented:\npresentFiles=%#v\napplication=%#v\n", presentFiles, application)
}
