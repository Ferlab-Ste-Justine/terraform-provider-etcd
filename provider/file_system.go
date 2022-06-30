package provider

import (
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"os"
)

func EnsureDirectoryExists(path string, dirPermission int32) error {
	_, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.Mkdir(path, os.FileMode(dirPermission)); err != nil {
			return err
		}
	}

	return nil
}

func GetDirectoryContent(path string) (map[string]KeyInfo, error) {
	keys := make(map[string]KeyInfo)

	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			keys[path] = KeyInfo{
				Key: path,
				Value: string(content),
				Version: 0,
				CreateRevision: 0,
				ModRevision: 0,
				Lease: 0,
			}
		}

		return nil
	})

	return keys, err
}

func ApplyDiffToDirectory(path string, diff KeysDiff, filesPermission int32, dirPermission int32) error {
	for _, file := range diff.Deletions {
		fPath := filepath.Join(path, file)
		err := os.Remove(fPath)
		if err != nil {
			return err
		}
	}

	for file, content := range diff.Upserts {
		fPath := filepath.Join(path, file)
		fdir := filepath.Dir(fPath)
		mkdirErr := os.MkdirAll(fdir, os.FileMode(dirPermission))
		if mkdirErr != nil {
			return mkdirErr
		}

		f, err := os.OpenFile(fPath, os.O_RDWR|os.O_CREATE, os.FileMode(filesPermission))
		if err != nil {
			return err
		}

		err = f.Truncate(0)
		if err != nil {
			f.Close()
			return err
		}

		_, err = f.Write([]byte(content))
		if err != nil {
			f.Close()
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}
	}

	return nil
}