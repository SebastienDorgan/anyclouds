package cloudinit

import (
	"io"

	"github.com/SebastienDorgan/anyclouds/serverconf"
)

//Configurationfactory abstract the creation of a server init script (cloud-init, shell, ...)
type Configurationfactory struct {
}

//Build create a cloud-init configuration
func (f *Configurationfactory) Build(cfg *serverconf.ServerConfiguration) (io.ByteReader, error) {
	return nil, nil
}
