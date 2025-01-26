package main

import (
	metadata "LiteCanary/internal"
	"LiteCanary/internal/client"
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
)

var (
	liteClient *client.Client
	url        string
	sc         chan os.Signal // SC (Shutdown Channel)

)

func main() {
	fmt.Printf(">-LiteCanary CLI %s>\n", metadata.Version)

	// Parses command line arguments
	flag.StringVar(&url, "url", "http://127.0.0.1:8080/api", "LiteCanary server url (http://127.0.0.1:8080/api)")
	flag.Parse()
	if url == "http://127.0.0.1:8080/api" {
		fmt.Println("warning: http://127.0.0.1:8080/api selected as default endpoint. use --url to change it.")
	}

	// Set up handler for shutdown signal
	sc = make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt)
	go catchInterrupt()

	// Gets user input and executes command
	reader := bufio.NewReader(os.Stdin)
	liteClient = client.New(url)
	for {
		if liteClient.LoggedIn {
			fmt.Printf("LiteCanary Client (%s) > ", liteClient.Username)
		} else {
			fmt.Print("LiteCanary Client > ")
		}
		command, _ := reader.ReadString('\n')
		CommandParser(command)
	}

}

func CommandParser(command string) {
	command = removeDoubleSpaces(command) // Removes double spaces ("rm   <id>" --> "rm <id>")
	sections := strings.Split(strings.TrimSpace(command), " ")
	sl := len(sections)

	switch strings.ToLower(sections[0]) {
	case "exit":
		exit()
	case "help":
		fmt.Print(`help: displays help page
exit: exits the program

user:
 reset <new password>: resets user password
 deleteme: deletes your account and canaries (WARNING: YOU WILL NOT BE PROMPTED FOR A CONFIRMATION)
 login <username> <password>: logs in
 register <username> <password>: registers a new user. please don't use spaces in your username nor password

acceptable canary types:
 image: a 1x1 cyan pixel. (for emails and documents)
 text: displays "This is a test page." 
 redirect: redirects the user to a specific url

canary:
 wipe <id>: clears the event history
 rm <id>: deletes specific canary.
 new <name> <type>: creates a new canary.
 update <id> <name> <type> <redirect>: update a canary. redirect can be anything if you don't use it.
 get <id>: gets all the events for a specific canary.

 `)
	case "wipe":
		if !liteClient.LoggedIn {
			fmt.Println("you need to log in first")
			break
		}
		if sl < 2 {
			fmt.Println("too few arguments")
			break
		}
		if !Error(liteClient.WipeCanary(sections[1])) {
			fmt.Println("canary events wiped successfully")
		}
	case "reset":
		if !liteClient.LoggedIn {
			fmt.Println("you need to log in first")
			break
		}
		if sl < 2 {
			fmt.Println("too few arguments")
			break
		}
		if !Error(liteClient.ResetPassword(sections[1])) {
			fmt.Println("password reset successfully")
		}
	case "deleteme":
		if !liteClient.LoggedIn {
			fmt.Println("you need to log in first")
			break
		}
		if !Error(liteClient.DeleteUser()) {
			fmt.Println("user deleted")
		}
	case "rm":
		if !liteClient.LoggedIn {
			fmt.Println("you need to log in first")
			break
		}
		if sl < 2 {
			fmt.Println("too few arguments")
			break
		}
		if !Error(liteClient.DeleteCanary(sections[1])) {
			fmt.Println("canary deleted")
		}
	case "login":
		if sl < 3 {
			fmt.Println("too few arguments")
			break
		}
		if !Error(liteClient.Login(sections[1], sections[2])) {
			fmt.Println("logged in")
		}
	case "register":
		if sl < 3 {
			fmt.Println("too few arguments")
			break
		}
		if !Error(liteClient.Register(sections[1], sections[2])) {
			fmt.Println("new user registered")
		}
	case "new":
		if !liteClient.LoggedIn {
			fmt.Println("you need to log in first")
			break
		}
		if sl < 3 {
			fmt.Println("too few arguments")
			break
		}
		if !Error(liteClient.NewCanary(sections[1], sections[2])) {
			fmt.Println("canary created")
		}
	case "update":
		if !liteClient.LoggedIn {
			fmt.Println("you need to log in first")
			break
		}
		if sl < 5 {
			fmt.Println("too few arguments")
			break
		}
		if !Error(liteClient.UpdateCanary(sections[1], sections[2], sections[3], sections[4])) {
			fmt.Println("canary updated")
		}
	case "get":
		if !liteClient.LoggedIn {
			fmt.Println("you need to log in first")
			break
		}
		if len(sections) == 2 {
			canary, err := liteClient.GetCanary(sections[1])
			if Error(err) {
				break
			}
			if canary.History == nil {
				fmt.Println("no events")
				break
			}
			fmt.Printf("canary: %s\n", canary.Name)
			for _, event := range *canary.History {
				fmt.Printf(" %s:\n  IP: %s\n  User Agent: %s\n  Keyboard: %s \n\n", event.Timestamp, event.Ip, event.Useragent, event.Keyboardlanguage)
			}
			break
		}
		canaries, err := liteClient.UpdateCanaries()
		if Error(err) {
			break
		}
		for _, canary := range canaries.Canaries {
			fmt.Printf("name: %s | type: %s | id: %s | url: %s\n", canary.Name, canary.Type, canary.Id, url+"/trigger/"+canary.Id)
			if canary.History == nil {
				continue
			}
			fmt.Printf(" triggered (%d times)\n\n", len(*canary.History))
		}
	default:
		fmt.Println("invalid command. use the 'help' command to get help")
	}
}

// Helpers
func catchInterrupt() {
	for range sc {
		exit()
	}
}

func exit() {
	fmt.Println("\nexiting...")
	os.Exit(0)
}

func removeDoubleSpaces(command string) string {
	for i := 0; i < strings.Count(command, " "); i++ {
		command = strings.ReplaceAll(command, "  ", " ")
	}
	return command
}

func Error(err error) bool {
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
	}
	return err != nil
}
