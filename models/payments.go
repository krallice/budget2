package models

import (
	"time"
	"fmt"
)

type Payment struct {
	Id int						`json:"id"`
	Payment_Type_Id int			`json:"payment_type_id"`
	Payment_Date time.Time		`json:"payment_date"`
	Amount float32				`json:"amount"`
}

// Horrible golang date formatting string for YYYY-MM-DD:
const dateFormat string = "2006-01-02"

// Getter and Setter for time.Time object:
func (p *Payment) GetPaymentDateString() string {
	return p.Payment_Date.Format(dateFormat)
}
func (p *Payment) addPaymentDate() {
	p.Payment_Date = time.Now()
}
// Return all Payments from DB:
func AllPayments() ([]*Payment, error) {

	rows, err := db.Query("SELECT * FROM payments")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payments := make([]*Payment, 0)
	for rows.Next() {
		payment := new(Payment)
		err := rows.Scan(&payment.Id, &payment.Payment_Type_Id, &payment.Payment_Date, &payment.Amount)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return payments, nil
}
// Insert pointer to payment into the DB
// It's also this function's responsibility to add the date:
func InsertPayment(p *Payment) (error) {

	p.addPaymentDate()
	sql := `
	INSERT INTO payments (Payment_Type_Id, Payment_Date, Amount)
	VALUES ($1, $2, $3)
	RETURNING id
	`
	id := 0
	err := db.QueryRow(sql, p.Payment_Type_Id, p.GetPaymentDateString(), p.Amount).Scan(&id)
	if err != nil {
		return err
	}
	fmt.Println("New Record is:", id)
	return nil
}
