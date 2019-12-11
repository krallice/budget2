package main

import (
	"budget2/config"
	"budget2/models"
	"fmt"
	"log"
	// "log/syslog"
	"encoding/json"
	"html/template"
	"net/http"

	"net/smtp"
	"strings"
)

func main() {

	/* Disable write to syslog
	logwriter, err := syslog.New(syslog.LOG_NOTICE, "budget2")
	if err != nil {
		log.Print("Unable to start syslog, exiting ...")
		return
	}
	log.SetOutput(logwriter)
	*/

	log.Print("Budget2 Daemon Starting ...")

	// Attempt to read our YAML config file, and bomb out if this fails:
	log.Print("Reading YAML config file")
	err := config.ReadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Print("Connecting to postgres DB")
	models.InitDB("postgres://postgres:password1@localhost/budget2")

	// Main Index:
	http.HandleFunc("/", getIndex)

	// AJAX Functions:
	http.HandleFunc("/api/v1/budgetsummary", ajaxBudgetSummary)
	http.HandleFunc("/api/v1/paymenttypes", ajaxPaymentTypes)
	http.HandleFunc("/api/v1/recenthousehistory", ajaxRecentHouseHistory)

	http.HandleFunc("/api/v1/payments", ajaxPayments)

	// Future Resource Serving:
	// http.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("./res"))))

	log.Print("Webserver UP")
	http.ListenAndServe(":3000", nil)
}

// Basic Placeholder Index Page:
func getIndex(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	tmpl.Execute(w, nil)
}

// Returns our master BudgetSummary Struct as JSON:
func ajaxBudgetSummary(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	sum, err := models.GetBudgetSummary()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	jreply, err := json.Marshal(sum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jreply)
}

// Returns all PaymentTypes in DB as a JSON object:
func ajaxPaymentTypes(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	pts, err := models.AllPaymentTypes()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	jreply, err := json.Marshal(pts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jreply)
}

// Either Get or Set our Payment(s):
func ajaxPayments(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// Return all Payments:
	case "GET":
		pys, err := models.AllPayments()
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}
		jreply, err := json.Marshal(pys)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jreply)
	// Insert a new Payment:
	case "POST":
		var p models.Payment

		err := json.NewDecoder(r.Body).Decode(&p)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = models.InsertPayment(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// We successfully made a payment, time to email:
		generateEmail(&p)
		return

	default:
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}
}

// Returns a summary of recent house payments:
func ajaxRecentHouseHistory(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	sums, err := models.GetRecentHouseHistory()
	if err != nil {
		fmt.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	jreply, err := json.Marshal(sums)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jreply)
}

// Generate email notifying of payment:
func generateEmail(p *models.Payment) {

	auth := smtp.PlainAuth("", config.Budget2Config.SenderAddress, "", "127.0.0.1")
	toHeader := strings.Join(config.Budget2Config.EmailRecipients, ",")

	// Get our Payment Type:
	pt, err := models.GetPaymentTypeById(p.PaymentTypeId)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get our complete summary for some quick stats:
	sum, err := models.GetBudgetSummary()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Prepare message:
	msg := []byte("To: " + toHeader + "\r\n" +
	"Subject: Budget2 Payment Made\r\n" +
	"\r\n" +
	"A Payment of $" + fmt.Sprintf("%.2f", p.Amount) + " has just been made into the " + pt.Name + " fund \r\n" +
	"\r\n" +
	"Total Locked     : $" + fmt.Sprintf("%.2f", sum.TotalLocked) + "\r\n" +
	"Locked this Month: $" + fmt.Sprintf("%.2f", sum.LockedThisMonth) + "\r\n")

	// Send email:
	err = smtp.SendMail("127.0.0.1:25", auth, config.Budget2Config.SenderAddress, config.Budget2Config.EmailRecipients, msg)
	if err != nil {
		fmt.Println(err)
	}

	// Print String:
	// s := string(msg)
	// fmt.Println(s)
	return
}

