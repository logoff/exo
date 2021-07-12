package statefile

import (
	"context"

	"github.com/deref/exo/atom"
)

func New(filename string) *StateFile {
	return &StateFile{filename: filename}
}

type StateFile struct {
	filename string
}

type State struct {
	Components map[string]Component `json:"components"`
}

type Component struct {
	IRI string `json:"iri"`
}

func (sf *StateFile) swapState(f func(state *State) error) (*State, error) {
	var state State
	err := atom.SwapJSON(sf.filename, &state, func() error {
		return f(&state)
	})
	return &state, err
}

func (sf *StateFile) Remember(ctx context.Context, name string, iri string) error {
	_, err := sf.swapState(func(state *State) error {
		if state.Components == nil {
			state.Components = make(map[string]Component)
		}
		state.Components[name] = Component{
			IRI: iri,
		}
		return nil
	})
	return err
}
