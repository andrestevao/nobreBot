package main

import (
	"fmt"
	"strconv"
	"strings"
)

var commands = make(map[string]string)

func LoadCommands() {
	for k := range commands {
		delete(commands, k)
	}

	//loading simple commands first
	query := `SELECT command, response
				FROM commands WHERE type = 'simple'
				AND active = 1;`
	
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var (
			command string
			privilege int
			response string
		)	

		if err := rows.Scan(&command, &response); err != nil {
			panic(err)
		}
		
		fmt.Println("command found: "+command)
		commands[command] = strconv.Itoa(privilege)+":::"+response

	}
}

func GetCommand(user string, command string, arguments []string) string{
	//remove ! from start of command
	command = command[1:]

	//handles special commands through another function
	specialCommands := []string{"addCommand", "removeCommand", "editCommand", "commands"}	
	if _, found := Find(specialCommands, command); found {
		return SpecialCommand(user, command, arguments)
	}
	//searches through the memory to find command
	//and checks if user has necessary privileges to execute it
	//map commands is structured like this:
	// string command name -> string details
	// details = "privilege:::response"
	
	if debug {
		fmt.Println("Commands.go -> GetCommand: received (user:"+user+";command:"+command+";arguments:"+strings.Join(arguments, ",")+")")
	}
	
	if _, ok := commands[command]; !ok {
		return "Command not found"
	}
	selectedCommand := commands[command]
	sliceSelectedCommand := strings.Split(selectedCommand, ":::")
	selectedCommandPrivilege, _ := strconv.Atoi(sliceSelectedCommand[0])
	selectedCommandResponse := sliceSelectedCommand[1]
	if(selectedCommandPrivilege == 0){
		return "@"+user+" "+selectedCommandResponse
	}
	
	return "@"+user+", you don't have permission to execute this command (to-do privilege logic)!"
	
	
}

func SpecialCommand(user string, command string, arguments []string) string {
	var response string
	if command == "addCommand" {
		phrase := arguments[1:]
		query := `insert into commands (privilege, command, "type", response, created_by) 
				values (0, '`+arguments[0]+`', 'simple', '`+strings.Join(phrase, " ")+`', '`+user+`')`
		_, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		LoadCommands()
		response = "Command "+arguments[0]+" added successfully!"
	} else if command == "removeCommand" {
		query := `update commands set active = 0
				where command = '`+arguments[0]+`' and active = 1`
		_, err := db.Query(query)
		if err != nil {
			panic(err)
		}
		LoadCommands()
		response = "Command "+arguments[0]+" removed successfully!"
	} else if command == "editCommand" {
		response = "TO-DO"
	} else if command == "commands" {
		var allCommands []string
		for key, _ := range commands {
			allCommands = append(allCommands, "!"+key)
		}

		return strings.Join(allCommands, ", ")

	}

	return "@"+user+" "+response
}