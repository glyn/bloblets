package appbits

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/cli/cf/models"
)

func UploadApp(appGuid string, zipFile *os.File, presentFiles []models.AppFileFields) {
	fmt.Println("appbits.UploadApp not implemented")
}
