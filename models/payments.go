package models

import (
	"budget2/config"
	"fmt"
	"time"
)

type Payment struct {
	Id              int       `json:"id"`
	Payment_Type_Id int       `json:"payment_type_id"`
	Payment_Date    time.Time `json:"payment_date"`
	Amount          float32   `json:"amount"`
}

type PaymentSummary struct {
	Payment_Type_Id int     `json:"payment_type_id"`
	Amount          float32 `json:"amount"`
}

type MonthlySummary struct {
	Payment_Type_Id int     `json:"payment_type_id"`
	Payment_Date    time.Time `json:"payment_date"`
	Amount          float32 `json:"amount"`
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

	sql := `
	SELECT * FROM payments
	`
	rows, err := db.Query(sql)
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
func InsertPayment(p *Payment) error {

	// Generate our payment date based upon server time:
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
	// Debug:
	fmt.Println("New Record is:", id)
	return nil
}

// Aggregate the payment amounts based upon our pay boundary:
func GetMonthlySummary() ([]*MonthlySummary, error) {

	sql := `
	SELECT 
    	payment_type_id,
    	CASE 
        WHEN date_part('day', payment_date) < %d THEN 
             date_trunc('month', payment_date) + interval '-1month %d days'
        ELSE date_trunc('month', payment_date) + interval '%d days'
    	END AS payment_date,
    	SUM(amount) AS amount
	FROM
    	payments	
	GROUP BY 1,2
	ORDER BY payment_date DESC;
	`
	// Substitute the %d values in sql with the Payday values from our master config structure
	// Subtract one from the value as months start on day 1, not day 0:
	paydayoffset := config.Budget2Config.Payday - 1
	sql = fmt.Sprintf(sql, config.Budget2Config.Payday, paydayoffset, paydayoffset)

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]*MonthlySummary, 0)
	for rows.Next() {
		summary := new(MonthlySummary)
		err := rows.Scan(&summary.Payment_Type_Id, &summary.Payment_Date, &summary.Amount)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return summaries, nil
}

// Aggregate the total payment amount for each payment_type:
func GetPaymentSummary() ([]*PaymentSummary, error) {

	sql := `
	SELECT payment_type_id, SUM(amount) as amount FROM payments GROUP BY payment_type_id
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]*PaymentSummary, 0)
	for rows.Next() {
		summary := new(PaymentSummary)
		err := rows.Scan(&summary.Payment_Type_Id, &summary.Amount)
		if err != nil {
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return summaries, nil
}
