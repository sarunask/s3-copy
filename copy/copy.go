package copy

import (
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"

	"github.com/sarunask/s3-copy/logging"
)

//Uploader provides class to upload files to S3
type Uploader struct {
	Client    s3manageriface.UploaderAPI
	S3Bucket  string
	S3SSEC    string
	S3SSECKey string
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func (u *Uploader) AddFileToS3(filename string) error {
	// Create an uploader with the session and default options
	info, err := os.Stat(filename)
	if err != nil {
		return errors.Wrapf(err, "could get stats for %v", filename)
	}
	if info.IsDir() {
		//No need to upload dirs
		log.Printf("[%s] Skiping %s as it's directory", logging.DEBUG, filename)
		return nil
	}
	//It's not directory we upload, so read content
	f, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to open file %v", filename)
	}

	// Upload the file to S3.
	result, err := u.Client.Upload(&s3manager.UploadInput{
		Bucket:               aws.String(u.S3Bucket),
		Key:                  aws.String(filepath.ToSlash(filename)),
		SSECustomerAlgorithm: aws.String(u.S3SSEC),
		SSECustomerKey:       aws.String(u.S3SSECKey),
		Body:                 f,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to upload file, %v", filename)
	}
	log.Printf("[%s] successfuly upload %v to %v", logging.DEBUG, filename, result.Location)
	return nil
}
