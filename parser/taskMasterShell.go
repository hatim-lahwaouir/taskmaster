package parser

import (
	"fmt"
	"strings"

	"github.com/hatim-lahwaouir/taskmaster/errorshandling"
)



type ParseCmds struct {
	programsName map[string]bool
}


func NewParseCmds(programNames []string) *ParseCmds{
	mp := make(map[string]bool)
	for s := range(programNames){
		mp[programNames[s]] = true
	}

	return &ParseCmds{programsName:  mp}
}

func (p *ParseCmds) ParseInput(input string) ([]string, *errorshandling.ErrorReporter) {
	input = strings.TrimSpace(input)
	validCmd := map[string]bool { "status" : true, "stop" : true, "start" : true, "shutdown" : true, "help" : true, "load" : true }
	cmd := strings.Fields(input)

	


	if _, ok := validCmd[cmd[0]] ; !ok {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrInvalidCMD, input)
	}

	if ( cmd[0] == "help" || cmd[0] == "load" ) && len(cmd) == 1 {
		return cmd,nil
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


func (p *ParseCmds) Help(){

	fmt.Println("available cmds: [help, load, (status, start, stop) program name]")
}


