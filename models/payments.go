package models

import (
	"budget2/config"
	"fmt"
	"time"
	"strconv"
)

// An individual payment:
type Payment struct {
	Id              int       `json:"id"`
	PaymentTypeId	int       `json:"payment_type_id"`
	PaymentDate		time.Time `json:"payment_date"`
	Amount          float32   `json:"amount"`
}

// Total amount paid per PaymentType bucket:
type PaymentSummary struct {
	PaymentTypeId	int     `json:"payment_type_id"`
	Amount          float32 `json:"amount"`
}

type MonthlySummary struct {
	PaymentTypeId	int     `json:"payment_type_id"`
	PaymentDate		time.Time `json:"payment_date"`
	Amount          float32 `json:"amount"`
}

// Parent holder object for all payment summary details
// This provides the high level details for overview
type BudgetSummary struct {
	RentPaid		bool					`json:"rentpaid"`
	Limit			float32					`json:"limit"`
	TotalLocked		float32					`json:"totallocked"`
	Totals			[]*PaymentSummary		`json:"totals"`
}

// Horrible golang date formatting string for YYYY-MM-DD and DD
const dateFormat string = "2006-01-02"
const dayFormat string = "02"

// Getter and Setter for time.Time object:
func (p *Payment) GetPaymentDateString() string {
	return p.PaymentDate.Format(dateFormat)
}
func (p *Payment) addPaymentDate() {
	p.PaymentDate = time.Now()
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
		err := rows.Scan(&payment.Id, &payment.PaymentTypeId, &payment.PaymentDate, &payment.Amount)
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
	err := db.QueryRow(sql, p.PaymentTypeId, p.GetPaymentDateString(), p.Amount).Scan(&id)
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
		err := rows.Scan(&summary.PaymentTypeId, &summary.PaymentDate, &summary.Amount)
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

func GetBudgetSummary() (*BudgetSummary, error) {

	var b BudgetSummary

	// Get Base Summary:
	sql := `
	SELECT SUM(amount) AS amount FROM payments
	`
	row := db.QueryRow(sql)
	err := row.Scan(&b.TotalLocked)
	if err != nil {
		return nil, err
	}

	// Adjust for rent:
	currentday := time.Now()
	cd, err := strconv.Atoi(currentday.Format(dayFormat))
	if err != nil {
		return nil, err
	}
	if cd <= config.Budget2Config.Rentday {
		b.RentPaid = false
		b.Limit = b.TotalLocked + config.Budget2Config.Rentamount
	} else {
		b.RentPaid = true
		b.Limit = b.TotalLocked
	}

	// Get individual payment type summaries:
	summaries, err := GetPaymentSummary()
	if err != nil {
		return nil, err
	}
	b.Totals = summaries

	return &b, nil
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
		err := rows.Scan(&summary.PaymentTypeId, &summary.Amount)
		if err != nil {
			return nil, err
		}

		// Add our adjustment from the InitalValues Config map, IF the key exists:
		if val, ok := config.Budget2Config.InitialValues[summary.PaymentTypeId]; ok {
			summary.Amount = summary.Amount + val
		}
		summaries = append(summaries, summary)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return summaries, nil
}
