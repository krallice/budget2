package main

import (
	"budget2/models"
	"fmt"
	"net/http"
	"encoding/json"
	"html/template"
)

func main() {

	models.InitDB("postgres://postgres:password1@localhost/budget2")

	// Main Index:
	http.HandleFunc("/", getIndex)

	// AJAX Functions:
	http.HandleFunc("/ajax/payment_types", ajaxPaymentTypes)
	http.HandleFunc("/ajax/payments", ajaxPayments)
	http.HandleFunc("/ajax/monthlysummary", ajaxMonthlySummary)

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
// Returns all Payment_Types in DB as a JSON object:
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
	sums, err := models.MonthlySummary()
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
		http.Error(w, http.StatusText(405), http.StatusMethodNotAllowed);
		return
	}
}
