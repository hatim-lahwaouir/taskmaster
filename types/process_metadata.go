package types

import (

	"fmt"
    "errors"
    "reflect"
)


type ProcessMetadata struct {
    ProcessName string
    Cmd string `name:"cmd"`
    NumProcess int `name:"numprocs"`
    Umask int `name:"umask"`
    Workdir string `name:"workingdir"`
    Autostart bool `name:"autostart"`
    Autorestart string `name:"autorestart"`
    Exitcodes []interface{} `name:"exitcodes"`
    Startretries int`name:"startretries"`
    Starttime int `name:"starttime"`
    Stopsignal string `name:"stopsignal"`
    Stoptime int `name:"stoptime"`
    Stdout string `name:"stdout"`
    Stderr string `name:"stderr"`
    Env map[string]any `name:"env"`
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
        return  nil
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
        return errors.New("syntax error at " + name )
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

func (s *ProcessMetadata) ExitStatusValidation() error {

    if reflect.ValueOf(s.Exitcodes).Kind() != reflect.Slice {
        return errors.New("Invalid Exit codes it must be an integer")
    }

    for _ , ele := range(s.Exitcodes){
        if reflect.ValueOf(ele).Kind() != reflect.Int || ele.(int) < 0 || ele.(int) > 255 {
            return errors.New("Invalid Exit code it must be between 255 and 0")
        }
    }
    return nil
}



func (s *ProcessMetadata) EnvValidation() error {

    for _ , ele := range(s.Env){
        if reflect.ValueOf(ele).Kind() == reflect.Map || reflect.ValueOf(ele).Kind() == reflect.Slice || reflect.ValueOf(ele).Kind() == reflect.Slice {
            return errors.New("env must be key:value pair")
        }
    }
    return nil
}
