package walker

import (
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNeed2Skip(t *testing.T) {
	t.Parallel()

	cases := []struct {
		Excludes *[]string
		Path     string
		Skip     bool
	}{
		{
			Excludes: &[]string{
				".*.go",
			},
			Path: "/home/test/Some/Test/test.go",
			Skip: true,
		},
		{
			Excludes: &[]string{
				".*/Some/.*",
			},
			Path: "/home/test/Some/Test/test.go",
			Skip: true,
		},
		{
			Excludes: &[]string{
				"NotFound",
			},
			Path: "/home/test/Some/Test/test.go",
			Skip: false,
		},
	}

	for i, c := range cases {
		fmt.Printf("Matching '%s' against '%#v'\n", c.Path, *c.Excludes)
		skip := need2skip(c.Path, c.Excludes)
		assert.Equal(t, skip, c.Skip,
			fmt.Sprintf("they should be equal in iteration %d", i))
	}
}

func TestGetSizeAndSum(t *testing.T) {
	t.Parallel()

	// prepare test file
	tF, err := os.CreateTemp("", "tmpfile-")
	assert.NoError(t, err)
	defer func() {
		_ = os.Remove(tF.Name())
	}()
	for i := 0; i < 100*1024; i++ {
		// nolint
		_, err = tF.Write([]byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Urna condimentum mattis pellentesque id. Et odio pellentesque diam volutpat commodo sed egestas. In dictum non consectetur a erat nam at lectus urna. Sapien et ligula ullamcorper malesuada proin. Eget mi proin sed libero enim sed. Nunc lobortis mattis aliquam faucibus purus in massa tempor. Nisl vel pretium lectus quam id leo. Amet mauris commodo quis imperdiet massa tincidunt nunc pulvinar. Amet consectetur adipiscing elit ut aliquam. Id semper risus in hendrerit gravida rutrum quisque. Nibh cras pulvinar mattis nunc sed blandit. Justo donec enim diam vulputate ut. Malesuada bibendum arcu vitae elementum curabitur vitae nunc. Nec dui nunc mattis enim ut tellus elementum sagittis vitae. Vitae tortor condimentum lacinia quis vel. Posuere sollicitudin aliquam ultrices sagittis orci a scelerisque. Tellus orci ac auctor augue. Mattis rhoncus urna neque viverra justo nec ultrices dui sapien."))
		assert.NoError(t, err)
	}
	tF.Close()
	sum, size, err := getSizeAndSum(tF.Name())
	assert.NoError(t, err)
	assert.Equal(t, sum, "6c62adc96b28bb8a141ca009f74ad345226c265806fe0eeecadcb524769f88c5")
	assert.Equal(t, size, uint64(0x63e7000))
}

func TestUseCSVFile(t *testing.T) {
	t.Parallel()
	path, err := os.Getwd()
	assert.NoError(t, err)
	testCases := []struct {
		name          string
		fileName      string
		dirName       string
		data          []byte
		fileDetails   SrcDest
		resultDetails SrcDest
	}{
		{
			name:     "Positive Case",
			fileName: "testFile1.txt",
			dirName:  "testFiles1",
			data:     []byte("./testFiles1/test1.txt,/customers/gu/upload/test1.txt"),
			fileDetails: SrcDest{
				SourceFile:   fmt.Sprintf("%s/testFiles1/test1.txt", path),
				SourceSha256: "d56ddee7d0fe47470cc19775dbe3ebc01b80bfee1f917b7fe3796b5ce7fb3d16",
				SourceSize:   0x12,
				DstObject:    "/customers/gu/upload/test1.txt",
				Error:        nil,
			},
			resultDetails: SrcDest{},
		},
		{
			name:     "Negative Case",
			fileName: "testFile2.txt",
			dirName:  "testFiles2",
			data:     []byte("./test1.txt,/customers/gu/upload/test1.txt"),
			resultDetails: SrcDest{
				SourceFile:   fmt.Sprintf("%s/test1.txt", path),
				SourceSha256: "",
				SourceSize:   0,
				DstObject:    "/customers/gu/upload/test1.txt",
				Error:        fmt.Errorf("no such file or directory"),
			},
		},
		{
			name:     "Wildcard Case",
			fileName: "testFile3.txt",
			dirName:  "testFiles3",
			data:     []byte("./testFiles3/test1.*,/customers/gu/upload/test1.*"),
			fileDetails: SrcDest{
				SourceFile:   fmt.Sprintf("%s/testFiles3/test1.tsv", path),
				SourceSha256: "d56ddee7d0fe47470cc19775dbe3ebc01b80bfee1f917b7fe3796b5ce7fb3d16",
				SourceSize:   0x12,
				DstObject:    "/customers/gu/upload/test1.tsv",
				Error:        nil,
			},
			resultDetails: SrcDest{},
		},
	}

	for _, tc := range testCases {
		tc := tc
		fileList := make(chan SrcDest)
		results := make(chan SrcDest)
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tearDown := setupTest(t, tc.dirName, tc.fileName, tc.data)
			defer tearDown(t)

			go UseCSVFile(getPath(tc.dirName, tc.fileName), fileList, results)
			for {
				select {
				case r := <-results:
					assert.Equal(t, tc.resultDetails.SourceFile, r.SourceFile)
					if tc.resultDetails.Error != nil {
						assert.ErrorContains(t, r.Error, tc.resultDetails.Error.Error())
					} else {
						assert.NoError(t, r.Error)
					}
					return
				case f := <-fileList:
					assert.Equal(t, tc.fileDetails, f, "Unexpected files channel message")
					return
				}
			}
		})
	}
}

func Test_fileExtension(t *testing.T) {

	tests := []struct {
		name         string
		fileName     string
		wantExt      string
		wantFileName string
	}{
		{
			name:         "Valid file extension",
			fileName:     "some/test/path/test.me",
			wantExt:      "me",
			wantFileName: "test",
		},
		{
			name:         "Wildcard file extension",
			fileName:     "test.*",
			wantExt:      "*",
			wantFileName: "test",
		},
		{
			name:         "Without extension",
			fileName:     "test",
			wantExt:      "",
			wantFileName: "test",
		},
		{
			name:         "Without extension 2",
			fileName:     "test.",
			wantExt:      "",
			wantFileName: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileName, ext := getFileNameAndExtension(tt.fileName)
			assert.Equalf(t, tt.wantExt, ext, "getFileNameAndExtension(%v)", tt.fileName)
			assert.Equalf(t, tt.wantFileName, fileName, "getFileNameAndExtension(%v)", tt.fileName)
		})
	}
}

func Test_prepareFilePath(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		want     string
		wantExt  string
		wantErr  error
	}{
		{
			name:     "Regular path",
			filePath: "./some/path/to/file.txt",
			want:     "./some/path/to/file.txt",
			wantExt:  "",
			wantErr:  nil,
		},
		{
			name:     "Wildcard path",
			filePath: "./testFiles/test1.*",
			want:     "./testFiles/test1.tsv",
			wantExt:  "tsv",
			wantErr:  nil,
		},
		{
			name:     "Wrong path",
			filePath: "./wrong/test1.*",
			want:     "",
			wantExt:  "",
			wantErr:  fmt.Errorf("no such file or directory"),
		},
		{
			name:     "Wrong file name",
			filePath: "./testFiles/test.*",
			want:     "",
			wantExt:  "",
			wantErr:  fmt.Errorf("unable to find file"),
		},
	}
	tearDown := setupTest(t, "testFiles", "someFile", []byte("./testFiles3/test1.*,/customers/gu/upload/test1.txt"))
	defer tearDown(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotExt, err := prepareFilePath(tt.filePath)
			if tt.wantErr != nil {
				assert.ErrorContains(t, err, tt.wantErr.Error(), fmt.Sprintf("prepareFilePath(%v)", tt.filePath))
			}
			assert.Equalf(t, tt.want, got, "prepareFilePath(%v)", tt.filePath)
			assert.Equalf(t, tt.wantExt, gotExt, "prepareFilePath(%v)", tt.filePath)
		})
	}
}

func setupTest(t *testing.T, dirName, fileName string, data []byte) func(t *testing.T) {
	tsvFile := "test1.tsv"
	txtFile := "test1.txt"
	testData := []byte("Some text for test")

	assert.NoError(t, os.Mkdir(dirName, 0700))
	assert.NoError(t, os.WriteFile(getPath(dirName, fileName), data, fs.ModePerm))
	assert.NoError(t, os.WriteFile(getPath(dirName, tsvFile), testData, fs.ModePerm))
	assert.NoError(t, os.WriteFile(getPath(dirName, txtFile), testData, fs.ModePerm))

	return func(t *testing.T) {
		assert.NoError(t, os.RemoveAll(fmt.Sprintf("./%s", dirName)))
	}
}

func getPath(dirName, fileName string) string {
	return fmt.Sprintf("./%s/%s", dirName, fileName)
}
