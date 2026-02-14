package processMetadata

import (
	"errors"
	"fmt"
	"github.com/hatim-lahwaouir/taskmaster/utils"
	"os"
	"reflect"
	"slices"
)

type ProcessMetadata struct {
	ProcessName  string
	User         string         `name: User`
	Cmd          string         `name:"cmd"`
	NumProcess   int            `name:"numprocs"`
	Umask        int            `name:"umask"`
	Workdir      string         `name:"workingdir"`
	Autostart    bool           `name:"autostart"`
	Autorestart  string         `name:"autorestart"`
	Exitcodes    []interface{}  `name:"exitcodes"`
	Startretries int            `name:"startretries"`
	Starttime    int            `name:"starttime"`
	StopSignal   string         `name:"stopsignal"`
	Stoptime     int            `name:"stoptime"`
	Stdout       string         `name:"stdout"`
	Stderr       string         `name:"stderr"`
	Env          map[string]any `name:"env"`
}

func New() ProcessMetadata {
	// set some default value
	return ProcessMetadata{NumProcess: 1, Workdir: ".", Stdout: os.DevNull, Stderr: os.DevNull}
}

func Flatten(pmetadata []ProcessMetadata) []ProcessMetadata {

	var (
		result []ProcessMetadata
	)
	for _, p := range pmetadata {
		for i := 0; i < p.NumProcess; i += 1 {

			result = append(result, p)
			result[len(result)-1].NumProcess = 1
		}
	}

	return result
}

func SetField(obj interface{}, name string, value interface{}) error {
	var structFieldValue reflect.Value

	s := reflect.TypeOf(obj).Elem()
	sv := reflect.ValueOf(obj).Elem()

	for i := 0; i < s.NumField(); i++ {
		if alias, ok := s.Field(i).Tag.Lookup("name"); ok && alias == name {
			structFieldValue = sv.Field(i)
			break
		}
	}

	if structFieldValue.IsValid() == false {
		return nil
	}

	if !structFieldValue.IsValid() {
		return fmt.Errorf("No such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("Cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()

	val := reflect.ValueOf(value)

	if structFieldType != val.Type() {
		return fmt.Errorf("Syntax at %s ", name)
	}

	structFieldValue.Set(val)
	return nil
}

func (s *ProcessMetadata) FillStruct(m map[string]interface{}) error {
	for k, v := range m {
		err := SetField(s, k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *ProcessMetadata) exitStatusValidation() error {

	if reflect.ValueOf(s.Exitcodes).Kind() != reflect.Slice {
		return errors.New("Invalid Exit codes it must be an integer")
	}

	for _, ele := range s.Exitcodes {
		if reflect.ValueOf(ele).Kind() != reflect.Int || ele.(int) < 0 || ele.(int) > 255 {
			return errors.New("Invalid Exit code it must be between 255 and 0")
		}
	}
	return nil
}

func (s *ProcessMetadata) envValidation() error {

	for _, ele := range s.Env {
		if reflect.ValueOf(ele).Kind() == reflect.Map || reflect.ValueOf(ele).Kind() == reflect.Slice || reflect.ValueOf(ele).Kind() == reflect.Slice {
			return errors.New("env must be key:value pair")
		}
	}
	return nil
}

func (s *ProcessMetadata) ParseValidate() error {
	if err := s.exitStatusValidation(); err != nil {
		return err
	}
	if err := s.envValidation(); err != nil {
		return err
	}
	return nil
}

// check if Cmd is exucutable
// NumProcess must be if it was 0
// Umask  must be validated
// Workdir must be an existing directoy
// startretries must 0 >= 0
// Starttime must be 0 >= 0
// StopSignal must be a Valid signal
// Stoptime must be 0 >= 0

func (s *ProcessMetadata) DataValidation() error {

	// validating if cmd is executable
	fi, err := os.Lstat(s.Cmd)
	if fi.Mode()&0111 == 0 {
		return fmt.Errorf("%s cmd  '%s' can't be executed", s.ProcessName, s.Cmd)
	}
	// check if working directory exists
	fi, err = os.Lstat(s.Workdir)
	if err != nil {
		return fmt.Errorf("%s working directory '%s' %s", s.ProcessName, s.Workdir, err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s working directory '%s' isn't a directory", s.ProcessName, s.Workdir)
	}

	// validating the umask value
	if 000 > s.Umask || s.Umask > 0777 {
		return fmt.Errorf("%s Invalid umask '%o'", s.ProcessName, s.Umask)
	}

	// validating the signal name
	if _, err := utils.GetSignal(s.StopSignal); len(s.StopSignal) != 0 && err != nil {
		return fmt.Errorf("%s Invalid signal '%s'", s.ProcessName, s.StopSignal)
	}

	// validating time must be positive
	if s.Starttime < 0 || s.Stoptime < 0 {
		return fmt.Errorf("%s Invalid time it must be positive", s.ProcessName)
	}
	if !slices.Contains([]string{"unexpected", "always", "never"}, s.Autorestart) {
		return fmt.Errorf("%s Invalid time autorestart value %s", s.ProcessName, s.Autorestart)
	}
	if "unexpected" == s.Autorestart && len(s.Exitcodes) == 0 {
		return fmt.Errorf("%s autorestart set to unexpected and no exitcodes ", s.ProcessName)
	}

	if s.Startretries < 0 {
		return fmt.Errorf("%s invalid Startretries", s.ProcessName)
	}
	if s.NumProcess <= 0 {
		return fmt.Errorf("%s invalid num process provided ", s.ProcessName)
	}
	return nil
}

func (s *ProcessMetadata) SetupFiles() error {

	f, err := os.Create(s.Stdout)
	if err != nil {
		return err
	}
	defer f.Close()

	f, err = os.Create(s.Stderr)
	if err != nil {
		return err
	}
	defer f.Close()

	return nil
}
