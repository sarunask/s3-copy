package walker

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

func UseCSVFile(csvPath string, filesChan chan<- SrcDest, errors chan<- SrcDest, newerThan time.Time) {
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
		filePath, ext, err := prepareFilePath(filePath)
		if err != nil {
			errors <- SrcDest{
				SourceFile: filePath,
				DstObject:  rec[1],
				Error:      fmt.Errorf("file on path %s can't be found %w", rec[0], err),
			}
			continue
		}
		rec[1] = replaceWildcard(rec[1], ext)
		sum, size, err := getSizeAndSum(filePath)
		if err != nil {
			errors <- SrcDest{
				SourceFile: filePath,
				DstObject:  rec[1],
				Error:      err,
			}
			continue
		}
		if need2SkipOlderThan(filePath, newerThan) {
			// we need to skip this file, because it's older than we require
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

func prepareFilePath(filePath string) (string, string, error) {
	filename, ext := getFileNameAndExtension(filePath)
	var filePaths []string
	if ext == "*" {
		fExt := ""
		basePath := strings.Split(filePath, filename)[0]
		dir, err := os.ReadDir(basePath)
		if err != nil {
			return "", "", err
		}
		for _, dirEntry := range dir {
			fName, ext := getFileNameAndExtension(dirEntry.Name())
			if !dirEntry.IsDir() && strings.EqualFold(fName, filename) {
				if len(fExt) == 0 {
					fExt = ext
				}
				filePaths = append(filePaths, fmt.Sprintf("%s%s", basePath, dirEntry.Name()))
			}
		}
		if len(filePaths) > 1 {
			log.Infof("multiple files with the same name available by path '%s'", filePath)
		}
		if len(filePaths) > 0 { // Even in case of multiple files with the same name return the first one.
			return filePaths[0], fExt, nil
		}
		return "", "", fmt.Errorf("unable to find file with name %s", filename)
	}
	return filePath, "", nil
}

func getFileNameAndExtension(fileName string) (string, string) {
	if pPos := strings.LastIndexByte(fileName, '.'); pPos != -1 {
		if sPos := strings.LastIndexByte(fileName, '/'); sPos != -1 {
			return fileName[sPos+1 : pPos], fileName[pPos+1:]
		}
		return fileName[:pPos], fileName[pPos+1:]
	}
	return fileName, ""
}

func replaceWildcard(fileName, ext string) string {
	return strings.Replace(fileName, "*", ext, 1)
}

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
