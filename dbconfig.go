package main

import (
	"fmt"
	"os"
	"bufio"
	"strings"
	_ "github.com/lib/pq"
	"database/sql"
)


var (
	db       *sql.DB	
)

func InitDb() {
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

	BasicDbConfig()
}

func BasicDbConfig() {
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

