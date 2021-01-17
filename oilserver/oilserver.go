package main

import (
	//"bufio"
	"fmt"
	"log"
	"os"

	//"strings"
	"server"
	"time"

	"github.com/adharapayments/banketh/demobank/data"
	"github.com/adharapayments/banketh/demobank/server"
	"github.com/adharapayments/banketh/lib/db"
)

///////////////////////////////////////////////////
// Main function
///////////////////////////////////////////////////
func main() {

	args := os.Args
	if len(args) != 9 {
		fmt.Printf(
			`Error calling %s
Usage: %s <port>
Example: %s 8050`,
			args[0],
			args[0],
			args[0],
		)
		return
	}
	server.HEADER_PAGE_TITLE = args[1]
	data.CURRENCY = args[2]
	data.DBNAME = args[3]
	server.WEBPORT = args[4]
	data.WSTOKEN = args[5]
	data.OMNIBUSACCOUNT = args[6]
	data.PICSPREFIX = args[7]
	data.BICCODE = args[8]
	log.Printf("Initializing demobank with bank name='%v', currency='%v', DBNAME='%v', web port='%v', token='%v', omnibus='%v', picsprefix='%v'",
		server.HEADER_PAGE_TITLE, data.CURRENCY, data.DBNAME, server.WEBPORT, data.WSTOKEN, data.OMNIBUSACCOUNT, data.PICSPREFIX)

	db.CheckAndCreateDB(data.DBNAME)

	if err := db.InitTable(data.DBNAME, data.DBTABLETRANSFERS, data.TransferT{}, nil); err != nil {
		log.Fatal("Error initializing DB: ", err)
	}
	if err := db.InitTable(data.DBNAME, data.DBTABLEACCOUNTS, data.AccountT{}, nil); err != nil {
		log.Fatal("Error initializing DB: ", err)
	}
	if err := db.InitTable(data.DBNAME, data.DBTABLEEVENTS, db.EventT{}, nil); err != nil {
		log.Fatal("Error initializing DB: ", err)
	}

	server.StartWeb()

	for {
		time.Sleep(10 * time.Second)
	}

	/*
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Demobank console: ")
			s, _ := reader.ReadString('\n')
			command := strings.Fields(s)
			if len(command) >= 1 {
				switch command[0] {
				case "exit":
					fmt.Println("Have a nice day!")
					os.Exit(0)
				case "showBalance":
					// To do
					fmt.Printf("Not implemented yet, stay tuned\n")
				case "showTransfer":
					// To do
					fmt.Printf("Not implemented yet, stay tuned\n")
				default:
					fmt.Printf("Unknown command %v\n", command)
				}
			}
		}
	*/
}
