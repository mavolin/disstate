package state

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state/store"
	"github.com/mavolin/dismock/v3/pkg/dismock"
)

// NewMocker returns a dismock.Mocker and mocked version of the State.
func NewMocker(t *testing.T) (*dismock.Mocker, *State) {
	m := dismock.New(t)

	s, err := New(Options{
		Token:      "a totally valid token",
		Cabinet:    store.NoopCabinet,
		HTTPClient: m.HTTPClient(),
		Gateways:   []*gateway.Gateway{gateway.NewCustomGateway("", "")},
	})
	if err != nil {
		panic("state: NewMocker " + err.Error())
	}

	return m, s
}

// CloneMocker clones the passed dismock.Mocker and returns a new mocker and
// mocked State.
func CloneMocker(m *dismock.Mocker, t *testing.T) (*dismock.Mocker, *State) {
	m = m.Clone(t)

	s, err := New(Options{
		Token:      "trust me, I'm real",
		Cabinet:    store.NoopCabinet,
		HTTPClient: m.HTTPClient(),
		Gateways:   []*gateway.Gateway{gateway.NewCustomGateway("", "")},
	})
	if err != nil {
		panic("state: NewMocker " + err.Error())
	}

	return m, s
}
