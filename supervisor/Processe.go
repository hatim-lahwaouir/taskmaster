package supervisor

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/hatim-lahwaouir/taskmaster/errorshandling"
	"github.com/hatim-lahwaouir/taskmaster/parser"
)




type ProcesseSupervisor struct {

	Id int
	Name string
	Cmd string 		
	ExecCmd *exec.Cmd									
	Umask string 					
	WorkingDIr string 				
	AutoStart bool 					
	AutoRestart string 					
	StartRetries int 					
	StartTime int64  					
	Stdout string 						
	Stderr string
	stdoutFile *os.File
    stderrFile *os.File			
	ExitCodes map[int]bool							
	Env map[string]string
	StopSingal os.Signal
	StopTime int64
	// to check status 
	StartedAt time.Time

	// context for canceling the cmd
	
	// started 



	// to check if cmd stoped 
	Stoped atomic.Bool

	// waiting for clearence
	wg *sync.WaitGroup

	// mutext 

	mu sync.RWMutex

	cuRetry int
}



func NewProcesseSupervisor(p *parser.Processes, wg *sync.WaitGroup) *ProcesseSupervisor {


	exitCodes := make(map[int]bool)


	if exitcodes, ok := p.ExitCodes.([]int); ok{
		for i := range(exitcodes){
			exitCodes[exitcodes[i]] = true
		}
	}

	if exitcodes, ok := p.ExitCodes.(int); ok{
			exitCodes[exitcodes] = true
	}

	process := ProcesseSupervisor{
		Name:         p.Name,
		Cmd:          p.Cmd,
		Umask:        fmt.Sprintf("%03o", p.Umask),
		WorkingDIr:   p.WorkingDIr,
		AutoStart:    p.AutoStart,
		AutoRestart:  p.AutoRestart,
		StartRetries: p.StartRetries,
		StartTime:    p.StartTime,
		Stdout:       p.Stdout,
		Stderr:       p.Stderr,
		ExitCodes:    exitCodes, 
		Env:          p.Env,
		StopTime: p.StopTime,
		StopSingal: p.GetSingal(),
		wg: wg,
	}

	process.Stoped.Store(true)
	return &process
}

func (p *ProcesseSupervisor) Status(w *tabwriter.Writer) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.Stoped.Load() {
		fmt.Fprintf(w," %s\t| -1\t| stoped\t| -1\t\n", p)
		return
	}
	duration := time.Since(p.StartedAt).Round(time.Millisecond)

	if p.StartedAt.Add(time.Duration(p.StartTime * int64(time.Second))).Compare(time.Now()) > 0{
		fmt.Fprintf(w," %s\t| %s\t| started\t| %d\t\n", p, duration, p.ExecCmd.Process.Pid)
		return 
	}

	fmt.Fprintf(w," %s\t| %s\t| running\t| %d\t\n", p, duration, p.ExecCmd.Process.Pid)
}

func (p *ProcesseSupervisor) Stop() {

	defer p.wg.Done()
	
	p.mu.Lock()
	if p.ExecCmd.Process != nil {
		p.ExecCmd.Process.Signal(p.StopSingal)
	}
	deadline := time.Now().Add(time.Second * time.Duration(p.StopTime))
	p.mu.Unlock()



	for time.Now().Before(deadline) {
	
		p.mu.Lock()
		fmt.Println(p.ExecCmd.ProcessState)
		if p.ExecCmd.ProcessState != nil {
			p.Stoped.Store(true)
			p.mu.Unlock()
			return 
		}
		p.mu.Unlock()
		time.Sleep(500 * time.Millisecond)
	}


	p.mu.Lock()
	if p.ExecCmd.ProcessState == nil {
		p.Stoped.Store(true)
		fmt.Println(p.ExecCmd.ProcessState)
		p.ExecCmd.Process.Kill()
	}
	p.mu.Unlock()
}



func (p *ProcesseSupervisor) KillProcess() {

	if p.Stoped.Load() {
		return
	}

	p.Stoped.Store(true)

	p.mu.Lock()
	if p.ExecCmd.Process != nil {
		p.ExecCmd.Process.Kill()
	}
	p.mu.Unlock()
}


func (p *ProcesseSupervisor) InitCmd() *errorshandling.ErrorReporter {
	var (
				err error
	)		

	p.mu.Lock()
	defer p.mu.Unlock()
	
	fullCommand := fmt.Sprintf("umask %s && %s", p.Umask, p.Cmd)


	p.ExecCmd = exec.Command( "bash", "-c", fullCommand)
	
	
	// set working directory 
	p.ExecCmd.Dir = p.WorkingDIr

	// f, err := os.OpenFile("lines", os.O_APPEND|os.O_WRONLY, 0644)

	if p.Stdout != ""{
		p.stdoutFile, err = os.OpenFile(p.Stdout, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0644)
		if err != nil {
			return errorshandling.NewErrorReporter(errorshandling.ErrCreatingTheProcess, err.Error() + " at " + p.Name)
		}
		p.ExecCmd.Stdout = p.stdoutFile
	}

	if p.Stderr != ""{
		p.stderrFile, err = os.OpenFile(p.Stderr, os.O_WRONLY | os.O_TRUNC| os.O_CREATE, 0644)
		if err != nil {
			p.stdoutFile.Close()
			return errorshandling.NewErrorReporter(errorshandling.ErrCreatingTheProcess, err.Error() + " at " + p.Name)
		}
		p.ExecCmd.Stderr = p.stderrFile
	}


	for k, val := range p.Env {
		p.ExecCmd.Env = append(p.ExecCmd.Environ(), fmt.Sprintf("%s=%s", k, val))
	}


	return nil
}


func (p *ProcesseSupervisor) CleanResources(){

	p.stderrFile.Close()
	p.stdoutFile.Close()
}

func (p *ProcesseSupervisor) startCmd() error {

	p.mu.Lock()
	defer p.mu.Unlock()

	if err := p.ExecCmd.Start(); err != nil {

				return err
	}
	return nil
}



func (p *ProcesseSupervisor) exitCode() int {
	err := p.ExecCmd.Wait()
	fmt.Printf("Process exited. Result: %v\n", err)

	p.mu.RLock()
	defer p.mu.RUnlock()


	if p.ExecCmd.ProcessState != nil {
		return p.ExecCmd.ProcessState.ExitCode()
	}

	return -1
}


func (p *ProcesseSupervisor) String() string {
	return fmt.Sprintf("%d:%s", p.Id, p.Name)
}


func (p *ProcesseSupervisor) MustStop(exitCode int ) bool {
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	
	if p.cuRetry <= p.StartRetries {
		return false
	}



	if p.AutoRestart == "always"{
		return false
	}


	if p.AutoRestart == "unexpected"{
		_, ok := p.ExitCodes[exitCode]
		return !ok
	}


	return true
}


func (p *ProcesseSupervisor) Loop() bool {

	p.mu.Lock()
	defer p.mu.Unlock()
	p.cuRetry++
	return p.cuRetry <= p.StartRetries
}

func (p *ProcesseSupervisor) Start() {
	defer p.wg.Done()
	p.Stoped.Store(false)
	defer p.Stoped.Store(true)
	p.ResetCurRetry()
	for p.Loop() {
			p.StartedAt = time.Now()
			if p.Stoped.Load() == true{
				break
			}
			if err := p.InitCmd(); err != nil {
				err.Report()
				continue
			}

			if p.startCmd() != nil {
				p.CleanResources()
				continue
			}
	

			exitCode := p.exitCode()
			p.CleanResources()
		
			// check if it need to be restarted 
			if p.MustStop(exitCode){
				break
			}
	}
}

func (p *ProcesseSupervisor) CanbeUpdated(newp *parser.Processes) bool {
	if p.Cmd != newp.Cmd || p.WorkingDIr != newp.WorkingDIr {
		return false
	}
	if p.Stderr != newp.Stderr || p.Stdout != newp.Stdout {
		return false
	}
	if p.Umask != fmt.Sprintf("%03o", newp.Umask) {
		return false
	}

	if !maps.Equal(p.Env, newp.Env) {
		return false
	}

	return true
}

func (p *ProcesseSupervisor) Update(newp *parser.Processes) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.AutoStart = newp.AutoStart
	p.AutoRestart = newp.AutoRestart
	p.StartRetries = newp.StartRetries
	p.StartTime = newp.StartTime
	p.StopTime = newp.StopTime
	exitCodes := make(map[int]bool)
	if exitcodes, ok := newp.ExitCodes.([]int); ok{
		for i := range(exitcodes){
			exitCodes[exitcodes[i]] = true
		}
	}
	if exitcodes, ok := newp.ExitCodes.(int); ok{
			exitCodes[exitcodes] = true
	}
	p.ExitCodes = exitCodes
}


func (p *ProcesseSupervisor) ResetCurRetry(){
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cuRetry = 0
}