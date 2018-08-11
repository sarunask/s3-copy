package walker

import (
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sarunask/s3-copy/logging"
)

func need2skip(pathToCheck string, excludes *[]string) bool {
	//Skip all files, which have such words in them
	for _, exclude := range *excludes {
		matched, err := regexp.MatchString(exclude, pathToCheck)
		if err != nil {
			log.Fatalf("[%s] Bad exclude regexp pattern '%s'",
				logging.ERROR, exclude)
			continue
		}
		if matched {
			log.Printf("[%s] Skipping %s", logging.DEBUG, pathToCheck)
			return true
		}
	}
	return false
}

//Walk would recursivly get all files (except but excluded)
//And would write files path to fileChan channel
func Walk(walkPath string, filesChan chan string, excludes *[]string) {
	defer close(filesChan)
	err := filepath.Walk(walkPath, func(path string, f os.FileInfo, err error) error {
		//Only append files which are not dirs and we don't need 2 skip that file
		if f != nil && f.IsDir() == false && need2skip(path, excludes) == false {
			log.Printf("[%s] Adding %s to be copied", logging.DEBUG, path)
			filesChan <- path
		}
		return nil
	})
	if err != nil {
		log.Printf("[%s] Error: %v\n", logging.ERROR, err)
	}
}
