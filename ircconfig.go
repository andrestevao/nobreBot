package main

import (
	irc "github.com/thoj/go-ircevent"
	"crypto/tls"
	"strings"
	"fmt"
)

const serverssl = "irc.chat.twitch.tv:6667"

func InitIrc(nickname string, password string) {
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
			irccon.Privmsg(config["channel"], GetCommand(e.Nick, commandAndArguments[0], commandAndArguments[1:]))			
		}
	})
	if err != nil {
		fmt.Printf("Err %s", err)
		return
	}

	irccon.Loop()
}