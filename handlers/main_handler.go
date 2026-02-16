package handlers

import (
	"github.com/hatim-lahwaouir/taskmaster/loggers"
	pm "github.com/hatim-lahwaouir/taskmaster/processMetadata"
	"github.com/hatim-lahwaouir/taskmaster/types"
	"strings"
	"sync"
    "time"
)

type PHandler struct {
	Pm  pm.ProcessMetadata
	Id  int
	Msg chan types.Msg
    StartedAt time.Time 
}

var Loggers types.Loggers = loggers.ProgramLogs

func convert(prcs []pm.ProcessMetadata) map[int]*PHandler {
	var (
		result map[int]*PHandler
	)
	result = make(map[int]*PHandler)

	for i, p := range prcs {
		result[i + 1] = &PHandler{Pm: p, Id: i + 1 }
	}

	return result
}

func MainHandler(prcsMetadata []pm.ProcessMetadata) {
	var (
		prcs       map[int]*PHandler
		wg         sync.WaitGroup
		Cmd        chan string
		Info       chan string
        ProcessName map[string][]int
	)

    ProcessName = make(map[string][]int)
	Cmd = make(chan string, 2)
	Info = make(chan string, 20)

    // start go routines that will handle processes
	prcs = convert(prcsMetadata)
	for id, p := range prcs {
		Loggers.InfoLogger.Printf("Starting %s:%d\n", p.Pm.ProcessName, id)

        // getting processNames
        ProcessName[strings.ToLower(p.Pm.ProcessName)] = append(ProcessName[strings.ToLower(p.Pm.ProcessName)], id)
        prcs[id].Msg = make(chan types.Msg, 50)
        
        

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
        CmdLine(Cmd, Info , ProcessName)
    }()


    // listening for any input comming from command line
    for ;; {

        select {
            case cmd := <- Cmd:
                handelCmd(cmd,Info, prcs, ProcessName)

        }
    }
	wg.Wait()
}


func handelCmd(cmd string, info chan  string , prcs map[int]*PHandler, prscName map[string][]int) {

    var (
        arg []string
        name string
        todo string
        result string
        reply chan string
    )
    arg = strings.Split(cmd , " ")
    todo = strings.ToLower(arg[0])
    name = strings.ToLower(arg[1])
    reply = make(chan string, 10)



    switch todo {
        case "start":
            Loggers.InfoLogger.Printf("%s Targeting \n", todo)
        case "stop":
            Loggers.InfoLogger.Printf("%s Targeting \n", todo)
        case "status":
            
            for _, id := range(prscName[name]) {
                prcs[id].Msg <- types.Msg{Task: types.Task["Status"], RespMsg: reply }
            }

            for _, _ = range(prscName[name]) {
                msg := <- reply 
                result = result + msg
            }
            close(reply)

            info <- result
    }

}





