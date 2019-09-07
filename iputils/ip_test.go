package iputils_test

import (
	"github.com/SebastienDorgan/anyclouds/iputils"
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

func TestConversions(t *testing.T) {
	ip := net.ParseIP("192.168.0.1")
	ui := iputils.Itou(&ip)
	ip2 := *iputils.Utoi(ui + 1)
	assert.Equal(t, "192.168.0.2", ip2.String())
}

func TestRange(t *testing.T) {
	r, err := iputils.GetRange("192.168.0.0/24")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.1", r.FirstIP.String())
	assert.Equal(t, "192.168.0.254", r.LastIP.String())
	r, err = iputils.GetRange("192.168.0.0/16")
	assert.NoError(t, err)
	assert.Equal(t, "192.168.0.1", r.FirstIP.String())
	assert.Equal(t, "192.168.255.254", r.LastIP.String())
	r, err = iputils.GetRange("192.0.0.0/8")
	assert.NoError(t, err)
	assert.Equal(t, "192.0.0.1", r.FirstIP.String())
	assert.Equal(t, "192.255.255.254", r.LastIP.String())
}
