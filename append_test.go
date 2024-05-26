package main

import (
	"encoding/json"
	"math/rand"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func FakeRaw(t *testing.T) map[string]interface{} {
	t.Helper()

	return map[string]interface{}{
		"str":   "bar",
		"int":   float64(rand.Intn(100)),
		"float": float64(rand.Intn(100)),
		"bool":  float64(rand.Intn(2)) == 1.0,
		"list":  []interface{}{"foo", "bar", "baz"},
		"complex": map[string]interface{}{
			"str":   "foo",
			"int":   float64(rand.Intn(100)),
			"float": float64(rand.Intn(100)),
			"bool":  float64(rand.Intn(2)) == 1.0,
			"list":  []interface{}{"qux", "quux", "quuz"},
			"complex": map[string]interface{}{
				"str":   "baz",
				"int":   float64(rand.Intn(100)),
				"float": float64(rand.Intn(100)),
				"bool":  float64(rand.Intn(2)) == 1.0,
				"list":  []interface{}{1.0, 2.0, 3.0},
			},
		},
	}
}

func GenerateInput(t *testing.T, n int) []interface{} {
	t.Helper()

	var input []interface{}
	for i := 0; i < n; i++ {
		input = append(input, FakeRaw(t))
	}
	return input
}

func TestAppendOnlyLog_WriteAndRead(t *testing.T) {
	t.Helper()

	// Create a temporary file for the log
	file, err := os.CreateTemp("", "log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	// Create a new log
	log, err := NewAppendOnlyLog(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	entries := GenerateInput(t, 50)

	// Write the entries to the log
	for _, entry := range entries {
		if err := log.Write(entry); err != nil {
			t.Fatal(err)
		}
	}

	// Read the entries back
	readEntries, err := log.Read(0, len(entries))
	if err != nil {
		t.Fatal(err)
	}

	// Verify that the read entries match the ones we wrote
	for i, readEntry := range readEntries {
		diff := cmp.Diff(entries[i], readEntry)
		if diff != "" {
			t.Fatalf(diff)
		}
	}

	// Test watch functionality
	// Create a channel to receive the entries
	stopCh := make(chan struct{})

	// Start watching for new entries
	watchCh, err := log.Watch(stopCh)
	if err != nil {
		t.Fatal(err)
	}

	// Write more entries to the log
	newInput := GenerateInput(t, 50)
	go func() {
		for _, entry := range newInput {
			if err := log.Write(entry); err != nil {
				t.Error(err)
			}
		}

		stopCh <- struct{}{}
	}()

	// Read the new entries from the watch channel
	var newEntries []json.RawMessage
	for entry := range watchCh {
		if rawEntry, ok := entry.(json.RawMessage); ok {
			newEntries = append(newEntries, rawEntry)
		} else {
			t.Fatal("entry is not of type json.RawMessage")
		}
	}

	// Verify that the new entries match the ones we wrote
	for i, entry := range newEntries {
		diff := cmp.Diff(newInput[i], entry)
		if diff != "" {
			t.Fatalf(diff)
		}
	}

}
