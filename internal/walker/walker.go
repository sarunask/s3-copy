package walker

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/sarunask/s3-copy/internal/logging"
)

func need2skip(pathToCheck string, excludes *[]string) bool {
	// Skip all files, which have such words in them
	for _, exclude := range *excludes {
		matched, err := regexp.MatchString(exclude, pathToCheck)
		if err != nil {
			log.Fatalf("[%s] Bad exclude regexp pattern '%s'",
				logging.ERROR, exclude)
			continue
		}
		if matched {
			log.Printf("[%s] Skipping %s", logging.WARN, pathToCheck)
			return true
		}
	}
	return false
}

func need2SkipOlderThan(pathToCheck string, newerThan time.Time) bool {
	stat, err := os.Stat(pathToCheck)
	if err != nil {
		// If we can't get Stat on file - skip it
		log.Printf("[%s] Can't get stat for %s", logging.ERROR, pathToCheck)
		return true
	}
	before := stat.ModTime().Before(newerThan)
	if before {
		log.Printf("[%s] Skipping file as it's mod time %v is before %v", logging.WARN,
			stat.ModTime(), newerThan)
	}
	return before
}

// Walk would recursivly get all files (except but excluded)
// And would write files path to fileChan channel
func Walk(walkPath string, filesChan chan string, excludes *[]string, newerThan time.Time) {
	defer close(filesChan)
	err := filepath.Walk(walkPath, func(path string, f os.FileInfo, err error) error {
		// Only append files which are not dirs and we don't need 2 skip that file
		if f != nil && !f.IsDir() && !need2skip(path, excludes) && !need2SkipOlderThan(path, newerThan) {
			log.Printf("[%s] Adding %s to be copied", logging.DEBUG, path)
			filesChan <- path
		}
		return nil
	})
	if err != nil {
		log.Printf("[%s] Error: %v\n", logging.ERROR, err)
	}
}
