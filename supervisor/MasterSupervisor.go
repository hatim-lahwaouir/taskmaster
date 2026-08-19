package supervisor

import (
	"bufio"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/hatim-lahwaouir/taskmaster/parser"
)




type  MasterSupervisor  struct {
	process map[string][]*ProcesseSupervisor
	wg sync.WaitGroup
	id 	   int
	shutdown atomic.Bool
	clients map[int]*Client
	cmdParser *parser.ParseCmds
	mu sync.Mutex
	LoadingConfig atomic.Bool
}






func NewMasterSupervisor(p []*parser.Processes) *MasterSupervisor{
	process := make(map[string][]*ProcesseSupervisor)
	m := MasterSupervisor{process:  process, clients: make(map[int]*Client), }
	

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


func (m *MasterSupervisor) KillProcess(processName string) {

		for i := range(m.process[processName]){
			m.process[processName][i].KillProcess()
		}
}






func (m *MasterSupervisor) logs(v ...any) {


	fmt.Fprintf(os.Stdout,">")
	for i := range(v){
		fmt.Fprintf(os.Stdout,"%v ",v[i])
	}
	fmt.Println()
}


func (m *MasterSupervisor) LoadConfig(p []*parser.Processes) {
	m.LoadingConfig.Store(true)
	defer m.LoadingConfig.Store(false)

	var (
		mp map[string]bool
	)
	mp = make(map[string]bool)

	m.mu.Lock()
	defer m.mu.Unlock()
	// remove process that no longer exitsts
	for prcs := range(p){
		mp[p[prcs].Name] = true
	}

	for k := range m.process{
		if mp[k] == true {
			continue
		}
		m.KillProcess(k)
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
				// kill all these processes and add new one
				m.KillProcess(k)
				delete(m.process, k)
				for j := 0; j < p[i].NumProcs; j++ {
					m.AddProcess(p[i])
				}
			}
		} else {
			m.AddProcess(p[i])
		}
	}

	m.cmdParser = parser.NewParseCmds(slices.Collect(maps.Keys(m.process)) )
	m.InitProcesses()
}


func (m *MasterSupervisor) InitProcesses() {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, " Name\t| Status\t")
	fmt.Fprintln(w, " ----\t| ------\t")
	
	for i := range(m.process){
		for j := range(m.process[i]){
			p := m.process[i][j]
			if p.AutoStart && p.Stoped.Load() && !p.shutdowning.Load(){  
				m.wg.Add(1)
				fmt.Fprintf(w," %s\t| is starting\t\n", p)
				go m.process[i][j].Start()
			}else if p.AutoStart && p.shutdowning.Load() {
				fmt.Fprintf(w," %s\t| can't start still shutdowning\t\n", p)
			}
		}
	}

	w.Flush()
}


func (m *MasterSupervisor) Start(processName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintln(w, " Name\t| Status\t")
	fmt.Fprintln(w, " ----\t| ------\t")
	
	for j := range(m.process[processName]){
		p := m.process[processName][j]

		if p.shutdowning.Load() == true {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "is shutdowning")
		}else if p.Stoped.Load() == true {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "starting")
			m.wg.Add(1)
			go m.process[processName][j].Start()
		}else {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "already started")
		}
	}
	w.Flush()
}


func (m *MasterSupervisor) Restart(processName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
	fmt.Fprintln(w, " Name\t| Status\t")
	fmt.Fprintln(w, " ----\t| ------\t")
	
	for j := range(m.process[processName]){
		p := m.process[processName][j]
		if p.Stoped.Load() == true {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "can't restart a stoped process")
		}else if p.shutdowning.Load() == true {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "can't restart a shutdowning process")
		} else {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "restaring process")
			m.process[processName][j].KillProcess()
			m.wg.Add(1)
			go m.process[processName][j].Start()
		}
	}
	w.Flush()
}




func (m *MasterSupervisor) Stop(processName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	
		fmt.Fprintln(w, " Name\t| Status\t")
		fmt.Fprintln(w, " ----\t| ------\t")

	for j := range(m.process[processName]){
		p := m.process[processName][j]

		if p.shutdowning.Load(){
			fmt.Fprintf(w," %s\t| %s\t\n", p, "in the process of stoping")
		}else if p.Stoped.Load() == false {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "stoping")
			m.wg.Add(1)
			go m.process[processName][j].Stop()
		}else {
			fmt.Fprintf(w," %s\t| %s\t\n", p, "already stoped")
		}
	}

	w.Flush()
}


func (m *MasterSupervisor) Status(processName string) {

	m.mu.Lock()
	defer m.mu.Unlock()

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



func (m *MasterSupervisor) Shutdown() {
	m.shutdown.Store(true)
	for processName := range m.process{
		// kill all processes
		m.KillProcess(processName)
	}

	for c := range(m.clients){
		m.clients[c].conn.Close()
	}


}


func (m *MasterSupervisor) Shell() {
	m.wg.Add(1)
	defer m.wg.Done()

	m.cmdParser = parser.NewParseCmds(slices.Collect(maps.Keys(m.process)) )
	scanner := bufio.NewScanner(os.Stdin)
	for ;; {

		

		if m.shutdown.Load(){
			return 
		}
		

		fmt.Print("> ") // Optional prompt indicator

		// Blocks until a new line is entered or EOF is reached (e.g., Ctrl+D)
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
			}
			// EOF or read error; exit loop gracefully
			return
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}


		m.mu.Lock()
		cmd, errParsing := m.cmdParser.ParseInput(input) 
		m.mu.Unlock()
		
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
		case "restart":
			m.Restart(cmd[1])
		case "load":
			m.LoadingConfig.Store(true)
			m.Load()
			// waiting for config to be loaded 
			for m.LoadingConfig.Load() {
				time.Sleep(50 * time.Millisecond)
			}
		case "help":
			m.cmdParser.Help()
		case "shutdown":
			m.Shutdown()
		}
	}
}



	



