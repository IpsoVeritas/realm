package realm

type AssetProvider interface {
	Read(string) ([]byte, error)
	List(string) ([]string, error)
	CopyToTempFile(string) (string, error)
}
