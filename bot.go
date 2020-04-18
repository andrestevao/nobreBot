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
var commands = make(map[string]string)

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
		panic("Expected tables \""+strings.Join(notFound, ",")+"\" not found ")
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
	// ircnick1 := "nobrebot"
	ircnick1 := nickname
	irccon := irc.IRC(ircnick1, "IRCTestSSL")
	// irccon.Password = "oauth:m0lr6jx3vm8n6zaa358b8xyry37gpt"
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
			irccon.Privmsg(config["channel"], getCommand(commandAndArguments[0], commandAndArguments[1:]))			
		}
	})
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

	irccon.Loop()
}

func getCommand(command string, arguments []string) string{
	return "Command: "+command+"; Arguments:"+strings.Join(arguments, ",")
}

func main() {
	createConfig()
	initDb()
	initIrc(config["nickname"], config["password"])
}
