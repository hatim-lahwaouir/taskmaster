package types

import (
	"time"
)

type ProcessStatus int
type ProcessTask int

const (
	Start ProcessTask = iota
	Stop
	Reload
	Wait
	Status
    IsRunning
)

const (
	Running ProcessStatus = iota
	Starting
	Stoped
)

var stateName = map[ProcessStatus]string{
	Running:  "Running",
	Stoped:   "Stoped",
	Starting: "Starting",
}

var Task = map[string]ProcessTask{
	"Start":  Start,
	"Stop":   Stop,
	"Status": Status,
	"IsRunning": IsRunning,
}

var StatusResp = map[string]ProcessStatus{
	"Running":  Running,
	"Stoped":   Stoped,
	"Starting": Starting,
}

func GetProcessStatus(status ProcessStatus) string {
	return stateName[status]
}

type Resp struct {
	Id         int
	PrcsName   string
	Status     string // process status returned from process running
	ExitCode   int
	UpDuration time.Duration
    RestartRetries int64
}

type Req struct {
	Task    ProcessTask // task given from main thread to it's children
	RespMsg chan Resp
}

func New() Req {
	return Req{Task: Wait}
}
