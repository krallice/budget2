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

// Master environment:
type Env struct {
	db models.Datastore
}

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
	connection := fmt.Sprintf("postgres://%s:%s@%s/%s", config.Budget2Config.DBUsername, config.Budget2Config.DBPassword,
		config.Budget2Config.DBServer, config.Budget2Config.DBName)
	db, err := models.InitDB(connection)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Init our env variable:
	env := &Env{db}

	// Main Index:
	http.HandleFunc("/", env.getIndex)

	// AJAX Functions:
	http.HandleFunc("/api/v1/budgetsummary", env.ajaxBudgetSummary)
	http.HandleFunc("/api/v1/paymenttypes", env.ajaxPaymentTypes)
	http.HandleFunc("/api/v1/recenthousehistory", env.ajaxRecentHouseHistory)

	http.HandleFunc("/api/v1/payments", env.ajaxPayments)

	// Future Resource Serving:
	// http.Handle("/res/", http.StripPrefix("/res/", http.FileServer(http.Dir("./res"))))

	log.Print("Webserver UP")
	http.ListenAndServe(":3000", nil)
}

// Basic Placeholder Index Page:
func (env *Env) getIndex(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	tmpl := template.Must(template.ParseFiles("./templates/index.html"))
	tmpl.Execute(w, nil)
}

// Returns our master BudgetSummary Struct as JSON:
func (env *Env) ajaxBudgetSummary(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	sum, err := env.db.GetBudgetSummary()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func (env *Env) ajaxPaymentTypes(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	pts, err := env.db.AllPaymentTypes()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func (env *Env) ajaxPayments(w http.ResponseWriter, r *http.Request) {

	switch r.Method {
	// Return all Payments:
	case "GET":
		pys, err := env.db.AllPayments()
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
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Do not allow withdrawals via the front end:
		if p.Amount < 1 {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		err = env.db.InsertPayment(&p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		// We successfully made a payment, time to email:
		env.generateEmail(&p)
		return

	default:
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
}

// Returns a summary of recent house payments:
func (env *Env) ajaxRecentHouseHistory(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	sums, err := env.db.GetRecentHouseHistory()
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
func (env *Env) generateEmail(p *models.Payment) {

	auth := smtp.PlainAuth("", config.Budget2Config.SenderAddress, "", "127.0.0.1")
	toHeader := strings.Join(config.Budget2Config.EmailRecipients, ",")

	// Get our Payment Type:
	pt, err := env.db.GetPaymentTypeById(p.PaymentTypeId)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Get our complete summary for some quick stats:
	sum, err := env.db.GetBudgetSummary()
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

