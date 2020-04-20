package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
)

var config = make(map[string]string)


func CreateConfig() {
	//if the file doesn't exist, create it
	_, err := os.Stat("config.nobrebot")
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		} else {
			fmt.Println("No config file found, please provide necessary information:")
			lines := [8]string{"nickname|", "password|", "channel|#", "dbhost|", "dbport|", "dbuser|", "dbpassword|", "dbname|"}
			newFile, err := os.Create("config.nobrebot")
			if err != nil {
				panic(err)
			}

			reader := bufio.NewReader(os.Stdin)
			for _, line := range lines {
				fmt.Print(line)
				text, _ := reader.ReadString('\n')
				fmt.Fprint(newFile, line+text)
			}

			newFile.Close()
		}

	}

	//then open the file
	configFile, err := os.Open("config.nobrebot")
	if err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(configFile)
	for scanner.Scan() {
		linha := scanner.Text()
		linhaArray := strings.Split(linha, "|")
		if len(linhaArray) == 2{
			fmt.Println("Setting "+linhaArray[0]+" to \""+linhaArray[1]+"\"")
			config[linhaArray[0]] = linhaArray[1]
		}
		
	}
}

