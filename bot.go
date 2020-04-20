package main

var debug bool = true

func main() {
	//Check/create configuration file for irc/db
	CreateConfig()
	//Initiate db connection with info from config file
	InitDb()
	//Load commands from db into memory
	LoadCommands()
	//Log into IRC and start listening for commands
	InitIrc(config["nickname"], config["password"])
}
