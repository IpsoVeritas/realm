package filestore

import (
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Brickchain/go-logger.v1"

	"github.com/julienschmidt/httprouter"
)

type Filesystem struct {
	base string
	dir  string
}

func NewFilesystem(base, dir string) (*Filesystem, error) {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return nil, err
	}

	f := &Filesystem{
		base: base,
		dir:  dir,
	}

	return f, nil
}

func (f *Filesystem) Read(name string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("%s/%s", f.dir, name))
}

func (f *Filesystem) Write(name string, input io.Reader) (string, error) {
	dir := filepath.Dir(name)
	err := os.MkdirAll(fmt.Sprintf("%s/%s", f.dir, dir), 0755)
	if err != nil {
		return "", err
	}

	file, err := os.OpenFile(fmt.Sprintf("%s/%s", f.dir, name), os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return "", err
	}
	defer file.Close()

	file.Truncate(0)
	file.Seek(0, 0)
	_, err = io.Copy(file, input)
	if err != nil {
		return "", err
	}
	file.Sync()

	fullname := fmt.Sprintf("%s/%s", f.base, name)

	return fullname, err
}

func (f *Filesystem) Handler(w http.ResponseWriter, r *http.Request, params httprouter.Params) error {
	name := params.ByName("filename")
	if name == "" {
		logger.Error("No filename given")
		http.Error(w, "No filename given", http.StatusBadRequest)
		return nil
	}

	bytes, err := f.Read(name)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}

	mimeType := mime.TypeByExtension(name)
	w.Header().Set("Content-Type", mimeType)

	w.Write(bytes)

	return nil
}
