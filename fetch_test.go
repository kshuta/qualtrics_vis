package main

import (
	"os/exec"
	"testing"
)

func TestProcessing(t *testing.T) {
	t.Run("Processing the data", func(t *testing.T) {
		if err := exec.Command("python3", "process_data.py", "test_data.csv", "testdb.db").Run(); err != nil {
			t.Error(err)
		}
	})

	t.Run("processing the data twice.", func(t *testing.T) {
		if err := exec.Command("python3", "process_data.py", "test_data.csv", "testdb.db").Run(); err != nil {
			t.Error(err)
		}

		de := setupDB("sqlite3", "./testdb.db")
		defer de()

		// second round
		if err := exec.Command("python3", "process_data.py", "test_data.csv", "testdb.db").Run(); err != nil {
			t.Error(err)
		}

		post_record, err := getCompleteRecord("peer", "8", "2019")
		if err != nil {
			t.Fatal(err)
		}

		if len(post_record.Counts) > 2 {
			t.Error("Created different topic count")
		}

	})
}
