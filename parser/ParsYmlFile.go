package parser

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/go-playground/validator/v10"
	"github.com/hatim-lahwaouir/taskmaster/errorshandling"
	"gopkg.in/yaml.v3"
)




type YMLParser struct {
	filePath string
	validate *validator.Validate


}


func NewYmlParser(filePath string)  *YMLParser {
	return &YMLParser{filePath: filePath, validate: validator.New(validator.WithRequiredStructEnabled())}
}


type Processes struct {
	Name string
	Cmd string 							`yaml:"cmd" validate:"required"`
	NumProcs int 						`yaml:"numprocs" validate:"required,min=1"`
	Umask int16 						`yaml:"umask" validate:"required,min=0,max=511"`
	WorkingDIr string 					`yaml:"workingdir" validate:"required"`
	AutoStart bool 						`yaml:"autostart"`
	AutoRestart string 					`yaml:"autorestart" validate:"required,oneof=unexpected always never"`
	StartRetries int 					`yaml:"startretries" validate:"required,min=0"`
	StartTime int64  					`yaml:"starttime" validate:"required,min=0"`
	Stdout string 						`yaml:"stdout"`
	Stderr string 						`yaml:"stderr"`
	ExitCodes any 						`yaml:"exitcodes"`
	Env map[string]string				`yaml:"env"`
	Stopsignal string    				`yaml:"stopsignal" validate:"required"` 
}

var ProcessParsingErrors  = map[string]string{
	"numprocs": "required, must be positif number",
	"umask" : "required, must be between 0, 511",
	"autorestart" : "required,  oneof=unexpected always never",
	"startretries": "required,  must be positif number",
	"starttime" : "required, must be positif number",
	"exitcodes" : "must be provided with autorestart set to unxpected",
}


var validSignals = map[string]os.Signal{
    "HUP":    syscall.SIGHUP,
    "INT":    syscall.SIGINT,
    "QUIT":   syscall.SIGQUIT,
    "ILL":    syscall.SIGILL,
    "TRAP":   syscall.SIGTRAP,
    "ABRT":   syscall.SIGABRT,
    "FPE":    syscall.SIGFPE,
    "KILL":   syscall.SIGKILL,
    "SEGV":   syscall.SIGSEGV,
    "PIPE":   syscall.SIGPIPE,
    "ALRM":   syscall.SIGALRM,
    "TERM":   syscall.SIGTERM,
    "USR1":   syscall.SIGUSR1,
    "USR2":   syscall.SIGUSR2,
    "CHLD":   syscall.SIGCHLD,
    "CONT":   syscall.SIGCONT,
    "STOP":   syscall.SIGSTOP,
    "TSTP":   syscall.SIGTSTP,
    "TTIN":   syscall.SIGTTIN,
    "TTOU":   syscall.SIGTTOU,
}


func (p * Processes) GetSingal() os.Signal{
	return validSignals[p.Stopsignal]
}


type Config struct {
    Programs map[string]*Processes `yaml:"programs"`
}



func (yml *YMLParser) validateConfig(config *Config) *errorshandling.ErrorReporter {
    
	

	for k, v := range config.Programs {
		err := yml.validate.Struct(v)
		
		if err != nil {

		// this check is only needed when your code could produce
		// an invalid value for validation such as interface with nil
		// value most including myself do not usually have code like this.
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(err, &invalidValidationError) {
			fmt.Println(err)
		}

			var validateErrs validator.ValidationErrors
			if errors.As(err, &validateErrs) {
				for _, e := range validateErrs {
					key := strings.ToLower(e.Field())
					fmt.Println(e)
					if err, ok := ProcessParsingErrors[key]; ok {
						return errorshandling.NewErrorReporter(errorshandling.ErrInvalidData, fmt.Sprintf("%s %s at %s", key, err, k))
					}else{
						return errorshandling.NewErrorReporter(errorshandling.ErrInvalidData, fmt.Sprintf("%s at %s", key, k))
					}
				
				}
			}
		}
		

		// validate autorestart
		if v.AutoRestart == "unexpected" && v.ExitCodes == nil {
			return errorshandling.NewErrorReporter(errorshandling.ErrInvalidData, fmt.Sprintf("exitcodes %s at %s", ProcessParsingErrors["exitcodes"], k))
		}

		// validate signal 
		if _, exists := validSignals[v.Stopsignal]; !exists {
			return errorshandling.NewErrorReporter(errorshandling.ErrInvalidData, fmt.Sprintf("stopsignal at %s", k))
	    }
		
	}
	return nil
}




func (yml *YMLParser) Start() ([]*Processes, * errorshandling.ErrorReporter) {
	var (
		config Config
		processes []*Processes
	)

	data, err := os.ReadFile(yml.filePath)


	if err != nil {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrInternal, err.Error())
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrDataWasntProvided, err.Error())
	}

	if len(config.Programs) == 0{
		return nil, errorshandling.NewErrorReporter(errorshandling.ErrDataWasntProvided, "Program argument in yml file")
	}


	if err := yml.validateConfig(&config); err != nil {
		return nil, err
	}
	
	for k := range(config.Programs){
		config.Programs[k].Name = k
		processes = append(processes, config.Programs[k])
	}

	return processes, nil
}



