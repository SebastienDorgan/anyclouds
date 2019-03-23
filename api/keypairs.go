package api

//KeyPair a key pair object
type KeyPair struct {
	Name        string
	Fingerprint string
}

//KeyPairManager manage ssh keys
type KeyPairManager interface {
	Load(name string, publicKey []byte) error
	Delete(name string) error
	List() ([]KeyPair, error)
}
