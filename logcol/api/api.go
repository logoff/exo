// TODO: Generate this with JOSH tools.

package api

import (
	"context"
	"net/http"

	"github.com/deref/exo/josh"
)

type LogCollector interface {
	AddLog(context.Context, *AddLogInput) (*AddLogOutput, error)
	RemoveLog(context.Context, *RemoveLogInput) (*RemoveLogOutput, error)
	DescribeLogs(context.Context, *DescribeLogsInput) (*DescribeLogsOutput, error)
	GetEvents(context.Context, *GetEventsInput) (*GetEventsOutput, error)
	Collect(context.Context, *CollectInput) (*CollectOutput, error)
}

type AddLogInput struct {
	Name       string `json:"name"`
	SourcePath string `json:"sourcePath"`
}

type AddLogOutput struct{}

type RemoveLogInput struct {
	Name string `json:"name"`
}

type RemoveLogOutput struct{}

type DescribeLogsInput struct {
	Names []string `json:"names"`
}

type DescribeLogsOutput struct {
	Logs []LogDescription `json:"logs"`
}

type LogDescription struct {
	Name        string  `json:"name"`
	SourcePath  string  `json:"sourcePath"`
	LastEventAt *string `json:"lastEventAt"`
}

type GetEventsInput struct {
	LogNames []string `json:"logNames"`
	Before   string   `json:"before"`
	After    string   `json:"after"`
}

type GetEventsOutput struct {
	Events []Event `json:"events"`
}

type Event struct {
	LogName   string `json:"logName"`
	SID       string `json:"sid"`
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
}

type CollectInput struct{}

type CollectOutput struct{}

func NewLogCollectorMux(prefix string, collector LogCollector) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(prefix+"add-log", josh.NewMethodHandler(collector.AddLog))
	mux.Handle(prefix+"remove-log", josh.NewMethodHandler(collector.RemoveLog))
	mux.Handle(prefix+"describe-logs", josh.NewMethodHandler(collector.DescribeLogs))
	mux.Handle(prefix+"get-events", josh.NewMethodHandler(collector.GetEvents))
	mux.Handle(prefix+"collect", josh.NewMethodHandler(collector.Collect))
	return mux
}
