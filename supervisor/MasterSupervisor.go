package supervisor

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"slices"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/hatim-lahwaouir/taskmaster/errorshandling"
	"github.com/hatim-lahwaouir/taskmaster/parser"
)




type  MasterSupervisor  struct {
	process map[string][]*ProcesseSupervisor
	wg sync.WaitGroup
	id 	   int
}






func NewMasterSupervisor(p []*parser.Processes) *MasterSupervisor{
	process := make(map[string][]*ProcesseSupervisor)
	m := MasterSupervisor{process:  process}
	

	for i := range(p){
		for j := 0; j < p[i].NumProcs; j++{
			prcs := NewProcesseSupervisor(p[i], &m.wg)
			prcs.Id = m.id
			m.id++
			process[p[i].Name] = append(process[p[i].Name],prcs )
		}
	}


	return &m
}


func (m *MasterSupervisor) AddProcess(p *parser.Processes) {

	prcs := NewProcesseSupervisor(p, &m.wg)
	prcs.Id = m.id
	m.id++
	m.process[p.Name] = append(m.process[p.Name],prcs)
}


func (m *MasterSupervisor) Print() {

	for i := range(m.process){
		for j := range(m.process[i]){
			fmt.Printf("%v\n", m.process[i][j])
		}
	}
}


func (m *MasterSupervisor) logs(v ...any) {

	fmt.Printf("taskmaster>")

	for i := range(v){
		fmt.Printf(" %v",v[i])
	}
	fmt.Println()
}


func (m *MasterSupervisor) LoadConfig(p []*parser.Processes) {
	fmt.Println("loading cofing", len(p))
	var (
		mp map[string]bool
	)
	mp = make(map[string]bool)

	// remove process that no longer exitsts
	for prcs := range(p){
		mp[p[prcs].Name] = true
	}

	for k := range m.process{
		if mp[k] == true {
			continue
		}
		fmt.Println("killing process ", k)
		for i := range(m.process[k]){
			m.process[k][i].KillProcess()
		}
		delete(m.process, k)
	}

	// add new processes 
	// update already existed processes


	for i := range p{
		k := p[i].Name
		_, ok := m.process[k]
		if mp[k] == true && ok{
			// here we talk about updating 
			if m.process[k][0].CanbeUpdated(p[i]){
				// update all the process
				for p[i].NumProcs > len(m.process[k]){
					fmt.Println("adding new Process to", k)
					m.AddProcess(p[i])
				}
				for p[i].NumProcs < len(m.process[k]){
					m.process[k][len(m.process[k])-1].KillProcess() 
					m.process[k] = m.process[k][:len(m.process[k])-1]
				}

				
				for j := range m.process[k]{
					m.process[k][j].Update(p[i])
				}
			}else {
				fmt.Println("can't be updated", k)
				
			}
		} else {
			m.AddProcess(p[i])
			fmt.Println("we need to Add", k)
		}
	}


	m.InitProcesses()
}


func (m *MasterSupervisor) InitProcesses() {

	for i := range(m.process){
		for j := range(m.process[i]){
			p := m.process[i][j]
			if p.AutoStart && p.Stoped.Load(){  
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
			m.logs("starting" , p)
			m.wg.Add(1)
			go m.process[processName][j].Start()
		}else {
			m.logs("already started" , p)
		}
	}
}


func (m *MasterSupervisor) Stop(processName string) {

	for j := range(m.process[processName]){
		p := m.process[processName][j]
		if p.Stoped.Load() == false {
			m.logs("stoping" ,m.process[processName][j])
			m.wg.Add(1)
			go m.process[processName][j].Stop()
		}else {

			m.logs("Already stoped",m.process[processName][j])
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


func (m *MasterSupervisor) Load() {
    pid := os.Getpid()
    proc, err := os.FindProcess(pid)
    if err != nil {
        panic(err)
    }

    err = proc.Signal(syscall.SIGHUP)
    if err != nil {
		m.logs("error sending sighup")
		return 
	}
}




func (m *MasterSupervisor) Wait() {
	m.wg.Wait()
}




func (m *MasterSupervisor) Shell() {
	

	cmdParser := parser.NewParseCmds(slices.Collect(maps.Keys(m.process)) )


	for ;; {
		reader := bufio.NewReader(os.Stdin)
		time.Sleep(100 * time.Millisecond)
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
			m.Load()
		case "help":
			cmdParser.Help()
		}

	}
}



