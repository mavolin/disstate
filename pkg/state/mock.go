package state

import (
	"testing"

	"github.com/diamondburned/arikawa/v2/state/store"
	"github.com/mavolin/dismock/v2/pkg/dismock"
)

// NewMocker returns a dismock.Mocker and mocked version of the State.
func NewMocker(t *testing.T) (*dismock.Mocker, *State) {
	m, s := dismock.NewSession(t)
	return m, NewFromSession(s, store.NoopCabinet)
}

// CloneMocker clones the passed dismock.Mocker and returns a new mocker and
// mocked State.
func CloneMocker(m *dismock.Mocker, t *testing.T) (*dismock.Mocker, *State) {
	m, s := m.CloneSession(t)
	return m, NewFromSession(s, store.NoopCabinet)
}
