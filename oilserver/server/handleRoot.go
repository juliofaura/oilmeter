package server

import (
	"net/http"

	"github.com/adharapayments/banketh/lib/webutil"
)

///////////////////////////////////////////////////
// Handle functions
///////////////////////////////////////////////////
func HandleRoot(w http.ResponseWriter, req *http.Request) {
	if webutil.Logged(w, req) {
		webutil.Reload(w, req, "/index")
		return
	}
	passdata := map[string]interface{}{
		"pagetitle":  HEADER_PAGE_TITLE + " - log in",
		"loginerror": false,
		"alerts":     webutil.PopAlerts(w, req),
	}
	session, err := webutil.Store.Get(req, webutil.SESSIONNAME)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	loginerror, exists := session.Values["loginerror"]
	if exists && loginerror.(bool) {
		passdata["loginerror"] = true
	}
	templates.ExecuteTemplate(w, "login.html", passdata)
}
