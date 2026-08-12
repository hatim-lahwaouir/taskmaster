package supervisor

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"slices"
	"sync"
	"text/tabwriter"
	"time"

	"github.com/hatim-lahwaouir/taskmaster/errorshandling"
	"github.com/hatim-lahwaouir/taskmaster/parser"
)




type  MasterSupervisor  struct {
	process map[string][]*ProcesseSupervisor
	wg sync.WaitGroup
}






func NewMasterSupervisor(p []*parser.Processes) *MasterSupervisor{
	process := make(map[string][]*ProcesseSupervisor)
	m := MasterSupervisor{process:  process}

	for i := range(p){

		for j := 0; j < p[i].NumProcs; j++{
			process[p[i].Name] = append(process[p[i].Name], NewProcesseSupervisor(p[i], &m.wg))
		}
	}


	return &m
}



func (m *MasterSupervisor) Print() {

	for i := range(m.process){
		for j := range(m.process[i]){
			fmt.Printf("%v\n", m.process[i][j])
		}
	}
}


func (m *MasterSupervisor) logs(v any) {
	fmt.Printf("taskmaster> %v\n", v)
}


func (m *MasterSupervisor) InitProcesses() {

	for i := range(m.process){
		for j := range(m.process[i]){
			p := m.process[i][j]
			if p.AutoStart {  
				m.wg.Add(1)
				go m.process[i][j].Start()
			}
		}
	}
}


func (m *MasterSupervisor) Start(processName string) {

	for j := range(m.process[processName]){
		p := m.process[processName][j]
		if p.Stoped.Load() == true {
			m.logs("starting " + processName)
			m.wg.Add(1)
			go m.process[processName][j].Start()
		}else {
			m.logs("already started " + processName)
		}
	}
}


func (m *MasterSupervisor) Stop(processName string) {

	for j := range(m.process[processName]){
		p := m.process[processName][j]
		if p.Stoped.Load() == false {
			m.logs("stoping " + processName)
			m.process[processName][j].Stop()
		}else {

			m.logs("Already stoped" + processName)
			// fmt.Println(, processName)
		}
	}
}


func (m *MasterSupervisor) Status(processName string) {

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintln(w, " Name\t| Duration\t| Status\t| PID\t")
	fmt.Fprintln(w, " ----\t| --------\t| ------\t| ---\t")
	for j := range(m.process[processName]){
		m.process[processName][j].Status(w)
	
	}

	w.Flush()

}

func (m *MasterSupervisor) Wait() {
	m.wg.Wait()
}




func (m *MasterSupervisor) Shell() {
	time.Sleep(100 * time.Millisecond)

	cmdParser := parser.NewParseCmds(slices.Collect(maps.Keys(m.process)) )


	for ;; {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("taskmaster> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println()
			errorshandling.NewErrorReporter(errorshandling.ErrTaskMaster, err.Error() + " can't read stdin").Report()
			continue
		}
		cmd, errParsing := cmdParser.ParseInput(input) 
		if errParsing != nil {
			errParsing.Report()
			continue
		}


		switch cmd[0] {
		case "status":
			m.Status(cmd[1])
		case "stop":
			m.Stop(cmd[1])
		case "start":
			m.Start(cmd[1])
		case "load":
			// m.Start(cmd[1])
		case "help":
			cmdParser.Help()
		}

	}
}



