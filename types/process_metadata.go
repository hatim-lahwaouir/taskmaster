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
    Exitcodes []any `name:"exitcodes"`
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
        return errors.New("Provided value type didn't match obj field type")
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
