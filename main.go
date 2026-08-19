package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/hatim-lahwaouir/taskmaster/parser"
	signalhandler "github.com/hatim-lahwaouir/taskmaster/signalHandler"
	"github.com/hatim-lahwaouir/taskmaster/supervisor"
)





func main(){
	var (
		wg sync.WaitGroup
	)

	if len(os.Args) != 2 {
		log.Fatal("Invalid args")		
	}

	fmt.Println(os.Getpid())
	parser := parser.NewYmlParser(os.Args[1])


	programs, err := parser.Start()
	if err != nil {
		log.Fatal(err)
	}

	master := supervisor.NewMasterSupervisor(programs)
	
	signalHandler := signalhandler.NewRealoadConfigSig(master, parser)

	wg.Add(1)
	go signalHandler.Handler(&wg)

	master.InitProcesses()
	go master.Shell()
	master.Wait()



	signalHandler.Stop()
	signalHandler.Clear()
	wg.Wait()
}