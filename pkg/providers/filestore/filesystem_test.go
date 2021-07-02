package filestore

import "testing"
import "os"
import "bytes"

func TestNewFilesystem(t *testing.T) {
	_, err := NewFilesystem("", ".test-files")
	defer os.RemoveAll(".test-files")
	if err != nil {
		t.Fatal(err)
	}
}

func TestFilesystem_Write(t *testing.T) {
	f, err := NewFilesystem("", ".test-files")
	defer os.RemoveAll(".test-files")
	if err != nil {
		t.Fatal(err)
	}

	data := bytes.NewBufferString("write test string")

	_, err = f.Write("write_test.txt", data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestFilesystem_WriteWithDir(t *testing.T) {
	f, err := NewFilesystem("", ".test-files")
	defer os.RemoveAll(".test-files")
	if err != nil {
		t.Fatal(err)
	}

	data := bytes.NewBufferString("write test string")

	_, err = f.Write("things/write_test.txt", data)
	if err != nil {
		t.Fatal(err)
	}
}
