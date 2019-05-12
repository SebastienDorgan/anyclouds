package cloudinit

import (
	"io"

	"github.com/SebastienDorgan/anyclouds/serverconf"
)

//ConfigurationFactory abstract the creation of a server init script (cloud-init, shell, ...)
type ConfigurationFactory struct {
}

//Build create a cloud-init configuration
func (f *Configurationfactory) Build(cfg *serverconf.ServerConfiguration) (io.ByteReader, error) {
	return nil, nil
}
