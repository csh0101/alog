package writers_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/csh0101/alog/writers"
)

func TestFileWriter(t *testing.T) {
	var (
		err        error
		writeToDir = "./testdata/"
	)

	os.RemoveAll(writeToDir)
	err = testFileWriter(writeToDir)
	os.RemoveAll(writeToDir)

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestFileTotalCountLimit(t *testing.T) {
	var (
		err        error
		writeToDir = "./testdata/"
	)

	os.RemoveAll(writeToDir)
	err = testFileTotalCountLimit(writeToDir)
	os.RemoveAll(writeToDir)

	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestFileRetention(t *testing.T) {
	var (
		err        error
		writeToDir = "./testdata/"
	)

	os.RemoveAll(writeToDir)
	err = testFileRetention(writeToDir)
	os.RemoveAll(writeToDir)

	if err != nil {
		t.Fatal(err.Error())
	}
}

func testFileWriter(dir string) error {
	var (
		contentToWrite     = []byte("Hello, this is a file writer test")
		fileRetention      = 5 * time.Second
		maxFileSizeInBytes = 50
	)

	w, err := writers.NewFileWriter(
		dir,
		writers.WithFilePrefix("test"),
		writers.WithFileExt(".log"),
		writers.WithFileRetention(fileRetention),
		writers.WithFileMaxSizeInBytes(int64(maxFileSizeInBytes)),
		writers.WithLogWriter(os.Stderr),
	)
	if err != nil {
		return fmt.Errorf("new file writer failed, %w", err)
	}
	defer w.Close()

	for i := 0; i < 3; i++ {
		if _, err = w.Write(contentToWrite); err != nil {
			return fmt.Errorf("write failed, %w", err)
		}
	}

	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir failed, %w", err)
	}

	// check rotated file count
	if len(fileInfoList) != 3 {
		return fmt.Errorf("rotate error: file count mismatch, expected %d, got %d", 3, len(fileInfoList))
	}

	// check file retention
	<-time.After(fileRetention + 5*time.Second)

	fileInfoList, err = ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir failed, %w", err)
	}
	for _, info := range fileInfoList {
		fmt.Println(info.Name())
	}
	if len(fileInfoList) != 0 {
		return fmt.Errorf("retention error: file count mismatch, expected %d, got %d", 0, len(fileInfoList))
	}
	return nil
}

func testFileTotalCountLimit(dir string) error {
	var (
		contentToWrite     = []byte("Hello, this is a file writer test")
		fileRetention      = time.Hour
		maxFileSizeInBytes = 50
		maxFileTotalCount  = 2
	)

	w, err := writers.NewFileWriter(
		dir,
		writers.WithFilePrefix("test"),
		writers.WithFileExt(".log"),
		writers.WithFileRetention(fileRetention),
		writers.WithFileMaxSizeInBytes(int64(maxFileSizeInBytes)),
		writers.WithFileTotalCountLimit(maxFileTotalCount),
		writers.WithLogWriter(os.Stderr),
	)
	if err != nil {
		return fmt.Errorf("new file writer failed, %w", err)
	}
	defer w.Close()

	for i := 0; i < 100; i++ {
		if _, err = w.Write(contentToWrite); err != nil {
			return fmt.Errorf("write failed, %w", err)
		}
	}

	// check file total count limit
	<-time.After(3 * time.Second)

	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir failed, %w", err)
	}
	for _, info := range fileInfoList {
		fmt.Println(info.Name())
	}
	if len(fileInfoList) != maxFileTotalCount {
		return fmt.Errorf("retention error: file total count mismatch, expected %d, got %d", maxFileTotalCount, len(fileInfoList))
	}
	return nil
}
func testFileRetention(dir string) error {
	var (
		contentToWrite     = []byte("Hello, this is a file writer test")
		fileRetention      = 2 * time.Second
		maxFileSizeInBytes = 50
		maxFileTotalCount  = 100000
	)

	w, err := writers.NewFileWriter(
		dir,
		writers.WithFilePrefix("test"),
		writers.WithFileExt(".log"),
		writers.WithFileRetention(fileRetention),
		writers.WithFileMaxSizeInBytes(int64(maxFileSizeInBytes)),
		writers.WithFileTotalCountLimit(maxFileTotalCount),
		writers.WithLogWriter(os.Stderr),
	)
	if err != nil {
		return fmt.Errorf("new file writer failed, %w", err)
	}
	defer w.Close()

	for i := 0; i < 10; i++ {
		if _, err = w.Write(contentToWrite); err != nil {
			return fmt.Errorf("write failed, %w", err)
		}
	}

	// check file total count limit
	<-time.After(3 * time.Second)

	fileInfoList, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir failed, %w", err)
	}
	for _, info := range fileInfoList {
		fmt.Println(info.Name())
	}
	if len(fileInfoList) != 1 {
		return fmt.Errorf("retention error: file total count mismatch, expected %d, got %d", 1, len(fileInfoList))
	}
	return nil
}
