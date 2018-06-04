package filestore

import "io"

type Filestore interface {
	Read(string) ([]byte, error)
	Write(string, io.Reader) (string, error)
}
