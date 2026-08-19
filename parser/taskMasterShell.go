package parser

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/hatim-lahwaouir/taskmaster/errorshandling"
)



type ParseCmds struct {
	programsName map[string]bool
	validCmd  map[string]bool

	cmdWithoutArgs  map[string]bool
}


func NewParseCmds(programNames []string) *ParseCmds{
	mp := make(map[string]bool)
	for s := range(programNames){
		mp[programNames[s]] = true
	}

	return &ParseCmds{programsName:  mp, validCmd:  map[string]bool{ "status" : true, "stop" : true, "start" : true, "shutdown" : true, "help" : true, "load" : true,"restart": true,},
	cmdWithoutArgs: map[string]bool{ "shutdown" : true, "help" : true, "load" : true,},}
}

func (p *ParseCmds) ParseInput(input string) ([]string, *errorshandling.ErrorReporter) {
	input = strings.TrimSpace(input)

	cmd := strings.Fields(input)

	


	if _, ok := p.validCmd[cmd[0]] ; !ok {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrInvalidCMD, input)
	}

	if ( cmd[0] == "help" || cmd[0] == "load" || cmd[0] == "shutdown"  ) && len(cmd) == 1 {
		return cmd,nil
	}


	if len(cmd) != 1 &&  p.cmdWithoutArgs[cmd[0]] {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrInvalidCMD, " args must be provided correctly " + input)
	}
	if len(cmd) != 2 {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrInvalidCMD, " args must be provided correctly " + input)
	}

	if _, ok := p.programsName[cmd[1]] ; !ok {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrInvalidCMD, " args must be provided correctly " + input)
	}


	if len(cmd) != 2 {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrInvalidCMD, input)
	}

	return cmd, nil
}


func (p *ParseCmds) Help(b *bytes.Buffer) {

	fmt.Fprintln(b,"available cmds: [help, load,shutdown, (status, start, stop) program name]")
	fmt.Fprintln(b,"load    : do it if you change the config.yml file")
	fmt.Fprintln(b,"shutdown: to shutdown you supervisor")
	fmt.Fprintln(b,"status  : status of a process [process name]")
	fmt.Fprintln(b,"stop    : to stop a process [process name]")
	fmt.Fprintln(b,"restart : to restart a process [process name]")
	fmt.Fprintln(b,"help    : to see available cmds")
}


