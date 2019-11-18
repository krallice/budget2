package main

import (
	"budget2/models"
	"fmt"
	"net/http"
)

func main() {

	models.InitDB("postgres://postgres:password1@localhost/budget2")

	http.HandleFunc("/payment_types", getPaymentTypes)
	http.ListenAndServe(":3000", nil)
}

func getPaymentTypes(w http.ResponseWriter, r *http.Request) {

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
