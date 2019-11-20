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

	http.HandleFunc("/ajax/payment_types", ajaxPaymentTypes)

	// Plain function handlers:
	http.HandleFunc("/plain_payment_types", getPaymentTypesPlain)
	http.HandleFunc("/plain_payments", getPaymentsPlain)
	http.HandleFunc("/plain_add_payment", addPaymentPlain)

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

// Todo: Implement actual persist-to-db functionality:
func addPaymentPlain(w http.ResponseWriter, r *http.Request) {

	var p models.Payment

	// Only allow POSTs:
	if r.Method != "POST" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	fmt.Fprintf(w, "%.2f\n", p.Amount)
}

// Undecorated/plain diagnostic functions below:
// (To be removed at a later stage)

// Diagnostic testing/building function to display table contents:
func getPaymentTypesPlain(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	payment_types, err := models.AllPaymentTypes()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	for _, payment_type := range payment_types {
		fmt.Fprintf(w, "%d, %s\n", payment_type.Id, payment_type.Name)
	}
}

// Get a plain dump of all payments:
func getPaymentsPlain(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, http.StatusText(405), 405)
		return
	}
	payments, err := models.AllPayments()
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}
	for _, payment := range payments {
		fmt.Fprintf(w, "%d, %d, %s, %.2f,\n", payment.Id, payment.Payment_Type_Id, payment.Payment_Date, payment.Amount)
	}
}
