package main

import (
	"log"
	"os"

	"github.com/hatim-lahwaouir/taskmaster/parser"
	"github.com/hatim-lahwaouir/taskmaster/supervisor"
)



func main(){
	if len(os.Args) != 2 {
		log.Fatal("Invalid args")		
	}

	parser := parser.NewYmlParser(os.Args[1])


	programs, err := parser.Start()
	if err != nil {
		log.Fatal(err)
	}

	s := supervisor.NewMasterSupervisor(programs)


	s.InitProcesses()
	s.Shell()
	s.Wait()
}