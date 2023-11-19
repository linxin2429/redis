package logger

import (
	"fmt"
	"os"
	"path"
)

func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

func mkDir(src string) error {
	return os.MkdirAll(src, os.ModePerm)
}

func isNotExistMkDir(src string) error {
	if checkNotExist(src) {
		return mkDir(src)
	}
	return nil
}

func mustOpen(filename, dir string) (*os.File, error) {
	if checkPermission(dir) {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}

	if err := isNotExistMkDir(dir); err != nil {
		return nil, fmt.Errorf("mkdir dir: %v", err)
	}

	f, err := os.OpenFile(path.Join(dir, filename), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("fail to open file: %v", err)
	}

	return f, nil
}
