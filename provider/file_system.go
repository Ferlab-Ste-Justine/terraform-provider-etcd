package provider

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
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

func GetDirectoryContent(path string) (map[string]client.KeyInfo, error) {
	keys := make(map[string]client.KeyInfo)

	err := filepath.WalkDir(path, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			keys[path] = client.KeyInfo{
				Key:            path,
				Value:          string(content),
				Version:        0,
				CreateRevision: 0,
				ModRevision:    0,
				Lease:          0,
			}
		}

		return nil
	})

	return keys, err
}

func applyFileToDirectory(path string, file string, content string, filesPermission int32, dirPermission int32) error {
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

	return nil
}

func ApplyDiffToDirectory(path string, diff client.KeyDiff, filesPermission int32, dirPermission int32) error {
	for _, file := range diff.Deletions {
		fPath := filepath.Join(path, file)
		err := os.Remove(fPath)
		if err != nil {
			return err
		}
	}

	for file, content := range diff.Inserts {
		applyErr := applyFileToDirectory(path, file, content, filesPermission, dirPermission)
		if applyErr != nil {
			return applyErr
		}
	}

	for file, content := range diff.Updates {
		applyErr := applyFileToDirectory(path, file, content, filesPermission, dirPermission)
		if applyErr != nil {
			return applyErr
		}
	}

	return nil
}
