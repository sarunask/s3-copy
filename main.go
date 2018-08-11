package main

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sarunask/s3-copy/copy"
	"github.com/sarunask/s3-copy/walker"

	"github.com/sarunask/s3-copy/env"
	"github.com/sarunask/s3-copy/logging"

	"github.com/hashicorp/logutils"
)

var wg sync.WaitGroup

func uploadOne(up *copy.Uploader, filePath string) {
	defer wg.Done()
	if env.Settings.DryRun {
		return
	}
	err := up.AddFileToS3(filePath)
	if err != nil {
		log.Printf("[%s] Error uploading '%s': %v",
			logging.ERROR, filePath, err)
	}
}

func uploadAll(filesList chan string, exit chan int) {
	defer close(exit)
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
	uploader := copy.Uploader{
		Client:    s3manager.NewUploader(sess),
		S3Bucket:  env.Settings.S3Bucket,
		S3SSEC:    env.Settings.S3SSEC,
		S3SSECKey: env.Settings.S3SSECKey,
	}

	goRoutinesCount := 0
	for filePath := range filesList {
		wg.Add(1)
		goRoutinesCount++
		log.Printf("[%s] %d Starting upload of '%s'",
			logging.DEBUG, goRoutinesCount, filePath)
		go uploadOne(&uploader, filePath)
		if goRoutinesCount%env.Settings.WorkersCount == 0 {
			goRoutinesCount = 0
			log.Printf("[%s] Waiting for some goroutines to stop", logging.WARN)
			wg.Wait()
		}
	}
	wg.Wait()
}

func main() {
	//Setting log filters to filter messages
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{logging.DEBUG, logging.WARN, logging.ERROR},
		MinLevel: logutils.LogLevel(logging.WARN),
		Writer:   os.Stderr,
	}
	if env.Settings.Debug {
		filter.MinLevel = logging.DEBUG
	}
	log.SetOutput(filter)

	fileList := make(chan string)
	exit := make(chan int)
	go walker.Walk(env.Settings.Path, fileList, env.Settings.Exclude)
	go uploadAll(fileList, exit)
	<-exit
}
