package filestore

import (
	"bytes"
	"os"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestNewGCS(t *testing.T) {
	if _, err := os.Stat("test_gcs_credentials.json"); err != nil {
		t.Skip("Could not find test_gcs_credentials.json")
	}

	id := uuid.Must(uuid.NewV4()).String()
	g, err := NewGCS(id, "EU", "brickchain-ci", "test_gcs_credentials.json")
	defer func() {
		err = g.DeleteBucket()
		if err != nil {
			t.Logf("ERROR: %s", err.Error())
		}
	}()
	if err != nil {
		t.Fatal(err)
	}
}

func TestGCS_Write(t *testing.T) {
	if _, err := os.Stat("test_gcs_credentials.json"); err != nil {
		t.Skip("Could not find test_gcs_credentials.json")
	}

	id := uuid.Must(uuid.NewV4()).String()
	g, err := NewGCS(id, "EU", "brickchain-ci", "test_gcs_credentials.json")
	defer func() {
		err = g.Delete("write_test.txt")
		if err != nil {
			t.Logf("ERROR: %s", err.Error())
		}

		err = g.DeleteBucket()
		if err != nil {
			t.Logf("ERROR: %s", err.Error())
		}
	}()
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewBufferString("write test string")

	url, err := g.Write("write_test.txt", r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Created file at %s", url)
}

func TestGCS_WriteWithDir(t *testing.T) {
	if _, err := os.Stat("test_gcs_credentials.json"); err != nil {
		t.Skip("Could not find test_gcs_credentials.json")
	}

	id := uuid.Must(uuid.NewV4()).String()
	g, err := NewGCS(id, "EU", "brickchain-ci", "test_gcs_credentials.json")
	defer func() {
		err = g.Delete("things/write_test.txt")
		if err != nil {
			t.Logf("ERROR: %s", err.Error())
		}

		err = g.DeleteBucket()
		if err != nil {
			t.Logf("ERROR: %s", err.Error())
		}
	}()
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewBufferString("write test string")

	url, err := g.Write("things/write_test.txt", r)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Created file at %s", url)
}

func TestGCS_Read(t *testing.T) {
	if _, err := os.Stat("test_gcs_credentials.json"); err != nil {
		t.Skip("Could not find test_gcs_credentials.json")
	}

	id := uuid.Must(uuid.NewV4()).String()
	g, err := NewGCS(id, "EU", "brickchain-ci", "test_gcs_credentials.json")
	defer func() {
		err = g.Delete("read_test.txt")
		if err != nil {
			t.Logf("ERROR: %s", err.Error())
		}

		err = g.DeleteBucket()
		if err != nil {
			t.Logf("ERROR: %s", err.Error())
		}
	}()
	if err != nil {
		t.Fatal(err)
	}

	r := bytes.NewBufferString("read test string")

	_, err = g.Write("read_test.txt", r)
	if err != nil {
		t.Fatal(err)
	}

	out, err := g.Read("read_test.txt")
	if err != nil {
		t.Fatal(err)
	}

	if string(out) != "read test string" {
		t.Fatalf("Read string not same as written: '%s' != 'read test string", string(out))
	}
}
