package services

import (
	"fmt"
	"io"
	"strings"

	filestore "gitlab.brickchain.com/libs/go-filestore.v1"
)

type FileService struct {
	p       filestore.Filestore
	realmID string
}

func (f *FileService) Read(name string) ([]byte, error) {
	return f.p.Read(fmt.Sprintf("%s/%s", strings.Replace(f.realmID, ":", "_", -1), name))
}

func (f *FileService) Write(name string, file io.Reader) (string, error) {
	return f.p.Write(fmt.Sprintf("%s/%s", strings.Replace(f.realmID, ":", "_", -1), name), file)
}
