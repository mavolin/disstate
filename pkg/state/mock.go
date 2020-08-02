package state

import (
	"testing"

	"github.com/diamondburned/arikawa/state"
	"github.com/mavolin/dismock/pkg/dismock"
)

// NewMocker returns a dismock.Mocker and mocked version of the State.
func NewMocker(t *testing.T) (*dismock.Mocker, *State) {
	m, s := dismock.NewSession(t)
	return m, NewFromSession(s, new(state.NoopStore))
}

// CloneMocker clones the passed mocker and returns a new dismock.Mocker and
// mocked State.
func CloneMocker(m *dismock.Mocker, t *testing.T) (*dismock.Mocker, *State) {
	m, s := m.CloneSession(t)
	return m, NewFromSession(s, new(state.NoopStore))
}
