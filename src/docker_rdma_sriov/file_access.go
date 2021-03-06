package main

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type fileObject struct {
	Path string
	File *os.File
}

func (attrib *fileObject) Exists() bool {
	return fileExists(attrib.Path)
}

func (attrib *fileObject) Open() (err error) {
	attrib.File, err = os.OpenFile(attrib.Path, os.O_RDWR|syscall.O_NONBLOCK, 0660)
	return err
}

func (attrib *fileObject) OpenRO() (err error) {
	attrib.File, err = os.Open(attrib.Path)
	return err
}

func (attrib *fileObject) OpenWO() (err error) {
	attrib.File, err = os.OpenFile(attrib.Path, os.O_WRONLY, 0444)
	return err
}

func (attrib *fileObject) Close() (err error) {
	err = attrib.File.Close()
	attrib.File = nil
	return err
}

func (attrib *fileObject) Read() (str string, err error) {
	if attrib.File == nil {
		err = attrib.OpenRO()
		if err != nil {
			return
		}
		defer func() {
			e := attrib.Close()
			if err == nil {
				err = e
			}
		}()
	}
	attrib.File.Seek(0, os.SEEK_SET)
	data, err := ioutil.ReadAll(attrib.File)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (attrib *fileObject) Write(value string) (err error) {
	if attrib.File == nil {
		err = attrib.OpenWO()
		if err != nil {
			return
		}
		defer func() {
			e := attrib.Close()
			if err == nil {
				err = e
			}
		}()
	}
	attrib.File.Seek(0, os.SEEK_SET)
	_, err = attrib.File.WriteString(value)
	return err
}

func (attrib *fileObject) ReadInt() (value int, err error) {
	s, err := attrib.Read()
	if err != nil {
		return 0, err
	}
	s = strings.Trim(s, "\n")
	value, err = strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return value, err
}

func (attrib *fileObject) WriteInt(value int) (err error) {
	return attrib.Write(strconv.Itoa(value))
}

func lsFilesWithPrefix(dir string, filePrefix string, ignoreDir bool) ([]string, error) {
	var desiredFiles []string

	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fileInfos, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for i := range fileInfos {
		if ignoreDir && fileInfos[i].IsDir() {
			continue
		}

		if filePrefix == "" ||
			strings.Contains(fileInfos[i].Name(), filePrefix) {
			desiredFiles = append(desiredFiles, fileInfos[i].Name())
		}
	}
	return desiredFiles, nil
}

func lsDirs(dir string) ([]string, error) {
	var dirList []string

	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fileInfos, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	for i := range fileInfos {
		dirList = append(dirList, fileInfos[i].Name())
	}
	return dirList, nil
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	return err == nil && info.IsDir()
}

func fileExists(dirname string) bool {
	info, err := os.Stat(dirname)
	return err == nil && !info.IsDir()
}

func AppendStringToFile(file string, data string) error {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(data))
	f.Close()
	return err
}

func WriteStringToFile(file string, data string) error {
	f, err := os.OpenFile(file, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(data))
	f.Close()
	return err
}

func ReadFileToLines(file string) ([]string, error) {

	fileObj := fileObject{
		Path: file,
	}
	data, _ := fileObj.Read()
	lines := strings.Split(data, "\n")
	return lines, nil
}
