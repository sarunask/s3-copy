package copy

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"

	"github.com/sarunask/s3-copy/internal/walker"
)

// Uploader provides class to upload files to S3
type Uploader struct {
	Client    s3manageriface.UploaderAPI
	S3Bucket  string
	S3SSEC    string
	S3SSECKey string
}

// AddFileToS3 will upload a single file to S3, it will require a pre-built aws session
// and will set file info like content type and encryption on the uploaded file.
func (u *Uploader) AddFileToS3(file walker.SrcDest) error {
	// Create an uploader with the session and default options
	info, err := os.Stat(file.SourceFile)
	if err != nil {
		return fmt.Errorf("could get stats for %v: %w", file.SourceFile, err)
	}
	if info.IsDir() {
		// No need to upload dirs
		log.Debugf("Skiping %s as it's directory", file.SourceFile)
		return nil
	}
	// It's not directory we upload, so read content
	f, err := os.Open(file.SourceFile)
	if err != nil {
		return fmt.Errorf("failed to open file %v: %w", file.SourceFile, err)
	}
	defer f.Close()

	input := &s3manager.UploadInput{
		Bucket: aws.String(u.S3Bucket),
		Key:    aws.String(filepath.ToSlash(file.DstObject)),
		Body:   f,
	}
	if len(u.S3SSECKey) != 0 {
		input.SSECustomerAlgorithm = aws.String(u.S3SSEC)
		input.SSECustomerKey = aws.String(u.S3SSECKey)
	}
	// Upload the file to S3.
	result, err := u.Client.Upload(input, func(u *s3manager.Uploader) {
		u.PartSize = 10 * 1024 * 1024 // 10MB part size
		u.LeavePartsOnError = true    // Don't delete the parts if the upload fails.
	})
	if err != nil {
		return fmt.Errorf("failed to upload file %v: %w", file.SourceFile, err)
	}
	log.Infof("successfuly uploaded %v to %v", file.SourceFile, result.Location)
	return nil
}
