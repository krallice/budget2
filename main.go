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
	http.HandleFunc("/ajax/budgetsummary", ajaxBudgetSummary)

	http.HandleFunc("/ajax/payments", ajaxPayments)
	http.HandleFunc("/ajax/paymentsummary", ajaxPaymentSummary)
	http.HandleFunc("/ajax/monthlysummary", ajaxMonthlySummary)

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

// Returns all PaymentTypes in DB as a JSON object:
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

// Returns a list of aggregated monthly summaries of Payments
func ajaxMonthlySummary(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	sums, err := models.GetMonthlySummary()
	if err != nil {
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

// Returns a total summed amount of payments:
func ajaxPaymentSummary(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	sums, err := models.GetPaymentSummary()
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
	default:
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed)
		return
	}
}
