package handlers

import (
    pm "github.com/hatim-lahwaouir/taskmaster/processMetadata" 
    "github.com/hatim-lahwaouir/taskmaster/loggers"
    "github.com/hatim-lahwaouir/taskmaster/types"
    "fmt"
    "strings"
    "os/exec"
    "sync"
    "os"
)


type PHandler struct {
    Pm pm.ProcessMetadata
    Msg chan types.Msg
    Id int
}


var  Loggers types.Loggers  =  loggers.ProgramLogs



func convert(prcs []pm.ProcessMetadata) []PHandler {
    var (
        result []PHandler
        i int
    )

    i = 1 

    for _, p := range(prcs) {
        result = append(result, PHandler{Pm :p, Msg: make(chan types.Msg, 3), Id: i}) 
        i += 1
    }

    return result
}

func MainHandler(prcsMetadata []pm.ProcessMetadata) {
    var (
        prcs []PHandler
        wg sync.WaitGroup
    )

    prcs = convert(prcsMetadata)
    for _, p := range(prcs) {
        Loggers.InfoLogger.Printf("Starting %s:%d\n", p.Pm.ProcessName, p.Id)
        wg.Add(1)
        go func(){
               defer wg.Done()
               ProcessHandler(p)
        }()
    }
    wg.Wait()
}


func ProcessHandler(prc PHandler) {
        var (
            cmd *exec.Cmd
            args []string
            wg sync.WaitGroup
            err error
        )
        //setting the cmd 
        args = strings.Split(prc.Pm.Cmd, " ")
        cmd = exec.Command(args[0], args[1:]...)

        // setting env
        for k, v := range(prc.Pm.Env) {
            cmd.Env = append(cmd.Environ(), fmt.Sprintf("%v=%v",k, v))
        }
        // setting the output and stderr
        cmd.Stdout, err =  os.OpenFile(prc.Pm.Stdout, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666) 
        if err != nil {
                Loggers.ErrorLogger.Printf("Opening the stdout log file %v\n", err)
                return
        }
        cmd.Stderr ,err =  os.OpenFile(prc.Pm.Stderr,os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if err != nil {
                Loggers.ErrorLogger.Printf("Opening the stdout log file %v\n", err)
                return
        }




        wg.Add(1)
        go func()  {
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





