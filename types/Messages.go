package types


type ProcessStatus int
type ProcessTask  int


const (
    Start ProcessTask = iota
    Stop
    Reload 
    Wait 
    Status
)

const (
    Running ProcessStatus = iota
    Starting
    Stoped
)

var stateName = map[ProcessStatus]string{
    Running:      "Running",
    Stoped :      "Stoped",
    Starting:     "Starting",
}

func GetProcessStatus(status ProcessStatus) string {
    return stateName[status]
}

type Msg struct {
    Status ProcessStatus // process status returned from process running
    Task   ProcessTask  // task given from main thread to it's children
    ExitCode int
}


func New() Msg {
    return Msg{Task: Wait}
}

