// Generated file. DO NOT EDIT.

package api

import (
	"context"
	"net/http"

	josh "github.com/deref/exo/internal/josh/server"
)

type Store interface {
	SetState(context.Context, *SetStateInput) (*SetStateOutput, error)
	GetState(context.Context, *GetStateInput) (*GetStateOutput, error)
}

type SetStateInput struct {
	ComponentID string            `json:"componentId"`
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	Tags        map[string]string `json:"tags"`
	Timestamp   string            `json:"timestamp"`
}

type SetStateOutput struct {
	Version int `json:"version"`
}

type GetStateInput struct {
	ComponentID string `json:"componentId"`
}

type GetStateOutput struct {
	State *State `json:"state"`
}

func BuildStoreMux(b *josh.MuxBuilder, factory func(req *http.Request) Store) {
	b.AddMethod("set-state", func(req *http.Request) interface{} {
		return factory(req).SetState
	})
	b.AddMethod("get-state", func(req *http.Request) interface{} {
		return factory(req).GetState
	})
}

type State struct {
	ComponentID string            `json:"componentId"`
	Version     int               `json:"version"`
	Type        string            `json:"type"`
	Content     string            `json:"content"`
	Tags        map[string]string `json:"tags"`
	Timestamp   string            `json:"timestamp"`
}
