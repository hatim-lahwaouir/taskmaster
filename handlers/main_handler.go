package handlers

import (
	"fmt"
	"github.com/hatim-lahwaouir/taskmaster/loggers"
	pm "github.com/hatim-lahwaouir/taskmaster/processMetadata"
	"github.com/hatim-lahwaouir/taskmaster/types"
	"github.com/hatim-lahwaouir/taskmaster/utils"
	"strconv"
	"strings"
	"sync"
	"time"
)

type PHandler struct {
	Pm             pm.ProcessMetadata
	Id             int
	Msg            chan types.Req
	StartedAt      time.Time
	RestartRetries int64
	Mutex          sync.Mutex
}

var Loggers types.Loggers = loggers.ProgramLogs

func convert(prcs []pm.ProcessMetadata) map[int]*PHandler {
	var (
		result map[int]*PHandler
	)
	result = make(map[int]*PHandler)

	for i, p := range prcs {
		result[i+1] = &PHandler{Pm: p, Id: i + 1}
	}

	return result
}

func MainHandler(prcsMetadata []pm.ProcessMetadata) {
	var (
		prcs        map[int]*PHandler
		wg          sync.WaitGroup
		Cmd         chan string
		Info        chan string
		ProcessName map[string][]int
	)

	ProcessName = make(map[string][]int)
	Cmd = make(chan string, 5)
	Info = make(chan string, 20)

	// start go routines that will handle processes
	prcs = convert(prcsMetadata)
	for id, p := range prcs {
		Loggers.InfoLogger.Printf("Starting %s:%d\n", p.Pm.ProcessName, id)

		// getting processNames
		ProcessName[strings.ToLower(p.Pm.ProcessName)] = append(ProcessName[strings.ToLower(p.Pm.ProcessName)], id)
		prcs[id].Msg = make(chan types.Req, 3)

		if p.Pm.Autostart {
			wg.Add(1)
			go func() {
				defer wg.Done()

				ProcessHandler(p)
			}()
		}
	}
	// start go routine for handling interaction with cmd line
	go func() {
		wg.Add(1)
		CmdLine(Cmd, Info, ProcessName)
	}()

	// listening for any input comming from command line
	for {

		select {
		case cmd := <-Cmd:
			handelCmd(cmd, Info, prcs, ProcessName, &wg)

		}
	}
	wg.Wait()
}

func getHeaderStatus() string {
	return fmt.Sprintf("%-20s %-10s %-25s %-25s\n",
		"ProcessName:Id",
		"Status",
		"UpDuration",
		"StartRetries")
}

func getHeaderStart() string {
	return fmt.Sprintf("%-20s %-20s\n",
		"ProcessName:Id",
		"Status",
	)
}

func getHeaderStop() string {
	return fmt.Sprintf("%-20s %-20s\n",
		"ProcessName:Id",
		"Status",
	)
}


func handleRespStatus(msg types.Resp, prcs *PHandler) string {

    var (
        ret string
    )

    if msg.Status == "Running" ||   msg.Status == "Starting" { 
        prcs.Mutex.Lock()
        ret =  fmt.Sprintf("%-20s %-10s %-25v %-25s\n",
            utils.Truncate(msg.PrcsName+":"+strconv.Itoa(msg.Id), 17),
            msg.Status,
            msg.UpDuration.Round(time.Second),
            utils.Truncate(strconv.FormatInt(msg.RestartRetries, 10)+"/"+strconv.FormatInt(prcs.Pm.Startretries, 10), 25-3))
        prcs.Mutex.Unlock()
    } else { 
        ret = fmt.Sprintf("%-20s %-10s %-25v %-25s\n",
            utils.Truncate(msg.PrcsName+":"+strconv.Itoa(msg.Id), 17),
            msg.Status,
            "NA",
            "NA")
    }

    return ret
}



func handleRespStart(prcs *PHandler) string {
	var (
		status string
	)
	prcs.Mutex.Lock()
	status = "running"
	if prcs.StartedAt.IsZero() {
		status = "starting"
	}

	prcs.Mutex.Unlock()
	return fmt.Sprintf("%-20s %s\n",
		utils.Truncate(prcs.Pm.ProcessName+":"+strconv.Itoa(prcs.Id), 17), status)
}


func handleRespStop(prcs *PHandler, running bool) string {
    if running {
        return fmt.Sprintf("%-20s %s\n",
		        utils.Truncate(prcs.Pm.ProcessName+":"+strconv.Itoa(prcs.Id), 17), "stoping process")
    }
        

   return fmt.Sprintf("%-20s %s\n",
		        utils.Truncate(prcs.Pm.ProcessName+":"+strconv.Itoa(prcs.Id), 17), "already stoped")

}

func startingProcess(wg *sync.WaitGroup, p *PHandler) {
    p.Mutex.Lock()
	if  p.StartedAt.IsZero() == false {

	       p.Mutex.Unlock()
           return
	}

    p.RestartRetries = 0 
	p.Mutex.Unlock()

	wg.Add(1)
	go func() {
		defer wg.Done()

		p.Msg = make(chan types.Req, 3)
		ProcessHandler(p)
	}()

}




func handlStart(info chan string, prcs map[int]*PHandler, prscName map[string][]int, name string, wg *sync.WaitGroup) {
    var (
        result string
    )
    // check if process already running
    result = getHeaderStart()
    for _, id := range prscName[name] {
        result = result + handleRespStart(prcs[id])
    }
    // we need to start process if they never started
    for _, id := range prscName[name] {
        startingProcess(wg,prcs[id])
    }
    info <- result
}


func handlStatus(info chan string, prcs map[int]*PHandler, prscName map[string][]int, name string) {
    var (
        result string
		resp   chan types.Resp
        prunning bool
    )
    resp = make(chan types.Resp, 10)
    defer close(resp)


    result = getHeaderStatus()
    for _, id := range prscName[name] {
        prunning = true
        prcs[id].Mutex.Lock()
	    if prcs[id].StartedAt.IsZero() {
            prunning = false
	    }
        prcs[id].Mutex.Unlock()



        if ! prunning  {
            result = result + handleRespStatus(types.Resp{Id: prcs[id].Id, PrcsName: prcs[id].Pm.ProcessName,
            Status:	types.GetProcessStatus(types.Stoped)}, prcs[id])
        } else {
            prcs[id].Msg <- types.Req{Task: types.Task["Status"], RespMsg: resp}
            result = result + handleRespStatus(<- resp, prcs[id])
        }
        
    }
    info <- result

}


func handlStop(info chan string, prcs map[int]*PHandler, prscName map[string][]int, name string) { 
    var (
        result string
        prunning bool
    )

    for _, id := range prscName[name] {
        prunning = true
        prcs[id].Mutex.Lock()
	    if prcs[id].StartedAt.IsZero() {
            prunning = false
	    }
        prcs[id].Mutex.Unlock()



        if ! prunning  {
            result = result + handleRespStop(prcs[id], prunning)
        } else {
            prcs[id].Msg <- types.Req{Task: types.Task["Stop"]}
            result = result + handleRespStop(prcs[id], prunning)
        }
    }


    info <- result
    for _, id := range prscName[name] {
        prcs[id].Mutex.Lock()
        prcs[id].StartedAt = time.Date(0001, 1, 1, 00, 00, 00, 00, time.UTC)
        prcs[id].Mutex.Unlock()

        close(prcs[id].Msg) 
    }
}





func handelCmd(cmd string, info chan string, prcs map[int]*PHandler, prscName map[string][]int, wg *sync.WaitGroup) {

	var (
		arg    []string
		name   string
		todo   string
	)
	arg = strings.Fields(cmd )
	todo = strings.ToLower(arg[0])
	name = strings.ToLower(arg[1])

	switch todo {
	case "start":
        handlStart(info, prcs, prscName, name, wg)
	case "stop":
        handlStop(info, prcs, prscName, name)
	case "status":
        handlStatus(info, prcs, prscName, name)
	}
}
