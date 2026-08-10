package main

import (
	"fmt"
	"log"
	"os"
)



func main(){
	if len(os.Args) != 2 {
		log.Fatal("Invalid args")		
	}

	parser := NewYmlParser(os.Args[1])


	programs, err := parser.Start()
	if err != nil {
		log.Fatal(err)
	}

	for i := range(programs){
		fmt.Println(programs[i])
	}
}