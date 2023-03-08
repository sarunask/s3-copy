package walker

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

type (
	SrcDest struct {
		SourceFile   string
		SourceSha256 string
		SourceSize   uint64
		DstObject    string
		Error        error
	}
)

func need2skip(pathToCheck string, excludes *[]string) bool {
	// Skip all files, which have such words in them
	for _, exclude := range *excludes {
		matched, err := regexp.MatchString(exclude, pathToCheck)
		if err != nil {
			log.Fatalf("bad exclude regexp pattern '%s': %v", exclude, err)
			continue
		}
		if matched {
			log.Debugf("Skipping %s", pathToCheck)
			return true
		}
	}
	return false
}

func need2SkipOlderThan(pathToCheck string, newerThan time.Time) bool {
	stat, err := os.Stat(pathToCheck)
	if err != nil {
		// If we can't get Stat on file - skip it
		log.Fatalf("can't get stat for %s: %v", pathToCheck, err)
		return true
	}
	before := stat.ModTime().Before(newerThan)
	if before {
		log.Debugf("skipping file as it's mod time %v is before %v",
			stat.ModTime(), newerThan)
	}
	return before
}

// Walk would recursivly get all files (except but excluded)
// And would write files path to fileChan channel
func Walk(walkPath string, filesChan chan<- SrcDest, errors chan<- SrcDest, excludes *[]string, newerThan time.Time) {
	defer close(filesChan)
	// nolint
	filepath.Walk(walkPath, func(path string, f os.FileInfo, err error) error {
		// Only append files which are not dirs and we don't need 2 skip that file
		if f != nil && !f.IsDir() && !need2skip(path, excludes) && !need2SkipOlderThan(path, newerThan) {
			log.Debugf("Adding %s to be copied", path)
			sum, size, err := getSizeAndSum(path)
			if err != nil {
				errors <- SrcDest{
					SourceFile: path,
					DstObject:  path,
					Error:      fmt.Errorf("file on path %s: %v", path, err),
				}
				return err
			}
			filesChan <- SrcDest{
				SourceFile:   path,
				SourceSha256: sum,
				SourceSize:   size,
				DstObject:    filepath.Base(path),
			}
		}
		return nil
	})
}

func UseCSVFile(csvPath string, filesChan chan<- SrcDest, errors chan<- SrcDest) {
	defer close(filesChan)
	f, err := os.Open(csvPath)
	if err != nil {
		// nolint
		log.Fatalf("error opening %s: %v", csvPath, err)
	}
	in := csv.NewReader(f)
	recs, err := in.ReadAll()
	if err != nil {
		log.Fatalf("error opening %s: %v", csvPath, err)
	}
	for i, rec := range recs {
		if len(rec[0]) == 0 && len(rec[1]) == 0 {
			log.Debugf("found empty record on line %v", i+1)
			continue
		}

		filePath, err := filepath.Abs(rec[0])
		if err != nil {
			errors <- SrcDest{
				SourceFile: filePath,
				DstObject:  rec[1],
				Error:      fmt.Errorf("file on path %s can't be absolutized %w", rec[0], err),
			}
			continue
		}
		sum, size, err := getSizeAndSum(filePath)
		if err != nil {
			errors <- SrcDest{
				SourceFile: filePath,
				DstObject:  rec[1],
				Error:      err,
			}
			continue
		}
		filesChan <- SrcDest{
			SourceFile:   filePath,
			SourceSha256: sum,
			SourceSize:   size,
			DstObject:    rec[1],
		}
	}
}

func getSizeAndSum(filePath string) (string, uint64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("can't get infor for %s: %w", filePath, err)
	}
	f, err := os.Open(filePath)
	if err != nil {
		return "", 0, fmt.Errorf("can't open %s: %w", filePath, err)
	}
	buf := make([]byte, 1024*1024)
	h := sha256.New()
	if _, err := io.CopyBuffer(h, f, buf); err != nil {
		return "", 0, fmt.Errorf("can't calculate sum for %s: %w", filePath, err)
	}
	log.Debugf("%s size=%d sum256=%s", filePath, info.Size(), fmt.Sprintf("%x", h.Sum(nil)))
	return fmt.Sprintf("%x", h.Sum(nil)), uint64(info.Size()), nil
}
