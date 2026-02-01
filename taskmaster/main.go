package main

import (
	"flag"
	"fmt"
    "github.com/hatim-lahwaouir/taskmaster/types"
    "log"
	"os"
    "gopkg.in/yaml.v3"
)


var  args types.CmdArgs

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



func Parsing(data []byte) error {
     programs := make(map[string]interface{})

     err := yaml.Unmarshal(data, &programs)
     if err != nil {
         return err
     }
     if _, ok := programs["programs"]; !ok {
        log.Fatal("we don't have programs")
     } 
     programs = programs["programs"].(map[string]interface{})
     
     for _ ,v := range(programs){
         m2 := v.(map[string]interface{})
         result := &types.ProcessMetadata{}
         err := result.FillStruct(m2)
         if err != nil {
            log.Fatal(err)
         }
         fmt.Println(result)
     }
	 return nil
}

func OpenConfig(confpath string) ([]byte, error) {
	return os.ReadFile(confpath)
}



func main() {
    // openning config file
	data, err := OpenConfig(args.ConfigPath)
	if err != nil {
		fmt.Printf("Invalid path '%s'\n", os.Args[1])
		os.Exit(1)
	}

    // Parsing config file 
    err =  Parsing(data)
    if err != nil {
    	fmt.Printf("Invalid config file error '%s' \n", err.Error())
		os.Exit(1)
    }
}
