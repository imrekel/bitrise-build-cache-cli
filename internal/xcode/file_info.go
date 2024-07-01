package xcode

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/bitrise-io/go-utils/v2/log"
)

type FileInfo struct {
	Path    string    `json:"path"`
	Hash    string    `json:"hash"`
	ModTime time.Time `json:"modTime"`
}

func processFileInfoFunc(rootDir string, logger log.Logger, fileInfos *[]FileInfo) func(string, fs.DirEntry, error) error {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		inf, err := d.Info()
		if err != nil {
			return fmt.Errorf("get file info: %w", err)
		}

		// Skip symbolic links
		if inf.Mode()&os.ModeSymlink != 0 {
			logger.Debugf("Skipping symbolic link: %s", path)

			return nil
		}

		hashString, err := checksumOfFile(path)
		if err != nil {
			logger.Debugf("Error calculating hash: %v", err)

			return nil
		}

		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}

		// Create FileInfo object
		fileInfo := FileInfo{
			Path:    relPath,
			Hash:    hashString,
			ModTime: inf.ModTime(),
		}

		// Append FileInfo to slice
		*fileInfos = append(*fileInfos, fileInfo)

		return nil
	}
}

func calculateFileInfos(rootDir string, logger log.Logger) ([]FileInfo, error) {
	if rootDir == "" {
		return nil, fmt.Errorf("missing rootDir")
	}

	var fileInfos []FileInfo

	// Walk through the directory tree
	err := filepath.WalkDir(rootDir, processFileInfoFunc(rootDir, logger, &fileInfos))
	if err != nil {
		return nil, fmt.Errorf("calculate file infos: %w", err)
	}

	logger.Infof("(i) Processed %d files", len(fileInfos))

	return fileInfos, nil
}
