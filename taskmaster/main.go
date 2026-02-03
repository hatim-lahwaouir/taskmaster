package main

import (
	"flag"
	"fmt"
    "errors"
    "github.com/hatim-lahwaouir/taskmaster/types"
	"os"
    "github.com/hatim-lahwaouir/taskmaster/loggers"
    pm "github.com/hatim-lahwaouir/taskmaster/processMetadata" 
    "gopkg.in/yaml.v3"
)


var  args types.CmdArgs
var  Loggers types.Loggers  =  loggers.ProgramLogs



func init() {
	// using for parssing arguments
	flag.StringVar(&args.ConfigPath, "config", "", "config path")

	flag.Usage = func() {
		fmt.Printf("Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
	flag.Parse()

	if len(args.ConfigPath) == 0 {
		flag.Usage()
	}
}




func Parsing(data []byte) ([]pm.ProcessMetadata, error) {
     var (
        programs map[string]interface{}
        result []pm.ProcessMetadata
        processMetadata pm.ProcessMetadata
     )

     programs = make(map[string]interface{})

     err := yaml.Unmarshal(data, &programs)
     if err != nil {
        return nil, fmt.Errorf("parssing yaml file ", err.Error())
     }
     if _, ok := programs["programs"]; !ok {
        return nil, errors.New("we don't have programs")
     } 
     programs = programs["programs"].(map[string]interface{})
     
     for key ,v := range(programs){
         m2 := v.(map[string]interface{})
         processMetadata = pm.New()
         processMetadata.ProcessName = key
         err := processMetadata.FillStruct(m2)
         if err != nil {
                return nil, err
         }
         if err := processMetadata.ParseValidate(); err != nil {
                return nil, err
         }
         result = append(result, processMetadata)
     }

	 return result,nil
}


func OpenConfig(confpath string) ([]byte, error) {
	return os.ReadFile(confpath)
}



func main() {
    var (
        processesMetadata []pm.ProcessMetadata
    )
    // openning config file
	data, err := OpenConfig(args.ConfigPath)
	if err != nil {
        Loggers.ErrorLogger.Printf("Reading config file %s\n",err.Error())
		os.Exit(1)
	}

    // Parsing config file 
    processesMetadata, err = Parsing(data)
    if err != nil {
    	Loggers.ErrorLogger.Printf("Invalid config file '%s' \n", err.Error())
		os.Exit(1)
    }


    for  _, pm := range(processesMetadata) {
       if  err := pm.DataValidation(); err != nil {
       	    Loggers.ErrorLogger.Printf("Invalid data, %s \n", err.Error())
		    os.Exit(1)
       }
    }
    Loggers.InfoLogger.Println(processesMetadata)

}
