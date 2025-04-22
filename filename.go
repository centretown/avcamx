package avcam

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	OutputBase = "./output/"
)

func NextFileName(parent, ext string) (string, error) {
	u := time.Now()
	foldername := u.Format("2006-01-02")

	path, err := filepath.Abs(filepath.Join(parent, foldername))
	if err != nil {
		return path, err
	}
	log.Println("directory", path)

	err = MakeFolder(path)
	if err != nil {
		return foldername, err
	}

	filename := u.Format("15:04:05-0700") + "-"
	filename = strings.ReplaceAll(filename, ":", "_")
	id := uuid.New()
	name := fmt.Sprintf("%s%s.%s", filename, id.String(), ext)
	path = filepath.Join(path, name)

	log.Println("filename", path)

	return path, err
}

func MakeFolder(folder string) (err error) {

	var info os.FileInfo
	info, err = os.Stat(folder)

	if err != nil && os.IsNotExist(err) {
		err = os.Mkdir(folder, os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
		info, err = os.Stat(folder)
		if err != nil {
			log.Println(err)
			return
		}
	}

	if !info.IsDir() {
		err = fmt.Errorf("not a folder")
	}
	return
}
