package providers

import (
	"github.com/SebastienDorgan/anyclouds/api"
	"github.com/SebastienDorgan/retry"
	"github.com/pkg/errors"
	"time"
)

const ResourceReachStableStateTimeout = 3 * time.Minute

//WaitUntilServerReachStableState wait until server reach stable state
func WaitUntilServerReachStableState(mgr api.ServerManager, serverID string) (*api.Server, error) {
	get := func() (interface{}, error) {
		return mgr.Get(serverID)
	}
	finished := func(v interface{}, e error) bool {
		state := v.(*api.Server).State
		return state != api.ServerPending
	}
	res := retry.With(get).For(ResourceReachStableStateTimeout).Every(time.Second).Until(finished).Go()
	if res.Timeout || res.LastError != nil {
		if res.LastError != nil {
			return nil, errors.Wrap(res.LastError, "Stop: server in error")
		} else {
			return nil, errors.Errorf("Timeout: server does not reach stable state")
		}

	}
	return res.LastValue.(*api.Server), nil
}
