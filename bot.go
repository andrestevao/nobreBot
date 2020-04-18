package main

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"fmt"
	"os"
	"strings"
	_ "github.com/lib/pq"
	irc "github.com/thoj/go-ircevent"
)

const serverssl = "irc.chat.twitch.tv:6667"

var (
	db       *sql.DB	
)

var config = make(map[string]string)
var simpleCommands = make(map[string]string)
var users = make(map[string]int)

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

func initDb() {
	var err error
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		config["dbhost"], config["dbport"],
		config["dbuser"], config["dbpassword"], config["dbname"])

	db, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected!")

	basicDbConfig()
}

func basicDbConfig() {
	query := "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE';"

	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	tables := make([]string, 0)
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			panic(err)
		}
		fmt.Println("table found: "+table)
		tables = append(tables, table)
	}

	expectedTables := []string{"commands", "users"}
	notFound := make([]string, 0)

	for _, expectedTable := range expectedTables {
		_, found := Find(tables, expectedTable)
		if !found {
			notFound = append(notFound, expectedTable)
		}
	}

	if len(notFound) > 0 {
		var text string
		for (text != "y" && text != "n") {
			fmt.Println("Expected tables \""+strings.Join(notFound, ",")+"\" not found. Do you wish to create them? (y/n)")
			reader := bufio.NewReader(os.Stdin)
			text, _ = reader.ReadString('\n')
			text = strings.TrimSpace(text)
			text = strings.ToLower(text)
		}

		if strings.ToLower(text) == "n" {
			panic("Cannot proceed without these tables created. Please check the HELPME file to create the tables manually or select Y next time to create the tables automatically.")
		}
		
		queries := make(map[string]string)

		queries["commands"] = `CREATE TABLE public.commands (
			id serial NOT NULL,
			privilege int NOT NULL,
			command varchar NOT NULL,
			type varchar NOT NULL,
			response varchar NULL,
			created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			created_by text NOT NULL,
			CONSTRAINT commands_pk PRIMARY KEY (id)
		);`

		queries["users"] = `CREATE TABLE public.users (
			id serial NOT NULL,
			privilege int NOT NULL,
			name varchar NOT NULL,
			created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			updated TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
			created_by text NOT NULL,
			CONSTRAINT users_pk PRIMARY KEY (id)
		);`
		
		for _, missingTable := range expectedTables {
			createQuery := queries[missingTable]
			_, err := db.Query(createQuery)
			if err != nil {
				fmt.Println("Error while creating table \""+missingTable+"\":")
				panic(err)
			} else {
				fmt.Println("Table \""+missingTable+"\" created sucessfully!")
			}
		}

	}

	
}

func loadCommands() {
	//loading simple commands first
	query := "SELECT command, response FROM commands WHERE type = 'simple';"
	
	rows, err := db.Query(query)
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var (
			command string
			response string
		)	

		if err := rows.Scan(&command, &response); err != nil {
			panic(err)
		}
		
		fmt.Println("command found: "+command)
		simpleCommands[command] = response

	}
}

func createConfig() {
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

func initIrc(nickname string, password string) {
	ircnick1 := nickname
	irccon := irc.IRC(ircnick1, "IRCTestSSL")
	irccon.Password = password

	irccon.VerboseCallbackHandler = true
	irccon.Debug = true
	irccon.UseTLS = false
	irccon.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	irccon.AddCallback("001", func(e *irc.Event) {
		irccon.Join(config["channel"])
		irccon.Privmsg(config["channel"], "Bot entrou na sala com sucesso! :)")

	})
	irccon.AddCallback("366", func(e *irc.Event) {})
	err := irccon.Connect(serverssl)
	irccon.AddCallback("PRIVMSG", func(e *irc.Event) {
		message := strings.TrimSpace(e.Message())
		if strings.HasPrefix(message, "!"){
			commandAndArguments := strings.Split(message, " ")
			irccon.Privmsg(config["channel"], getCommand(e.Nick, commandAndArguments[0], commandAndArguments[1:]))			
		}
	})
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

	irccon.Loop()
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

func main() {
	createConfig()
	initDb()
	loadCommands()
	initIrc(config["nickname"], config["password"])
}
