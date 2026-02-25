package handlers

import (
	"fmt"
	"github.com/hatim-lahwaouir/taskmaster/types"
	"github.com/hatim-lahwaouir/taskmaster/utils"
	"os"
	"os/exec"
	"strings"
	"sync"
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
		wg       sync.WaitGroup
		exitCode chan int
		processChan chan *os.Process
        process *os.Process
	)

	exitCode = make(chan int)
    processChan = make(chan *os.Process)
    defer close(processChan)
    defer close(exitCode)



	// the start time

    
	go startCmd(prc, &wg, exitCode, processChan)
    process =  <- processChan


	for {
		select {
		case msg := <-prc.Msg:
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
                case types.Stop:
                    sig, _ := utils.GetSignal(prc.Pm.StopSignal)
                    process.Signal(sig)

                    wg.Wait()
                    return
			}
        case exitStatus := <- exitCode:
            if MustRestart(prc, exitStatus) == true {
	            go startCmd(prc, &wg, exitCode, processChan)
            } else {
                prc.Mutex.Lock()
                prc.StartedAt = time.Date(0001, 1, 1, 00, 00, 00, 00, time.UTC)
                prc.Mutex.Unlock()
                return
            }
        }
	}
}


func startCmd(prc *PHandler, wg *sync.WaitGroup, exitCode chan int, process chan *os.Process) {

    var (
		cmd      *exec.Cmd
		args     []string
		err      error
	)

    prc.Mutex.Lock()
	prc.StartedAt = time.Now()
    prc.Mutex.Unlock()


	wg.Add(1)
    defer wg.Done()

    args = strings.Fields(prc.Pm.Cmd)
    cmd = exec.Command(args[0], args[1:]...)

    // setting env
    for k, v := range prc.Pm.Env {
        cmd.Env = append(cmd.Environ(), fmt.Sprintf("%v=%v", k, v))
    }
    // setting the output and stderr
    cmd.Stdout, err = os.OpenFile(prc.Pm.Stdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return
    }
    cmd.Stderr, err = os.OpenFile(prc.Pm.Stderr, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return
    }       

    if err := cmd.Start(); err != nil {
        Loggers.ErrorLogger.Printf("%v\n", err)
        return
    }
    process <- cmd.Process

    if err := cmd.Wait(); err != nil {
        if exiterr, ok := err.(*exec.ExitError); ok {
            exitCode <- exiterr.ExitCode()
            return
        } else {
            Loggers.ErrorLogger.Printf("%v\n", err)
            return
        }
    }
}


func MustRestart(p *PHandler, exitCode int) bool {

    

    if p.Pm.Autorestart == "never" {
        return false 
    }


    
    if p.Pm.Autorestart == "unexpected" {
        for _, ele := range  p.Pm.Exitcodes {
            if ele.(int) == exitCode {
                return false
            }
        }
    }



    p.Mutex.Lock()
    p.RestartRetries +=  1
    if p.RestartRetries >=   p.Pm.Startretries {
        p.Mutex.Unlock()
        return false 
    }
    p.Mutex.Unlock()

    return true 
}
