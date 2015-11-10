package main

import (
	"net/http"
	"os"

	"github.com/cloudfoundry/cli/cf/models"
	"github.com/glyn/bloblets/cliutil"
	"github.com/glyn/bloblets/resmatch"
	"github.com/glyn/bloblets/scanner"
)

func main() {
	client := &http.Client{}

	scanner.Scan(os.Args[1], func(affs []models.AppFileFields) {
		request := resmatch.ResMatchRequest(integrityFields(affs))

		response := cliutil.Converse(client, request)

		resmatch.ProcessResponse(response)
	})

}

func integrityFields(affs []models.AppFileFields) []resmatch.IntegrityFields {
	ifs := make([]resmatch.IntegrityFields, 0, len(affs))

	for _, aff := range affs {
		ifs = append(ifs, resmatch.IntegrityFields{
			Sha1: aff.Sha1,
			Size: aff.Size,
		})
	}

	return ifs
}
