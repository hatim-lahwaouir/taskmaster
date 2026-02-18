package handlers

import (
	"fmt"
	"github.com/hatim-lahwaouir/taskmaster/types"
	"os"
	"os/exec"
	"strings"
	"time"
)

func statusState(dur time.Duration, startTime int64) string {
	var (
		t int64
	)

	t = int64(dur.Seconds())

	if t < startTime {
		return types.GetProcessStatus(types.Starting)
	}
	return types.GetProcessStatus(types.Running)
}

func ProcessHandler(prc *PHandler) {
	var (
		cmd      *exec.Cmd
		args     []string
		err      error
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

	for {
		select {
		case exitStatus := <-exitCode:
			Loggers.ErrorLogger.Printf("Exit Status: %d", exitStatus)
		case msg := <-prc.Msg:
			// we need at switch statement for checking what user is asking for

			switch msg.Task {
			case types.Status:
				msg.RespMsg <- types.Resp{
					Id:         prc.Id,
					PrcsName:   prc.Pm.ProcessName,
					UpDuration: time.Since(prc.StartedAt),
					ExitCode:   -1,
					Status: statusState(time.Since(prc.StartedAt),
						prc.Pm.Starttime),
					RestartRetries: prc.RestartRetries}
			}
		}
	}
}
