package utils

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/PremiereGlobal/stim/pkg/stimlog"
)

const (
	UserOnlyMode  os.FileMode = 0700
	UserGroupMode os.FileMode = 0770
	WorldMode     os.FileMode = 0777
)

// CreateFileIfNotExist attempts to create the path and file if it does not exist
func CreateFileIfNotExist(filePath string, perm os.FileMode) error {
	log := stimlog.GetLogger()

	stat, err := os.Stat(filePath)
	if err == nil && !stat.IsDir() {
		return nil
	} else if err == nil && stat.IsDir() {
		return errors.New("given file path is a directory and not a path to a file")
	}

	// Check and create the base path if needed
	dir, _ := filepath.Split(filePath)
	if len(dir) > 0 {
		err := CreateDirIfNotExist(dir, perm)
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(filePath)
	if os.IsNotExist(err) {
		log.Debug("Creating file: '{}''", filePath)
		f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
		f.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateDirIfNotExist attempts to create the path if it does not exist
func CreateDirIfNotExist(path string, perm os.FileMode) error {
	log := stimlog.GetLogger()

	s, err := os.Stat(path)
	if err == nil && s.IsDir() {
		return nil
	} else if err == nil && !s.IsDir() {
		return errors.New("Path '" + path + "' is already a file!")
	}
	if !os.IsNotExist(err) {
		return err
	}
	log.Debug("Creating folder: '{}'", path)
	err = os.MkdirAll(path, perm)
	if err == nil {
		return nil
	}
	return err
}

// IsDirectory returns true if the given path is a directory, false otherwise
func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

// EnsureLink ensures that a link to the source exists at the target
// Both source and target should be full paths
func EnsureLink(source, target string) error {

	// Deal with the file already existing
	if data, err := os.Lstat(target); !os.IsNotExist(err) {
		if data.Mode()&os.ModeSymlink != 0 {
			linkPath, err := os.Readlink(target)
			if err != nil {
				return err
			}

			// Already pointing to the right place, nothing to do
			if source == linkPath {
				return nil
			}

			// Symlink is not pointing to the right place, remove it
			err = os.Remove(target)
			if err != nil {
				return err
			}

		} else {
			// File is not a symlink, remove it
			err := os.Remove(target)
			if err != nil {
				return err
			}
		}
	}

	return os.Symlink(source, target)
}
