package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
    "time"
	"github.com/hatim-lahwaouir/taskmaster/types"
)

func ProcessHandler(prc *PHandler) {
	var (
		cmd  *exec.Cmd
		args []string
		err  error
        exitCode chan int 

	)
	//setting the cmd
    exitCode = make(chan int)
	args = strings.Split(prc.Pm.Cmd, " ")
	cmd = exec.Command(args[0], args[1:]...)

	// setting env
	for k, v := range prc.Pm.Env {
		cmd.Env = append(cmd.Environ(), fmt.Sprintf("%v=%v", k, v))
	}
	// setting the output and stderr
	cmd.Stdout, err = os.OpenFile(prc.Pm.Stdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Loggers.ErrorLogger.Printf("Opening the stdout log file %v\n", err)
		return
	}
	cmd.Stderr, err = os.OpenFile(prc.Pm.Stderr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Loggers.ErrorLogger.Printf("Opening the stdout log file %v\n", err)
		return
	}

    // the start time
    prc.StartedAt = time.Now()
	go func() {
		if err := cmd.Start(); err != nil {
			Loggers.ErrorLogger.Printf("%v\n", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				Loggers.ErrorLogger.Printf("Exit Status: %d", exiterr.ExitCode())
                exitCode <- exiterr.ExitCode()
				return
			} else {
				Loggers.ErrorLogger.Printf("%v\n", err)
				return
			}
		}
	}()

    for ;; {
        select  {
            case exitStatus := <- exitCode:
				Loggers.ErrorLogger.Printf("Exit Status: %d", exitStatus)
            case msg := <- prc.Msg: 
                // we need at switch statement for checking what user is asking for 

                switch msg.Task {
                    case types.Status:
                         result := fmt.Sprintf("%s:%d  UP  %v\n",prc.Pm.ProcessName, prc.Id, time.Since(prc.StartedAt))
			             msg.RespMsg <- result 	
                }
          }
    }
}
