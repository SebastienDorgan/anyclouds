package api

//KeyPairManager manage ssh keys
type KeyPairManager interface {
	Load(name string, publicKey []byte) error
	Delete(name string) error
	List() ([]string, error)
}
