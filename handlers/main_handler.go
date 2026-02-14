package handlers

import (
	"fmt"
	"github.com/hatim-lahwaouir/taskmaster/loggers"
	pm "github.com/hatim-lahwaouir/taskmaster/processMetadata"
	"github.com/hatim-lahwaouir/taskmaster/types"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type PHandler struct {
	Pm  pm.ProcessMetadata
	Id  int
	Msg chan types.Msg
}

var Loggers types.Loggers = loggers.ProgramLogs

func convert(prcs []pm.ProcessMetadata) map[int]PHandler {
	var (
		result map[int]PHandler
	)
	result = make(map[int]PHandler)

	for i, p := range prcs {
		result[i] = PHandler{Pm: p, Id: i + 1}
	}

	return result
}

func MainHandler(prcsMetadata []pm.ProcessMetadata) {
	var (
		prcs       map[int]PHandler
		wg         sync.WaitGroup
		MsgChannel chan types.Msg
		Cmd        chan string
        ProcessName map[string]bool
	)

	MsgChannel = make(chan types.Msg, 3)
    ProcessName = make(map[string]bool)
	Cmd = make(chan string)

    // start go routines that will handle processes
	prcs = convert(prcsMetadata)
	for id, p := range prcs {
		p.Msg = MsgChannel
		Loggers.InfoLogger.Printf("Starting %s:%d\n", p.Pm.ProcessName, id)

        // getting processNames
        ProcessName[strings.ToLower(p.Pm.ProcessName)] = true

		if p.Pm.Autostart {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ProcessHandler(p)
			}()
		}
	}
    // start go routine for handling interaction with cmd line 
    go CmdLine(Cmd, ProcessName)


    // listening for any input comming from command line
    for ;; {

        select {
            case cmd := <- Cmd:
                Loggers.InfoLogger.Printf("Cmd from user %s\n", cmd)
        }
    }
	wg.Wait()
}

func ProcessHandler(prc PHandler) {
	var (
		cmd  *exec.Cmd
		args []string
		wg   sync.WaitGroup
		err  error
	)
	//setting the cmd
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

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := cmd.Start(); err != nil {
			Loggers.ErrorLogger.Printf("%v\n", err)
			return
		}

		if err := cmd.Wait(); err != nil {
			if exiterr, ok := err.(*exec.ExitError); ok {
				Loggers.ErrorLogger.Printf("Exit Status: %d", exiterr.ExitCode())
				return
			} else {
				Loggers.ErrorLogger.Printf("%v\n", err)
				return
			}
		}
	}()
	wg.Wait()
}
