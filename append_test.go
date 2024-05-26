package main

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type logEntry struct {
	ID   int    `json:"id"`
	Data string `json:"data"`
}

// setupLog creates a new AppendOnlyLog for testing.
func setupLog(t *testing.T) *AppendOnlyLog {
	t.Helper()
	filePath := "test_log.json"
	log, err := NewAppendOnlyLog(filePath)
	if err != nil {
		t.Fatalf("Failed to create AppendOnlyLog: %v", err)
	}

	t.Cleanup(func() {
		log.file.Close()
		os.Remove(log.file.Name())
	})

	return log
}

func TestAppendOnlyLog_Write(t *testing.T) {
	log := setupLog(t)

	tests := []struct {
		name  string
		entry logEntry
	}{
		{"Write entry 1", logEntry{ID: 1, Data: "test data 1"}},
		{"Write entry 2", logEntry{ID: 2, Data: "test data 2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := log.Write(tt.entry)
			if err != nil {
				t.Fatalf("Write() error = %v", err)
			}
		})
	}
}

func TestAppendOnlyLog_Read(t *testing.T) {
	log := setupLog(t)

	entries := []logEntry{
		{ID: 1, Data: "test data 1"},
		{ID: 2, Data: "test data 2"},
	}

	for _, entry := range entries {
		log.Write(entry)
	}

	tests := []struct {
		name   string
		offset int64
		count  int
		want   []logEntry
	}{
		{"Read first entry", 0, 1, entries[:1]},
		{"Read both entries", 0, 2, entries},
		{"Read from offset", int64(len(entries[0].Data) + 12), 1, entries[1:]},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := log.Read(tt.offset, tt.count)
			if err != nil {
				t.Fatalf("Read() error = %v", err)
				return
			}

			var gotEntries []logEntry
			for _, g := range got {
				var entry logEntry
				if err := json.Unmarshal(g, &entry); err != nil {
					t.Fatalf("Failed to unmarshal entry: %v", err)
				}
				gotEntries = append(gotEntries, entry)
			}

			if len(gotEntries) != len(tt.want) {
				t.Fatalf("Read() got %v entries, want %v entries", len(gotEntries), len(tt.want))
			}

			for i, entry := range gotEntries {
				if !reflect.DeepEqual(entry, tt.want[i]) {
					t.Fatalf("Read() got entry %v, want %v", entry, tt.want[i])
				}
			}
		})
	}
}

func TestAppendOnlyLog_Count(t *testing.T) {
	log := setupLog(t)

	entries := []logEntry{
		{ID: 1, Data: "test data 1"},
		{ID: 2, Data: "test data 2"},
	}

	for _, entry := range entries {
		log.Write(entry)
	}

	tests := []struct {
		name string
		want int
	}{
		{"Count entries", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := log.Count()
			if err != nil {
				t.Fatalf("Count() error = %v", err)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Count() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendOnlyLog_Seek(t *testing.T) {
	log := setupLog(t)

	entries := []logEntry{
		{ID: 1, Data: "test data 1"},
		{ID: 2, Data: "test data 2"},
	}

	for _, entry := range entries {
		log.Write(entry)
	}

	tests := []struct {
		name   string
		offset int64
	}{
		{"Seek to start", 0},
		{"Seek to second entry", int64(len(entries[0].Data) + 12)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := log.Seek(tt.offset)
			if err != nil {
				t.Fatalf("Seek() error = %v", err)
			}

			got, err := log.Read(tt.offset, 1)
			if err != nil {
				t.Fatalf("Read() after Seek error = %v", err)
				return
			}

			var gotEntry logEntry
			if err := json.Unmarshal(got[0], &gotEntry); err != nil {
				t.Fatalf("Failed to unmarshal entry: %v", err)
			}

			if !reflect.DeepEqual(gotEntry, entries[tt.offset/(int64(len(entries[0].Data)+12))]) {
				t.Fatalf("Seek() got entry %v, want %v", gotEntry, entries[tt.offset/(int64(len(entries[0].Data)+12))])
			}
		})
	}
}
