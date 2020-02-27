package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/adharapayments/banketh/demobank/data"
	"github.com/adharapayments/banketh/lib/db"
	"github.com/adharapayments/banketh/lib/webutil"
)

func HandleIndex(w http.ResponseWriter, req *http.Request) {
	account, amIlogged, err := logged(w, req)
	if err != nil {
		webutil.ShowErrorf(w, req, "Internal error [%v]", err)
		return
	}
	if !amIlogged {
		webutil.Reload(w, req, "/")
		return
	}
	transfers, err := db.ReadTable(data.DBNAME, data.DBTABLETRANSFERS, &data.TransferT{},
		fmt.Sprintf("Sender='%v' or Receiver='%v'", account.AccountID, account.AccountID), "Time desc")
	if err != nil {
		webutil.ShowErrorf(w, req, "DB error reading past transactions [%v]", err)
		return
	}

	type renderTransferT struct {
		Time       db.MyTime
		TransferID uint64
		Type       string
		Account    string
		Amount     float64
		Message    string
	}

	renderTransfers := make([]renderTransferT, 0)
	for _, v := range transfers {
		thisTransfer := v.(*data.TransferT)
		newT := renderTransferT{
			Time:       thisTransfer.Time,
			TransferID: thisTransfer.TransferID,
			Amount:     thisTransfer.Amount,
			Message:    thisTransfer.Message,
		}
		if thisTransfer.Sender == account.AccountID {
			newT.Type = data.TRANSFER_TYPE_OUTBOUND
			newT.Account = thisTransfer.Receiver
		} else {
			newT.Type = data.TRANSFER_TYPE_INBOUND
			newT.Account = thisTransfer.Sender
		}
		renderTransfers = append(renderTransfers, newT)
	}

	passdata := map[string]interface{}{
		"account":   account,
		"transfers": renderTransfers,
		"currency":  data.CURRENCY,
		"biccode":   data.BICCODE,
	}
	webutil.PlaceHeader(w, req)
	templates.ExecuteTemplate(w, "index.html", passdata)
}

func HandleSendTransfer(w http.ResponseWriter, req *http.Request) {
	data.DBmux.Lock()
	defer data.DBmux.Unlock()

	account, amIlogged, err := logged(w, req)
	if err != nil {
		webutil.ShowErrorf(w, req, "Internal error [%v]", err)
		return
	}
	if !amIlogged {
		webutil.Reload(w, req, "/")
		return
	}

	req.ParseForm()
	bicCodeStr, bicSupplied := req.Form["biccode"] // If !bicSuplied that means this is an internal transfer
	toAccountA, okacct := req.Form["toaccount"]
	amountStr, okammt := req.Form["amount"]
	message, _ := req.Form["message"]
	if !okacct || !okammt {
		webutil.PushAlert(w, req, webutil.ALERT_DANGER, "Bad transfer, missing data")
		webutil.Reload(w, req, "/")
		return
	}
	amount, err := strconv.ParseFloat(amountStr[0], 64)
	if err != nil {
		webutil.PushAlertf(w, req, webutil.ALERT_DANGER, "Bad amount (%v)", amountStr[0])
		webutil.Reload(w, req, "/")
		return
	}
	if amount < 0 {
		webutil.PushAlertf(w, req, webutil.ALERT_DANGER, "Negative amount (%v)", amountStr[0])
		webutil.Reload(w, req, "/")
		return
	}
	for _, c := range strings.ToLower(toAccountA[0]) {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			continue
		} else {
			webutil.PushAlertf(w, req, webutil.ALERT_DANGER, "Bad account (%v) - only numbers and digits allowed", toAccountA[0])
			webutil.Reload(w, req, "/")
			return
		}
	}
	if !bicSupplied && account.AccountID == toAccountA[0] {
		webutil.PushAlert(w, req, webutil.ALERT_DANGER, "Sender and receiver accounts are the same")
		webutil.Reload(w, req, "/")
		return
	}
	if amount > account.Balance {
		webutil.PushAlertf(w, req, webutil.ALERT_DANGER, "Excessive amount (bigger than account balance)")
		webutil.Reload(w, req, "/")
		return
	}

	// Now to check whether the transfer is internal or external
	var MsgToPass, ToAcctToPass string
	if bicSupplied {
		MsgToPass = "to:" + bicCodeStr[0] + " acct:" + toAccountA[0] + " msg:" + message[0]
		ToAcctToPass = data.OMNIBUSACCOUNT
	} else {
		MsgToPass = message[0]
		ToAcctToPass = toAccountA[0]
	}

	if len(message[0]) > data.MAXMESSAGELENGTH {
		webutil.PushAlertf(w, req, webutil.ALERT_DANGER, "Message cannot have more than %v characters (%v over limit)", data.MAXMESSAGELENGTH, len(message[0])-data.MAXMESSAGELENGTH)
		webutil.Reload(w, req, "/")
		return
	}

	tables := make([]string, 0)
	entries := make([]db.Converter, 0)

	newTransfer := data.TransferT{
		Time:     db.MyTime(time.Now()),
		Sender:   account.AccountID,
		Receiver: ToAcctToPass,
		Amount:   amount,
		Message:  MsgToPass,
	}

	tables = append(tables, data.DBTABLETRANSFERS)
	entries = append(entries, newTransfer)

	var toAccount data.AccountT
	found, err := db.ReadEntry(data.DBNAME, data.DBTABLEACCOUNTS, "accountID", newTransfer.Receiver, &toAccount)
	if err != nil {
		webutil.ShowErrorf(w, req, "DB error reading accounts table [%v]", err)
		return
	}
	if !found {
		webutil.PushAlertf(w, req, webutil.ALERT_DANGER, "Destination account %v does not exist", toAccount.AccountID)
		webutil.Reload(w, req, "/")
		return
	}
	if account.AccountID == toAccount.AccountID {
		webutil.PushAlertf(w, req, webutil.ALERT_DANGER, "Cannot make transfer to self")
		webutil.Reload(w, req, "/")
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
		webutil.ShowErrorf(w, req, "DB error when attempting to write the new transfer [%v]", err)
		return
	}

	webutil.PushAlertf(w, req, webutil.ALERT_SUCCESS, "Transfer done: %.02f from acct %v to acct %v (%v)", amount, account.AccountID, toAccount.AccountID, message[0])
	webutil.Reload(w, req, "/")
}
