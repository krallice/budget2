package main

import (
	"budget2/models"
	"fmt"
	"net/http"
)

func main() {

	models.InitDB("postgres://postgres:password1@localhost/budget2")

	// payment_type specific handlers:
	http.HandleFunc("/plain_payment_types", getPaymentTypesPlain)

	// payments specific handlers:
	http.HandleFunc("/plain_payments", getPaymentsPlain)

	http.ListenAndServe(":3000", nil)
}

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
		fmt.Fprintf(w, "%d, %d, %f,\n", payment.Id, payment.Payment_Type_Id, payment.Amount)
	}
}
