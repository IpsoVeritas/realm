package assets

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

type AssetsProvider struct {
	path string
}

func NewAssetsProvider(path string) *AssetsProvider {
	return &AssetsProvider{
		path: path,
	}
}

func (a *AssetsProvider) Read(name string) ([]byte, error) {
	return ioutil.ReadFile(fmt.Sprintf("%s/%s", a.path, name))
}

func (a *AssetsProvider) List(name string) ([]string, error) {
	files, err := ioutil.ReadDir(fmt.Sprintf("%s/%s", a.path, name))
	if err != nil {
		return nil, err
	}

	out := make([]string, 0)
	for _, f := range files {
		out = append(out, f.Name())
	}

	return out, nil
}

func (a *AssetsProvider) CopyToTempFile(name string) (string, error) {
	dir, err := ioutil.TempDir(".", ".templates-")
	if err != nil {
		return "", errors.Wrap(err, "Could not create temporary directory")
	}

	filename := fmt.Sprintf("%s/%s", dir, filepath.Base(name))

	in, err := os.Open(fmt.Sprintf("%s/%s", a.path, name))
	if err != nil {
		return "", err
	}
	defer in.Close()

	out, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return "", err
	}

	return filename, out.Close()
}
