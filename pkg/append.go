package main

import (
	"encoding/json"
	"io"
	"os"
	"sync"
)

// AppendOnlyLog handles a single file append-only data engine.
type AppendOnlyLog struct {
	file   *os.File
	mutex  sync.Mutex
	offset int64
}

// NewAppendOnlyLog creates a new AppendOnlyLog instance.
func NewAppendOnlyLog(filePath string) (*AppendOnlyLog, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	return &AppendOnlyLog{
		file:   file,
		offset: stat.Size(),
	}, nil
}

// Write appends a JSON entry to the log.
func (log *AppendOnlyLog) Write(entry interface{}) error {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	_, err = log.file.Write(data)
	if err != nil {
		return err
	}

	_, err = log.file.Write([]byte("\n"))
	if err != nil {
		return err
	}

	log.offset += int64(len(data) + 1)
	return nil
}

// Read reads entries from the log. The offset is the line offset from the beginning of the file.
// The count is the number of entries to read.
func (log *AppendOnlyLog) Read(offset int64, count int) ([]json.RawMessage, error) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	if _, err := log.file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	var entries []json.RawMessage
	decoder := json.NewDecoder(log.file)

	for i := 0; i < count; i++ {
		var entry json.RawMessage
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

