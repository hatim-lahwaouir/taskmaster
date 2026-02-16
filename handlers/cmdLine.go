package handlers


import (
    "fmt"
    "strings"
    "bufio"
    "os"
    "slices"
)




func cmdValidation(input string, pnames map[string][]int) error {
    var (
        arg []string
    )

    arg = strings.Split(input, " ")


    if len(arg) != 2 {
        return fmt.Errorf("usage: 'cmd:[start:status:stop] [process Name]'")
    }

    if ! slices.Contains([]string{"status", "start", "stop"}, strings.ToLower(arg[0])) {
        return fmt.Errorf("'%s'", arg[0])
    }

    if _, ok := pnames[strings.ToLower(arg[1])]; ok == false{ 
        return fmt.Errorf("process name unknown '%s'", arg[1])
    }

    return nil
}

func CmdLine(cmd chan <- string, info  chan string,  pnames map[string][]int) {

    var (
        input string
        resp string
    )


    scanner := bufio.NewScanner(os.Stdin)
    for ;; {
        fmt.Printf("taskmaster> ")
        scanner.Scan()
        input = scanner.Text() 
        if err := cmdValidation(input, pnames); err != nil {
		    Loggers.ErrorLogger.Printf("Invalid Cmd  %v\n", err)
            continue
        }
        cmd <- input 
        // after this we need to wait for process to give us info about process running
        resp = <- info
        fmt.Printf("taskmaster>\n%s", resp)
    }
}
