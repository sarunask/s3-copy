package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sarunask/s3-copy/internal/copy"
	"github.com/sarunask/s3-copy/internal/walker"

	"github.com/sarunask/s3-copy/internal/env"
)

func uploadOne(
	file walker.SrcDest,
	results chan<- walker.SrcDest,
	wg *sync.WaitGroup,
) {
	defer wg.Done()
	if env.Settings.DryRun {
		return
	}
	var logLevel aws.LogLevelType
	if env.Settings.DebugHTTP {
		logLevel = aws.LogDebugWithHTTPBody
	}
	// Create a single AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region:   aws.String(env.Settings.S3Region),
		LogLevel: &logLevel,
	}))

	// Create an uploader with the session and default options
	up := copy.Uploader{
		Client:    s3manager.NewUploader(sess),
		S3Bucket:  env.Settings.S3Bucket,
		S3SSEC:    env.Settings.S3SSEC,
		S3SSECKey: env.Settings.S3SSECKey,
	}

	// actually copy files to s3
	err := up.AddFileToS3(file)
	if err != nil {
		// add error to results
		file.Error = fmt.Errorf("error uploading %s: %w",
			file.SourceFile, err)
	}
	// we send to results channel file with error or without as success
	results <- file
}

func uploadAll(
	filesList chan walker.SrcDest,
	results chan walker.SrcDest,
) {
	defer close(results)

	var wg sync.WaitGroup
	goRoutinesCount := 0
	for filePath := range filesList {
		wg.Add(1)
		goRoutinesCount++
		log.Debugf("%d starting upload of '%#v'",
			goRoutinesCount, filePath)
		// add go routine to upload file
		go uploadOne(filePath, results, &wg)
		if goRoutinesCount%env.Settings.WorkersCount == 0 {
			goRoutinesCount = 0
			log.Warnf("waiting for some goroutines to stop")
			wg.Wait()
		}
	}
	wg.Wait()
}

// closeFile will close file or report error
func closeFile(f *os.File) {
	err := f.Close()
	if err != nil {
		log.Debugf("error closing %s: %v", f.Name(), err)
	}
}

// openFile will open file for fName and will return file handler or exit with fatal error
func openFile(fName string) *os.File {
	f, err := os.OpenFile(fName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("can't write to output %s: %v", fName, err)
	}
	return f
}

// writeOutput will write output CSV files with results of file upload
func writeOutput(
	results chan walker.SrcDest,
	exit chan struct{},
) {
	// create 2 files to output results
	success := openFile(env.Settings.OutputSuccessFile)
	defer closeFile(success)
	failure := openFile(env.Settings.OutputFailureFile)
	defer closeFile(failure)
	// wait for new record to add or for exit
	for res := range results {
		out := success
		if res.Error != nil {
			out = failure
		}
		log.Debugf("writing {%s} to %s",
			fmt.Sprintf("%s,%s,%s,%d,%v\n", res.SourceFile, res.DstObject, res.SourceSha256, res.SourceSize, res.Error),
			out.Name())
		_, err := out.WriteString(fmt.Sprintf("%s,%s,%s,%d,%v\n", res.SourceFile, res.DstObject, res.SourceSha256, res.SourceSize, res.Error))
		if err != nil {
			log.Errorf("can't write to %s: %v", out.Name(), err)
		}
	}
	// we exit only after we write all our files
	close(exit)
}

func main() {
	// Use more CPU's when available
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Setting log filters to filter messages
	// Initialize logger
	log.SetOutput(os.Stdout)
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.InfoLevel)

	if env.Settings.Debug {
		log.SetLevel(log.DebugLevel)
	}

	fileList := make(chan walker.SrcDest)
	results := make(chan walker.SrcDest)
	// exit is closed by last go routine when it's finished
	exit := make(chan struct{})
	if len(strings.Trim(env.Settings.InputCSVFile, "\n\r\t ")) != 0 {
		go walker.UseCSVFile(env.Settings.InputCSVFile, fileList, results, env.Settings.NewerThan)
	} else {
		go walker.Walk(env.Settings.Path, fileList, results, env.Settings.Exclude, env.Settings.NewerThan)
	}
	go uploadAll(fileList, results)
	go writeOutput(results, exit)
	<-exit
	log.Debugf("done - exiting")
}
