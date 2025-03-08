<img width=700 src="https://github.com/user-attachments/assets/b1d995db-9fdc-4782-bbdd-5f2b07a05f49"></img><br>
[![Go/Golang](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
[![SQLite](https://img.shields.io/badge/sqlite-%2307405e.svg?style=for-the-badge&logo=sqlite&logoColor=white)](https://img.shields.io/badge/sqlite-%2307405e.svg?style=for-the-badge&logo=sqlite&logoColor=white)

<b>Light weight self-hostable canary alerts to catch snoopers red-handed.</b>

## What is a canary?
A canary, in the context of this project, is a URL set up so that when someone requests it, an alert is generated and sent to its owner. The URL is usually hidden by linking it to a document with a name like "my passwords" or something similar. Then, when someone comes snooping around and opens the file, you are notified. This provides an effective way to detect hackers during post-exploitation.

## Roadmap
- [x] <b>Basic server & client</b>
- [x] <b>Log file for alerts (could integrate with SIEM)</b>
- [ ] <b>Twilio/email integration?</b>
- [ ] <b>More server response types</b>
<p>More things might appear...</p>

## Installation
You have 2 options: you can either download a precompiled version of both the client and the server from the releases tab. Downloading a precompiled version is recommended.

### Compiling from source:
```bash
git clone https://github.com/SpoofIMEI/LiteCanary # Clone the repo
cd LiteCanary # Go into the directory
go build ./cmd/server/server.go # Compile the server (make sure Go is installed)
go build ./cmd/cli/cli.go # Compile the cli
```

## Usage
```
./server & # Starts server 
./cli --url http://host:port/basepath # Open CLI
```

## Configuration
You can configure the server in 2 ways, via a config file called "litecanary.conf" in the same directory as the executable or by using the command line parameters. 

### Config file
The following settings are currently available:
```env
noregistration=<bool> # Disables registration, you will be generated random admin credentials when server is started Default: false

debug=<bool> # Shows debug information Default: false

databaselocation=<string> # SQLite server path. Examples: :memory:, ./test.db  Default: :memory:

listener=<string> # Host:port to listen on. Default: 127.0.0.1:8080

basepath=<string> # HTTP base path for the api. Default: /api/

publickey=<string> # Path to SSL public key. SSL is disabled by default. Default: ""

privatekey=<string> # Path to SSL private key. SSL is disabled by default. Default: ""

log=<string> # Path to log file. Default: "" (disabled)
```

### Command line parameters
```
-base string
      base path for the api (/api/)
-cert string
      public key for the rest api
-database string
      database location (./test.db, :memory:)
-debug
      enables or disables debug information
-key string
      private key for the rest api
-listener string
      listener (127.0.0.1:8080)
-log string
      log file (disabled by default)
-no-req
      disables registration
```

## Other

### Cli help
```
help: displays help page
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
```
