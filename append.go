// Package main provides an append-only log file system.
// It allows for writing, reading, watching, and seeking operations on a log file.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
)

// AppendOnlyLog represents an append-only log file.
// It provides concurrent-safe operations for writing, reading, watching, and seeking in the log file.
type AppendOnlyLog struct {
	file   *os.File
	mutex  sync.Mutex
	offset int64
}

// NewAppendOnlyLog creates a new AppendOnlyLog instance.
// It opens the file located at filePath. If the file doesn't exist, it creates a new one.
// It returns the new AppendOnlyLog instance or an error if one occurred.
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
// It locks the log file, writes the entry, and then unlocks the file.
// It returns an error if one occurred during the marshalling or writing process.
func (log *AppendOnlyLog) Write(entry interface{}) error {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Save the current offset
	oldOffset := log.offset

	// Update the offset
	log.offset += int64(len(data) + 1)

	_, err = log.file.Write(data)
	if err != nil {
		// Roll back the offset if the write operation failed
		log.offset = oldOffset
		return err
	}

	_, err = log.file.Write([]byte("\n"))
	if err != nil {
		// Roll back the offset if the write operation failed
		log.offset = oldOffset
		return err
	}

	return nil
}

// Read reads entries from the log starting from the given offset.
// It reads count number of entries or until it reaches the end of the file.
// It returns the read entries or an error if one occurred during the reading process.
func (log *AppendOnlyLog) Read(offset int64, count int) ([]interface{}, error) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	if _, err := log.file.Seek(offset, io.SeekStart); err != nil {
		return nil, err
	}

	log.offset = offset

	var entries []interface{}
	decoder := json.NewDecoder(log.file)

	for i := 0; i < count; i++ {
		var entry interface{}
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

// Watch starts a new goroutine that watches for new entries in the log file.
// It sends new entries through the returned channel.
// The watching can be stopped by sending a signal to the provided stopCh channel.
// It returns an error if one occurred during the watching process.
func (log *AppendOnlyLog) Watch(stopCh <-chan struct{}) (<-chan interface{}, error) {
	outCh := make(chan interface{})

	go func() {
		defer close(outCh)
		decoder := json.NewDecoder(log.file)

		for {
			select {
			case <-stopCh:
				return
			default:
				var entry interface{}
				if err := decoder.Decode(&entry); err != nil {
					if err == io.EOF {
						continue
					}
					fmt.Println("Error decoding entry:", err)
					return
				}
				outCh <- entry
			}
		}
	}()

	return outCh, nil
}

// Seek moves the read/write offset to a specific position in the log file.
// It returns an error if one occurred during the seeking process.
func (log *AppendOnlyLog) Seek(offset int64) error {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	_, err := log.file.Seek(offset, io.SeekStart)
	return err
}

// Count returns the number of entries in the log file.
// It returns an error if one occurred during the counting process.
func (log *AppendOnlyLog) Count() (int, error) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	if _, err := log.file.Seek(0, io.SeekStart); err != nil {
		return 0, err
	}

	count := 0
	decoder := json.NewDecoder(log.file)
	for {
		var entry json.RawMessage
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return 0, err
		}
		count++
	}

	return count, nil
}
