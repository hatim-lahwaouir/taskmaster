package signalhandler

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hatim-lahwaouir/taskmaster/parser"
	"github.com/hatim-lahwaouir/taskmaster/supervisor"
)



type RealoadConfigSig struct{

	sig chan os.Signal
	done chan bool
	master * supervisor.MasterSupervisor
	ymlParser *parser.YMLParser
} 



func NewRealoadConfigSig(m * supervisor.MasterSupervisor, p *parser.YMLParser) *RealoadConfigSig{
	return &RealoadConfigSig{done: make(chan bool), sig : make(chan os.Signal, 1), master: m , ymlParser: p}
}



func (r *RealoadConfigSig) Handler(wg *sync.WaitGroup) {
	defer wg.Done()

	signal.Notify(r.sig, syscall.SIGHUP)
	for ;; {

		select{
		case <- r.sig:
			processes, err := r.ymlParser.Start()
			if err  != nil {
				err.Report()
				continue
			}
			r.master.LoadConfig(processes)
		case <- r.done:
			return			
		}
	}
}


func (r *RealoadConfigSig) Stop() {
	r.done <- true
}


func (r *RealoadConfigSig) Clear() {
	close(r.done)
	close(r.sig)
}




