package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adharapayments/banketh/demobank/data"
	"github.com/adharapayments/banketh/lib/db"
)

type WSmessage struct {
	Success bool
	Data    interface{}
}

type TransferEncoded struct {
	TransferID uint64
	Time       time.Time
	Account    string
	Amount     float64
	Type       string
	Message    string
}

func HandleServices(w http.ResponseWriter, req *http.Request) {
	// ToDo
	req.ParseForm()

	token, ok := req.Form["token"]
	if !ok || token[0] != data.WSTOKEN {
		b, err := json.Marshal(WSmessage{false, "Athentication error"})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return
	}
	command, ok := req.Form["command"]
	if !ok {
		b, err := json.Marshal(WSmessage{false, "No command specified"})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return
	}
	accountA, ok := req.Form["account"]
	if !ok {
		b, err := json.Marshal(WSmessage{false, "Error in readbalance command: need to specify account"})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return
	}
	for _, c := range strings.ToLower(accountA[0]) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		} else {
			b, err := json.Marshal(WSmessage{false, "Error: account can only include leters and digits"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
	}
	var account data.AccountT
	accountExists, err := db.ReadEntry(data.DBNAME, data.DBTABLEACCOUNTS, "AccountID", accountA[0], &account)
	if err != nil {
		log.Printf("Internal DB error when trying to read account balance [%v]\n", err)
		return
	}

	switch command[0] {
	case "readbalance":
		if !accountExists {
			b, err := json.Marshal(WSmessage{false, fmt.Sprintf("Error in readbalance command: account %v not found", accountA[0])})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		b, err := json.Marshal(WSmessage{true, account.Balance})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return
	case "doesaccountexist":
		b, err := json.Marshal(WSmessage{true, accountExists})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return
	case "readtransfers": // only the incoming ones!
		if !accountExists {
			b, err := json.Marshal(WSmessage{false, fmt.Sprintf("Error in readtransfers command: account %v not found", accountA[0])})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		since, sinceSpecified := req.Form["since"]
		var where string
		if sinceSpecified {
			where = fmt.Sprintf("Time >= '%v'", since[0])
			// To do
		}
		limitStr, limitSpecified := req.Form["limit"]
		var limit uint64
		var err error
		if limitSpecified {
			limit, err = strconv.ParseUint(limitStr[0], 0, 64)
			if err != nil {
				b, err := json.Marshal(WSmessage{false, "Error in readtransfers command: bad limit specified"})
				if err != nil {
					log.Printf("Internal error when marshaling a json response [%v]\n", err)
					return
				}
				fmt.Fprint(w, string(b))
				return
			}
		}
		sort := "Time desc"
		if limitSpecified {
			sort += fmt.Sprintf(" limit %v", limit)
		}

		transfers, err := db.ReadTable(data.DBNAME, data.DBTABLETRANSFERS, &data.TransferT{},
			fmt.Sprintf("%v and (Sender='%v' or Receiver='%v')", where, account.AccountID, account.AccountID), sort)
		if err != nil {
			log.Printf("Internal DB error when trying to read transfers [%v]\n", err)
			return
		}

		transfersEncoded := make([]TransferEncoded, 0)
		for _, t := range transfers {
			thisTransfer := t.(*data.TransferT)
			var thisType, accountInTE string
			if thisTransfer.Sender == account.AccountID {
				thisType = "Outbound"
				accountInTE = thisTransfer.Receiver
			} else if thisTransfer.Receiver == account.AccountID {
				thisType = "Inbound"
				accountInTE = thisTransfer.Sender
			} else {
				b, err := json.Marshal(WSmessage{false, "Internal DB error!!"})
				if err != nil {
					log.Printf("Internal error when marshaling a json response [%v]\n", err)
					return
				}
				fmt.Fprint(w, string(b))
				return
			}
			transfersEncoded = append(transfersEncoded,
				TransferEncoded{
					TransferID: thisTransfer.TransferID,
					Time:       time.Time(thisTransfer.Time),
					Account:    accountInTE,
					Amount:     thisTransfer.Amount,
					Type:       thisType,
					Message:    thisTransfer.Message,
				})
		}

		b, err := json.Marshal(WSmessage{true, transfersEncoded})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return

	case "maketransfer":
		data.DBmux.Lock()
		defer data.DBmux.Unlock()
		if !accountExists {
			b, err := json.Marshal(WSmessage{false, fmt.Sprintf("Error in maketransfer command: account %v not found", accountA[0])})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		toaccountA, okAcct := req.Form["toaccount"]
		amountA, okAmnt := req.Form["amount"]
		messageA, okMsg := req.Form["message"]
		if !okAcct || !okAmnt || !okMsg {
			b, err := json.Marshal(WSmessage{false, "Error in maketransfer command: need to specify toaccount, amount and message"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		amount, err := strconv.ParseFloat(amountA[0], 64)
		if err != nil {
			b, err := json.Marshal(WSmessage{false, "Error in maketransfer command: bad amount"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		if amount <= 0 {
			b, err := json.Marshal(WSmessage{false, "Error in maketransfer command: zero or negative amount"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		for _, c := range strings.ToLower(toaccountA[0]) {
			if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
				continue
			} else {
				b, err := json.Marshal(WSmessage{false, "Error in maketransfer command: toaccount can only include leters and digits"})
				if err != nil {
					log.Printf("Internal error when marshaling a json response [%v]\n", err)
					return
				}
				fmt.Fprint(w, string(b))
				return
			}
		}
		if len(messageA[0]) > data.MAXMESSAGELENGTH {
			b, err := json.Marshal(WSmessage{false, "Error in maketransfer command: message can only contain up to 140 characters"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}

		tables := make([]string, 0)
		entries := make([]db.Converter, 0)

		newTransfer := data.TransferT{
			Time:     db.MyTime(time.Now()),
			Sender:   account.AccountID,
			Receiver: toaccountA[0],
			Amount:   amount,
			Message:  messageA[0],
		}

		tables = append(tables, data.DBTABLETRANSFERS)
		entries = append(entries, newTransfer)

		var toAccount data.AccountT
		found, err := db.ReadEntry(data.DBNAME, data.DBTABLEACCOUNTS, "accountID", toaccountA[0], &toAccount)
		if err != nil {
			b, err := json.Marshal(WSmessage{false, "Internal database err"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		if !found {
			b, err := json.Marshal(WSmessage{false, "Receiving account not found"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		if account.AccountID == toAccount.AccountID {
			b, err := json.Marshal(WSmessage{false, "Sending and receiving account are the same"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}

		toAccount.Balance += amount
		tables = append(tables, data.DBTABLEACCOUNTS)
		entries = append(entries, toAccount)

		account.Balance -= amount
		tables = append(tables, data.DBTABLEACCOUNTS)
		entries = append(entries, account)

		err = db.AtomicWrite(data.DBNAME, tables, entries)
		if err != nil {
			b, err := json.Marshal(WSmessage{false, "Internal database err"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		_, values := newTransfer.Convert()
		found, kName, kVal, err := db.FindKey(data.DBNAME, data.DBTABLETRANSFERS, values, &data.TransferT{})
		if err != nil {
			b, err := json.Marshal(WSmessage{false, "Internal database err"})
			if err != nil {
				log.Printf("Internal error when marshaling a json response [%v]\n", err)
				return
			}
			fmt.Fprint(w, string(b))
			return
		}
		b, err := json.Marshal(WSmessage{true, map[string]uint64{kName: kVal.(uint64)}})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return
	default:
		b, err := json.Marshal(WSmessage{false, "Unknown command"})
		if err != nil {
			log.Printf("Internal error when marshaling a json response [%v]\n", err)
			return
		}
		fmt.Fprint(w, string(b))
		return
	}
}
