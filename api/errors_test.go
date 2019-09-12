package api_test

import (
	"fmt"
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func imageErrorWithCausef() error {
	return api.NewGetImageError(fmt.Errorf("terrible error"), "i-4545644")
}

func sprt(s string) *string {
	return &s
}

func TestErrors(t *testing.T) {
	err := imageErrorWithCausef()
	assert.Error(t, err)
	err = api.NewCreateNetworkInterfaceError(err, api.CreateNetworkInterfaceOptions{
		Name:             "a name",
		NetworkID:        "1234",
		SubnetID:         "sn-456",
		ServerID:         sprt("srv-678"),
		SecurityGroupID:  "",
		Primary:          false,
		PrivateIPAddress: nil,
	})

	err = api.NewUpdateNetworkInterfaceError(err, api.UpdateNetworkInterfaceOptions{
		ServerID:        sprt("srv-6789"),
		SecurityGroupID: nil,
	})

	err = api.NewListNetworkInterfacesError(err, nil)
	err = api.NewErrorStackFromError(err, nil)
	err = api.NewDeleteNetworkInterfaceError(err, "ni-5678")
	type Object struct {
		A bool
		B string
	}
	err = api.NewErrorStack(err, "end of stack", "Abc", 456, 23.8, Object{
		A: false,
		B: "Bcd",
	})
	fmt.Println(err.Error())

}

func TestFuncForPC(t *testing.T) {
	pc, file, no, ok := runtime.Caller(0)
	if ok {
		fmt.Printf("called from %s#%d\n", file, no)
	}
	r := runtime.FuncForPC(pc)
	fmt.Println(r.Name(), r.Entry())
}
