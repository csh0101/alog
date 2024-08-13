package writers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var _ io.WriteCloser = &FileWriter{}

type FileWriter struct {
	once   sync.Once
	ctx    context.Context
	cancel context.CancelFunc

	logger              *log.Logger
	dir                 string
	fileMaxSizeInBytes  int64
	fileRetention       time.Duration
	fileTotalCountLimit int
	filePrefix          string
	fileExt             string

	mutex sync.RWMutex
	f     *safeCloseFile
}

func NewFileWriter(dir string, opts ...FileWriterOption) (*FileWriter, error) {
	if dir == "" {
		return nil, errors.New("params dir is required")
	}

	w := &FileWriter{
		logger:              log.New(io.Discard, "", log.LstdFlags),
		dir:                 dir,
		fileMaxSizeInBytes:  2 * 1024 * 1024 * 1024,
		fileRetention:       7 * 24 * time.Hour,
		fileTotalCountLimit: 10000,
	}
	for _, opt := range opts {
		opt(w)
	}

	if err := w.init(); err != nil {
		return nil, fmt.Errorf("init failed, %w", err)
	}

	return w, nil
}

func (w *FileWriter) init() error {
	if err := os.MkdirAll(w.dir, os.ModePerm); err != nil {
		return fmt.Errorf("create dir failed, %w", err)
	}

	fileName, fileSequence, err := w.analysisFiles()
	if err != nil {
		return fmt.Errorf("analysis files failed, %w", err)
	}
	f, err := newSafeCloseFile(w.filePath(fileName, fileSequence))
	if err != nil {
		return fmt.Errorf("open current file failed, %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	w.ctx = ctx
	w.cancel = cancel
	w.f = f

	go w.setupAutomationWorker()

	return nil
}

func (w *FileWriter) setupAutomationWorker() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	tickCount := int64(0)

	for {
		select {
		case <-w.ctx.Done():
			w.logger.Println("[I] automation worker exit")
			return
		case <-ticker.C:
			w.mutex.Lock()

			// 自动按天rotate
			w.autoRotateByDayWithoutLock()

			// 自动清理过期文件
			tickCount++
			if tickCount > 0 && tickCount%1 == 0 {
				w.autoRetentionWithoutLock()
			}

			w.mutex.Unlock()
		}
	}
}

func (w *FileWriter) autoRotateByDayWithoutLock() {
	if timeMatch := strings.HasPrefix(filepath.Base(w.f.Name()), w.fileName()); !timeMatch {
		w.logger.Println("[D] auto-rotate by day")
		w.rorateWithoutLock(true)
	}
}

func (w *FileWriter) autoRetentionWithoutLock() {
	dirEntries, err := os.ReadDir(w.dir)
	if err != nil {
		w.logger.Printf("[E] scan directory failed, %v", err)
		return
	}

	// fileInfoList was sorted by filename asc
	fileInfoList := make([]os.FileInfo, 0, len(dirEntries))
	for _, entry := range dirEntries {
		fInfo, err := entry.Info()
		if err != nil {
			w.logger.Printf("[E] get file info failed, %v", err)
			continue
		}
		if fInfo.IsDir() {
			continue
		}
		if (w.filePrefix != "" && !strings.HasPrefix(fInfo.Name(), w.filePrefix)) ||
			(w.fileExt != "" && !strings.HasSuffix(fInfo.Name(), w.fileExt)) {
			continue
		}
		fileInfoList = append(fileInfoList, fInfo)
	}

	now := time.Now()
	for idx, info := range fileInfoList {
		// Forbidden clearing the file that is in using
		if filepath.Base(info.Name()) == filepath.Base(w.f.Name()) {
			continue
		}
		// Clear condition:
		//   (1) when the file was over the total count limit
		//   (2) when the file was over the retention time
		if (w.fileTotalCountLimit > 0 && len(fileInfoList)-idx > w.fileTotalCountLimit) ||
			(w.fileRetention > 0 && info.ModTime().Add(w.fileRetention).Before(now)) {

			_ = os.Remove(filepath.Join(w.dir, info.Name()))
			w.logger.Printf("[D] retention clear file `%s`\n", info.Name())
			continue
		}
	}
}

func (w *FileWriter) Write(b []byte) (int, error) {
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if int64(len(b))+w.f.Size() < w.fileMaxSizeInBytes {
		return w.f.Write(b)
	}
	if int64(len(b)) > w.fileMaxSizeInBytes {
		return 0, fmt.Errorf("data to write is too large, it exceeds the max file size setting. Limit is %d bytes", w.fileMaxSizeInBytes)
	}
	select {
	case <-w.ctx.Done():
		return 0, w.ctx.Err()
	default:
		w.rorateWithoutLock(false)
	}

	return w.f.Write(b)
}

func (w *FileWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.once.Do(func() {
		if w.cancel != nil {
			w.cancel()
		}
		if w.f != nil {
			w.f.Close()
		}
	})
	return nil
}

func (w *FileWriter) rorateWithoutLock(isDayRotate bool) {
	fileName, fileSequence, err := w.analysisFiles()
	if err != nil {
		w.logger.Printf("[E] analysis files failed, %v\n", err)
		return
	}
	if !isDayRotate {
		fileSequence++
	}
	f, err := newSafeCloseFile(w.filePath(fileName, fileSequence))
	if err != nil {
		w.logger.Printf("[E] open file to write failed, %v\n", err)
		return
	}
	if w.f != nil {
		w.f.Close()
	}
	w.f = f
}

func (w *FileWriter) analysisFiles() (fileName string, fileSequence int, err error) {
	fileSequence = 0
	fileName = w.fileName()

	dirEntries, err := os.ReadDir(w.dir)
	if err != nil {
		return "", 0, fmt.Errorf("scan dir failed, %w", err)
	}
	for _, entry := range dirEntries {
		fileInfo, err := entry.Info()
		if err != nil {
			continue
		}
		if !strings.HasPrefix(fileInfo.Name(), fileName) {
			continue
		}
		fields := strings.Split(strings.TrimSuffix(fileInfo.Name(), filepath.Ext(fileInfo.Name())), "-")
		if len(fields) == 0 {
			continue
		}
		seq, err := strconv.Atoi(fields[len(fields)-1])
		if err != nil {
			continue
		}
		if seq < fileSequence {
			continue
		}
		if fileInfo.Size() >= w.fileMaxSizeInBytes {
			fileSequence = seq + 1
		} else {
			fileSequence = seq
		}
	}
	return fileName, fileSequence, nil
}

func (w *FileWriter) filePath(fileNamed string, fileSequence int) string {
	fPath := filepath.Join(
		w.dir,
		fmt.Sprintf("%s-%04d", w.fileName(), fileSequence),
	) + w.fileExt

	absPath, err := filepath.Abs(fPath)
	if err != nil {
		return fPath
	}
	return absPath
}

func (w *FileWriter) fileName() string {
	fNameFields := make([]string, 0, 3)
	if w.filePrefix != "" {
		fNameFields = append(fNameFields, w.filePrefix)
	}
	fNameFields = append(fNameFields, time.Now().UTC().Format("20060102"))
	return strings.Join(fNameFields, "-")
}

type FileWriterOption func(w *FileWriter)

func WithFileMaxSizeInBytes(v int64) FileWriterOption {
	return func(w *FileWriter) {
		if v > 0 {
			w.fileMaxSizeInBytes = v
		}
	}
}

func WithFileRetention(v time.Duration) FileWriterOption {
	return func(w *FileWriter) {
		if v >= 0 {
			w.fileRetention = v
		}
	}
}

func WithFileTotalCountLimit(v int) FileWriterOption {
	return func(w *FileWriter) {
		if v <= 0 {
			w.fileTotalCountLimit = 10000
		} else {
			w.fileTotalCountLimit = v
		}
	}
}

func WithFilePrefix(v string) FileWriterOption {
	return func(w *FileWriter) {
		w.filePrefix = strings.TrimSpace(v)
	}
}

func WithFileExt(v string) FileWriterOption {
	return func(w *FileWriter) {
		w.fileExt = strings.TrimSpace(v)
	}
}

func WithLogWriter(writer io.Writer) FileWriterOption {
	return func(w *FileWriter) {
		if writer != nil {
			w.logger.SetOutput(writer)
		}
	}
}

type safeCloseFile struct {
	once     sync.Once
	info     os.FileInfo
	fileSize int64
	*os.File
}

func newSafeCloseFile(path string) (*safeCloseFile, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file failed, %w", err)
	}

	return &safeCloseFile{
		info:     info,
		File:     f,
		fileSize: info.Size(),
	}, nil
}

func (f *safeCloseFile) Close() error {
	f.once.Do(func() {
		if f.File != nil {
			f.File.Close()
		}
	})
	return nil
}

func (f *safeCloseFile) Write(b []byte) (n int, err error) {
	if f.File == nil {
		return 0, errors.New("file is nil")
	}
	n, err = f.File.Write(b)
	atomic.AddInt64(&f.fileSize, int64(n))
	return n, err
}

func (f *safeCloseFile) Size() int64 {
	return atomic.LoadInt64(&f.fileSize)
}
