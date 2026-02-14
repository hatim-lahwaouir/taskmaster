package handlers


import (
    "fmt"
    "time"
    "strings"
    "bufio"
    "os"
    "slices"
)




func cmdValidation(input string, pnames map[string]bool) error {
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

func CmdLine(cmd chan <- string, pnames map[string]bool) {

    var (
        input string
    )


    scanner := bufio.NewScanner(os.Stdin)
    for ;; {
        time.Sleep(100 * time.Millisecond)
        fmt.Printf("taskmaster> ")
        scanner.Scan()
        input = scanner.Text() 
        if err := cmdValidation(input, pnames); err != nil {
		    Loggers.ErrorLogger.Printf("Invalid Cmd  %v\n", err)
            continue
        }
        cmd <- input 
    }
}
