package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/gorilla/sessions"
	"github.com/adharapayments/banketh/demobank/data"
	"github.com/adharapayments/banketh/lib/db"
	"github.com/adharapayments/banketh/lib/webutil"
)

func HandleLogin(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	loginA, okl := req.Form["login"]
	passwordA, okp := req.Form["password"]
	if !okl || !okp {
		http.Error(w, errors.New("Form error").Error(), http.StatusInternalServerError)
		return
	}
	session, err := webutil.Store.Get(req, SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var thisAccount data.AccountT
	found, _, _, err := db.FindKey(data.DBNAME, data.DBTABLEACCOUNTS, []db.DataElementT{db.DataElementT{Name: "Login", Val: loginA[0]}}, &thisAccount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !found {
		webutil.PushAlert(w, req, webutil.ALERT_DANGER, "Login error")
		webutil.Reload(w, req, "/")
		return
	}
	if passwordA[0] != thisAccount.Password {
		webutil.PushAlert(w, req, webutil.ALERT_DANGER, "Login error")
		webutil.Reload(w, req, "/")
		return
	}
	session.Values["accountID"] = thisAccount.AccountID
	session.Values["login"] = thisAccount.Login
	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 1,
		HttpOnly: true,
	}
	session.Save(req, w)
	webutil.Reload(w, req, "/")
}

func HandleLogout(w http.ResponseWriter, req *http.Request) {
	session, err := webutil.Store.Get(req, webutil.SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for i := range session.Values {
		delete(session.Values, i)
	}
	session.Save(req, w)
	webutil.Reload(w, req, "/")
}

func HandleRetrievePassword(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	emailA := req.Form["email"]
	found, keyName, keyVal, err := db.FindKey(data.DBNAME, data.DBTABLEACCOUNTS, []db.DataElementT{db.DataElementT{Name: "Email", Val: emailA[0]}}, &data.AccountT{})
	if err != nil {
		webutil.ShowErrorf(w, req, "DB Error [%v]", err)
		return
	}
	if found && keyName != "" && keyVal != nil {
		account := data.NewAccountT()
		_, err = db.ReadEntry(data.DBNAME, data.DBTABLEACCOUNTS, keyName, keyVal, &account)
		if err != nil {
			errMsg := fmt.Sprintf("DB error [%v]", err)
			db.RegisterEvent(data.DBNAME, data.DBTABLEEVENTS, db.EVENT_DB_ERR, errMsg)
			webutil.ShowError(w, req, errMsg)
			return
		}
		msg := "From: " + data.EMAILACCOUNT + "\nTo: " + account.Email + "\nSubject: " + HEADER_PAGE_TITLE + " service\n\n" +
			"Hello from " + HEADER_PAGE_TITLE + ". You (or someone) asked to retrieve your password, so here it is:\n" +
			account.Password + "\n\nSo long!\n"

		err = smtp.SendMail("smtp.gmail.com:587",
			//err = smtp.SendMail("smtp-mail.outlook.com:587",
			smtp.PlainAuth("", data.EMAILACCOUNT, data.EMAILPASSWORD, "smtp.gmail.com"),
			//smtp.PlainAuth("", data.EMAILACCOUNT, data.EMAILPASSWORD, "smtp-mail.outlook.com"),
			data.EMAILACCOUNT, []string{account.Email}, []byte(msg))

		if err != nil {
			webutil.ShowErrorf(w, req, "smtp error: %v", err)
			return
		}
		db.RegisterEventf(data.DBNAME, data.DBTABLEEVENTS, db.EVENT_INFO,
			"User %v (Account ID %v) retrieved his/her password, it was sent to %v", account.Login, account.AccountID, account.Email)

	} else {
		db.RegisterEventf(data.DBNAME, data.DBTABLEEVENTS, db.EVENT_INFO,
			"Someone tried to retrieve a password for a non-existent user with email %v", emailA[0])
	}
	webutil.PushAlert(w, req, webutil.ALERT_SUCCESS, "A message with the password was sent to the appropriate email (if the user was found)")
	webutil.Reload(w, req, "/")
	return
}

func HandleChangePassword(w http.ResponseWriter, req *http.Request) {
	account, amIlogged, err := logged(w, req)
	if err != nil {
		webutil.ShowErrorf(w, req, "Internal error [%v]", err)
		return
	}
	if !amIlogged {
		webutil.Reload(w, req, "/")
		return
	}

	data.DBmux.Lock()
	defer data.DBmux.Unlock()

	req.ParseForm()
	newpasswordA, okpass := req.Form["newpassword"]
	passwordagainA, okpassagain := req.Form["passwordagain"]
	if !okpass || !okpassagain {
		webutil.PushAlert(w, req, webutil.ALERT_DANGER, "Bad call, missing data")
		webutil.Reload(w, req, "/")
		return
	}
	if newpasswordA[0] != passwordagainA[0] {
		webutil.PushAlert(w, req, webutil.ALERT_DANGER, "passwords do not match")
		webutil.Reload(w, req, "/")
		return
	}
	var entry data.AccountT
	found, err := db.ReadEntry(data.DBNAME, data.DBTABLEACCOUNTS, "Login", account.Login, &entry)
	if err != nil || !found {
		webutil.ShowErrorf(w, req, "Error reading accounts database [%v]\n", err)
		return
	}
	entry.Password = newpasswordA[0]
	err = db.WriteEntry(data.DBNAME, data.DBTABLEACCOUNTS, entry)
	if err != nil {
		webutil.ShowErrorf(w, req, "Error writing the new password to the database [%v]\n", err)
		return
	}
	webutil.PushAlert(w, req, webutil.ALERT_SUCCESS, "Password hapily updated")
	webutil.Reload(w, req, "/")
}
