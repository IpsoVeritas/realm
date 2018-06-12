//go:generate go-bindata -prefix "" -pkg bindata -o assets.go ../../../assets/...

package bindata

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
)

type BindataProvider struct{}

func NewBindataProvider() *BindataProvider {
	return &BindataProvider{}
}

func (b *BindataProvider) Read(name string) ([]byte, error) {
	return Asset(fmt.Sprintf("../../../assets/%s", name))
}

func (b *BindataProvider) List(name string) ([]string, error) {
	return AssetDir(fmt.Sprintf("../../../assets/%s", name))
}

func (b *BindataProvider) CopyToTempFile(name string) (string, error) {
	bytes, err := b.Read(name)
	if err != nil {
		return "", errors.Wrapf(err, "Could not get template file %s", name)
	}

	dir, err := ioutil.TempDir(".", ".templates-")
	if err != nil {
		return "", errors.Wrap(err, "Could not create temporary directory")
	}

	filename := fmt.Sprintf("%s/%s", dir, filepath.Base(name))
	err = ioutil.WriteFile(filename, bytes, 0644)
	if err != nil {
		return "", errors.Wrap(err, "Could not write temporary file")
	}

	return filename, nil
}
