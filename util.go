package main

import (
	"os"
	"path/filepath"

	. "github.com/xaxys/oasis/api"
)

func CheckFolder(args ...string) string {
	folder := filepath.Join(args...)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		getLogger().Infof("Creating directory %s", folder)
		if err := os.MkdirAll(folder, os.ModePerm); err != nil {
			getLogger().Warnf("Failed create directory %s", folder)
		}
	}
	return folder
}

func GetFile(args ...string) *os.File {
	path := filepath.Join(args...)
	folder, _ := filepath.Split(path)
	CheckFolder(folder)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			getLogger().Infof("Creating file %s", path)
			file, err = os.Create(path)
			if err != nil {
				getLogger().Warnf("Failed create file %s", path)
			}
		}
	}
	return file
}

func CheckFile(args ...string) {
	path := filepath.Join(args...)
	f := GetFile(path)
	f.Close()
}

func Compare(a string, b string, opt COMPARATOR) bool {
	switch opt {
	case GREATER:
		return a > b
	case GREATER_EQUAL:
		return a >= b
	case LESS:
		return a < b
	case LESS_EQUAL:
		return a <= b
	case EQUAL:
		return a == b
	case UNEQUAL:
		return a != b
	case ANY:
		return true
	default:
		return false
	}
}
