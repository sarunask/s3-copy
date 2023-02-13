package env

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/araddon/dateparse"
	"github.com/sarunask/s3-copy/internal/logging"
	"github.com/spf13/pflag"
)

const shortTimeForm = time.RFC3339 // "2001-Jan-24 01:45"

// MaxWorkersCount notes how many workers we should have sending to S3
const MaxWorkersCount = 100

func (c *Config) validatePath() {
	var err error
	c.Path = filepath.ToSlash(filepath.Clean(c.Path))
	_, err = os.Stat(c.Path)
	if err != nil {
		log.Fatalf("[%s] Path '%s' is not valid dir or file.",
			logging.ERROR, c.Path)
	}
}

func (c *Config) validateKeyAndAlg() {
	if len(c.S3SSECKey) != 32 {
		log.Fatalf("[%s] S3 Customer Key must be 32 bytes long and not %d.",
			logging.ERROR, len(c.S3SSECKey))
	}
	if c.S3SSEC != "AES256" {
		log.Fatalf("[%s] S3 Customer algorithm must be AES256.",
			logging.ERROR)
	}
}

func (c *Config) validateWorkersCount() {
	if c.WorkersCount < 1 || c.WorkersCount > MaxWorkersCount {
		log.Fatalf("[%s] Workers should be in this range [1,100]",
			logging.ERROR)
	}
}

func (c *Config) validateExcludes() {
	for _, exclude := range *c.Exclude {
		_, err := regexp.Compile(exclude)
		if err != nil {
			log.Fatalf("[%s] Bad exclude regexp pattern '%s'",
				logging.ERROR, exclude)
		}
	}
}

func (c *Config) validateNewerThanAndAdd(newerThan *string) {
	if len(*newerThan) == 0 {
		// Don't check empty string
		return
	}
	timeNewerThan, err := dateparse.ParseAny(*newerThan)
	if err != nil {
		log.Fatalf("[%s] Correct format for newer-than is '%s': %v",
			logging.ERROR, shortTimeForm, err)
	}
	c.NewerThan = timeNewerThan
}

// Config is configuration which would be used in our project
type Config struct {
	S3Bucket     string
	S3Region     string
	S3SSEC       string
	S3SSECKey    string
	Exclude      *[]string
	Path         string
	Debug        bool
	WorkersCount int
	DryRun       bool
	DebugHTTP    bool
	NewerThan    time.Time
}

// Settings holds all settings we have in our app
var Settings *Config

func init() {
	sseC := pflag.String("sse-c", "AES256", "encryption type to be used in S3")
	sseCKey := pflag.String("sse-c-key", "", "encryption key to be used in S3")
	s3Region := pflag.String("s3-region", "eu-west-1", "S3 region")
	exclude := pflag.StringArray("exclude", nil, "which files to exclude (Regexp match)")
	s3bucket := pflag.String("s3-bucket", "", "S3 bucket where to upload")
	path := pflag.String("path", ".", "From which path to copy")
	debug := pflag.Bool("debug", false, "Enable debuging")
	debugHTTP := pflag.Bool("debug-http", false, "Enable debuging")
	workers := pflag.Int("workers", 5, "Number of workers")
	dryRun := pflag.Bool("dry-run", false, "Enable dry run - no upload")
	newerThan := pflag.String("newer-than", "", fmt.Sprintf("Include files with modification time Newer tha this time. Example time format is '%s'.",
		shortTimeForm))
	pflag.Parse()
	if len(*s3bucket) == 0 || len(*sseCKey) == 0 {
		fmt.Println("Not enough parameters")
		pflag.PrintDefaults()
		os.Exit(1)
	}
	Settings = &Config{
		S3Bucket:     *s3bucket,
		S3Region:     *s3Region,
		S3SSEC:       *sseC,
		S3SSECKey:    *sseCKey,
		Exclude:      exclude,
		Path:         *path,
		Debug:        *debug,
		DebugHTTP:    *debugHTTP,
		WorkersCount: *workers,
		DryRun:       *dryRun,
		NewerThan:    time.Time{},
	}
	Settings.validatePath()
	Settings.validateKeyAndAlg()
	Settings.validateWorkersCount()
	Settings.validateExcludes()
	Settings.validateNewerThanAndAdd(newerThan)
}
