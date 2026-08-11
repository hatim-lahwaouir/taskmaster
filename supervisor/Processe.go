package supervisor

import (
	"context"
	"fmt"
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
	Name string
	Cmd string 											
	Umask string 					
	WorkingDIr string 				
	AutoStart bool 					
	AutoRestart string 					
	StartRetries int 					
	StartTime int64  					
	Stdout string 						
	Stderr string 						
	ExitCodes map[int]bool							
	Env map[string]string
	StopSingal os.Signal
	
	// to check status 
	StartedAt time.Time

	// context for canceling the cmd
	Ctx context.Context
	cancel context.CancelFunc
	
	// started 



	// to check if cmd stoped 
	Stoped atomic.Bool

	// waiting for clearence
	wg *sync.WaitGroup
}



func NewProcesseSupervisor(p *parser.Processes, wg *sync.WaitGroup) *ProcesseSupervisor {

	ctx, cancel := context.WithCancel(context.Background())
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
		Ctx:          ctx,
		cancel:       cancel,
		StopSingal: p.GetSingal(),
		wg: wg,
	}

	process.Stoped.Store(true)
	return &process
}

func (p *ProcesseSupervisor) Status(w *tabwriter.Writer) {
	if p.Stoped.Load() {
		fmt.Fprintf(w," %s\t| -1\t| stoped\t\n", p.Name)
		return
	}
	duration := time.Since(p.StartedAt).Round(time.Millisecond)

	if p.StartedAt.Add(time.Duration(p.StartTime * int64(time.Second))).Compare(time.Now()) > 0{
		fmt.Fprintf(w," %s\t| %s\t| started\t\n", p.Name, duration)
		return 
	}

	fmt.Fprintf(w," %s\t| %s\t| running\t\n", p.Name, duration)
}

func (p *ProcesseSupervisor) Stop() {
	p.cancel()
	p.Stoped.Store(true)

}



// we need to start processes if autostart was set to true 


func (p *ProcesseSupervisor) Start() {
	defer p.wg.Done()
	p.Stoped.Store(false)
	defer p.Stoped.Store(true)
	p.StartedAt = time.Now()

	for i := 0 ; i < p.StartRetries; i++ {
			fullCommand := fmt.Sprintf("umask %s && %s", p.Umask, p.Cmd)


			cmd := exec.CommandContext(p.Ctx, "bash", "-c", fullCommand)
			
			
			// set working directory 
			cmd.Dir = p.WorkingDIr

			// f, err := os.OpenFile("lines", os.O_APPEND|os.O_WRONLY, 0644)

			if p.Stdout != ""{
				stdout, err := os.OpenFile(p.Stdout, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0644)
				if err != nil {
					errorshandling.NewErrorReporter(errorshandling.ErrCreatingTheProcess, err.Error() + " at " + p.Name).Report()
					continue
				}
				cmd.Stdout = stdout
			}

			if p.Stderr != ""{
				stdErr, err := os.OpenFile(p.Stderr, os.O_WRONLY | os.O_TRUNC| os.O_CREATE, 0644)
				if err != nil {
					errorshandling.NewErrorReporter(errorshandling.ErrCreatingTheProcess, err.Error() + " at " + p.Name).Report()
					continue
				}
				cmd.Stderr = stdErr
			}


			// setting stdout 
			

			// setting env 
			for k, val := range p.Env {
				cmd.Env = append(cmd.Environ(), fmt.Sprintf("%s=%s", k, val))
			}

			
			if err := cmd.Run(); err != nil {
				fmt.Println("err command stoped", err)
			}
	}

}


