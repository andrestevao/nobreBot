package main

import (
	"fmt"
	"strings"
)

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}



func getCommand(user string, command string, arguments []string) string{
	fmt.Println("Command: "+command+"; Arguments:"+strings.Join(arguments, ","))

	//check if special command: add, remove, edit
	specialCommands := []string{"!add", "!remove", "!edit"}
	if _, found := Find(specialCommands, command); found{
		specialCommand(user, command, arguments)
	}

	//else keep on it like it is a regular command
	//to do
	return "to-do"
}

func specialCommand(user string, command string, arguments []string) string {
	if command == "!add" {
		fmt.Println("todo")
	}

	return "to-do"
}