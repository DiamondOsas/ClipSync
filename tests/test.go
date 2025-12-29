package main

import (
	"fmt"
	"log"
	"os"
)
func main(){
	name, err := os.Hostname();if err!= nil{
		log.Fatal(err)
	}
	fmt.Println(name)
	select{}
}
