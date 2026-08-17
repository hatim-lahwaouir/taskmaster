package supervisor

import (
	"bytes"
	"fmt"
	"log"
	"maps"
	"net"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"syscall"
	"text/tabwriter"

	"github.com/hatim-lahwaouir/taskmaster/parser"
)




type  MasterSupervisor  struct {
	process map[string][]*ProcesseSupervisor
	wg sync.WaitGroup
	id 	   int
	shutdown atomic.Bool
	buffer bytes.Buffer 
	clients map[int]*Client
	request chan *Msg
}






func NewMasterSupervisor(p []*parser.Processes) *MasterSupervisor{
	process := make(map[string][]*ProcesseSupervisor)
	m := MasterSupervisor{process:  process, clients: make(map[int]*Client), request: make(chan *Msg, 10)}
	

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

	for i := range(v){
		fmt.Fprintf(&m.buffer," %v",v[i])

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

	w := tabwriter.NewWriter(&m.buffer, 0, 0, 2, ' ', 0)
	
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
	
	
}


func (m *MasterSupervisor) Shell() {
	m.wg.Add(1)
	defer m.wg.Done()

	cmdParser := parser.NewParseCmds(slices.Collect(maps.Keys(m.process)) )

	for ;; {

		
		if m.shutdown.Load(){
			return 
		}
		msg := <- m.request
		cmd, errParsing := cmdParser.ParseInput(msg.GetMesg()) 
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
		case "shutdown":
			m.Shutdown()
		}

		if m.buffer.Len() != 0 {
			m.clients[msg.GetClientID()].WriteGorotine(m.buffer.Bytes())
			m.buffer.Reset()
		}
	}
}


func (m *MasterSupervisor) Server() {
		// 1. Remove the old socket file if it exists to avoid "address already in use" errors.
	socketPath := "/tmp/taskmaster.sock"
	if err := os.RemoveAll(socketPath); err != nil {
		log.Fatalf("Failed to clean up socket file: %v", err)
	}

	// 2. Start the Unix domain socket listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", socketPath, err)
	}
	defer listener.Close()


	// 3. Keep accepting client connections indefinitely
	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
	
		if len(m.clients) >= 3 {
			newClient :=  NewClient(conn, m.request)
			newClient.WriteGorotine([]byte("we can't add more then 3 clients"))
			continue
		}

		newClient :=  NewClient(conn, m.request)
		// add NEw Client
		m.clients[newClient.ID()] = newClient
		go m.clients[newClient.ID()].ReadGorotine(&m.wg)
		// cleanning other connections 

		// deleting clients 
		for key := range(m.clients){
			if m.clients[key].IsBad(){
				delete(m.clients, key)
			}
		}
	}
}
	



