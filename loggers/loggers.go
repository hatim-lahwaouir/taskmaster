package loggers 


import (
    "github.com/hatim-lahwaouir/taskmaster/types"
    "os"
    "log"
)



var  ProgramLogs  types.Loggers


func init() {
    // using this function for setting up loggers
    ProgramLogs.ErrorLogger =  log.New(os.Stdout, "taskmaster> Error: ", 0)
    ProgramLogs.InfoLogger  =  log.New(os.Stdout, "taskmaster> Info: ", 0)
}
