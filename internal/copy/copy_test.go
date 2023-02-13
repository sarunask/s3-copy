package copy

import (
	"fmt"
	"path"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

type mockS3Manager struct {
	s3manageriface.UploaderAPI
	Resp s3manager.UploadOutput
}

func (m mockS3Manager) Upload(inp *s3manager.UploadInput, _ ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	// fmt.Printf("%#v", *inp)
	m.Resp.Location = fmt.Sprintf("https://%s/%s", *inp.Bucket, path.Clean(*inp.Key))
	// mock response/functionality
	return &m.Resp, nil
}

func TestAddFileToS3(t *testing.T) {
	cases := []struct {
		Resp s3manager.UploadOutput
	}{
		{
			Resp: s3manager.UploadOutput{
				Location: "https://mockS3Bucket_0/copy.go",
			},
		},
	}

	for i, c := range cases {
		u := Uploader{
			Client:    mockS3Manager{Resp: c.Resp},
			S3Bucket:  fmt.Sprintf("mockS3Bucket_%d", i),
			S3SSEC:    "AES256",
			S3SSECKey: fmt.Sprintf("czn8qrbUsT/5y5Hr2i93ImWmIQLCZ1%0d", i),
		}
		err := u.AddFileToS3("./copy.go")
		if err != nil {
			t.Fatalf("%d, unexpected error", err)
		}
	}
}
