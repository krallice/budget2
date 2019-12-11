package models

import (
	"budget2/config"
	"fmt"
	"time"
	"strconv"
)

// An individual payment record, maps directly to the payment 
// schema in DB
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

// Total amount paid per date, per payment_type_id:
type MonthlySummary struct {
	PaymentTypeId	int     `json:"payment_type_id"`
	PaymentDate		time.Time `json:"payment_date"`
	Amount          float32 `json:"amount"`
}

// Parent holder object for all payment summary details
// This provides the high level details for overview
type BudgetSummary struct {
	RentPaid		bool					`json:"rentpaid"`
	// Limit = RentAmount + TotalLocked
	Limit			float32					`json:"limit"`
	TotalLocked		float32					`json:"totallocked"`
	LockedThisMonth float32					`json:"lockedthismonth"`
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
func (db *DB) AllPayments() ([]*Payment, error) {

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
func (db *DB) InsertPayment(p *Payment) error {

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
func (db *DB) GetMonthlySummary() ([]*MonthlySummary, error) {

	// Old Ineffient Query:
	// sql := `
	// SELECT 
		// payment_type_id,
		// CASE 
        // WHEN date_part('day', payment_date) < %d THEN 
             // date_trunc('month', payment_date) + interval '-1month %d days'
        // ELSE date_trunc('month', payment_date) + interval '%d days'
		// END AS payment_date,
		// SUM(amount) AS amount
	// FROM
		// payments	
	// GROUP BY 1,2
	// ORDER BY payment_date DESC;
	// `
	// paydayoffset := config.Budget2Config.Payday - 1
	// sql = fmt.Sprintf(sql, config.Budget2Config.Payday, paydayoffset, paydayoffset)

	// Substitute the %d values in sql with the Payday values from our master config structure
	// Subtract one from the value as months start on day 1, not day 0:

	// New efficient monthly summary query:
	sql :=
	`
	SELECT 
		payment_type_id,
		date_trunc('month', payment_date - interval '%d day') + interval '%d day' as payment_date,
		SUM(amount)
	FROM payments
	GROUP by 1, 2
	ORDER by 2 DESC;
	`
	paydayoffset := config.Budget2Config.Payday - 1
	sql = fmt.Sprintf(sql, paydayoffset, paydayoffset)

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

func (db *DB) GetBudgetSummary() (*BudgetSummary, error) {

	var b BudgetSummary

	currentday := time.Now()

	// Get Base Summary:
	sql := `
	SELECT SUM(amount) AS amount FROM payments
	`
	row := db.QueryRow(sql)
	err := row.Scan(&b.TotalLocked)
	if err != nil {
		return nil, err
	}
	// Don't forget to add our starting values:
	for _, v := range config.Budget2Config.InitialValues {
		b.TotalLocked = b.TotalLocked + v
	}

	/* Old locked this month query:
	sql = `
	SELECT amount FROM 
		(SELECT 
			payment_type_id,
			CASE 
				WHEN date_part('day', payment_date) < %d THEN 
					date_trunc('month', payment_date) + interval '-1month %d days'
				ELSE date_trunc('month', payment_date) + interval '%d days'
			END AS payment_date,
			SUM(amount) AS amount
		FROM
			payments
		WHERE
			payment_type_id = 1
		GROUP BY 1,2) AS agg_values
	WHERE
		CASE 
			WHEN date_part('day', CURRENT_DATE) < %d THEN 
				payment_date = date_trunc('month', CURRENT_DATE) + interval '-1month %d days'
			ELSE 
				payment_date = date_trunc('month', CURRENT_DATE) + interval '%d days'
		END
	`
	// Substitute the %d values in sql with the Payday values from our master config structure
	// Subtract one from the value as months start on day 1, not day 0:
	paydayoffset := config.Budget2Config.Payday - 1
	sql = fmt.Sprintf(sql, config.Budget2Config.Payday, paydayoffset, paydayoffset, config.Budget2Config.Payday, paydayoffset, paydayoffset)
	*/

	// New locked this month query:
	sql = `
	SELECT amount FROM 
		( SELECT 
			payment_type_id,
			date_trunc('month', payment_date - interval '%d day') + interval '%d day' as payment_date,
			SUM(amount) as amount
		FROM payments
		WHERE payment_type_id = 1
		GROUP by 1, 2
		) AS agg_values
	WHERE
		payment_date = date_trunc('month', CURRENT_DATE - interval '%d day') + interval '%d day';
	`
	paydayoffset := config.Budget2Config.Payday - 1
	sql = fmt.Sprintf(sql, paydayoffset, paydayoffset, paydayoffset, paydayoffset)
	row = db.QueryRow(sql)
	err = row.Scan(&b.LockedThisMonth)
	if err != nil {
		return nil, err
	}

	// Adjust for rent:
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
	summaries, err := db.GetPaymentSummary()
	if err != nil {
		return nil, err
	}
	b.Totals = summaries

	return &b, nil
}

// Aggregate the total payment amount for each payment_type:
func (db *DB) GetPaymentSummary() ([]*PaymentSummary, error) {

	sql := `
	SELECT payment_type_id, SUM(amount) as amount FROM payments GROUP BY payment_type_id ORDER BY payment_type_id ASC
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

// Get our last 6 month recent house history figures:
func (db *DB) GetRecentHouseHistory() ([]*MonthlySummary, error) {

	sql := `
	SELECT
			payment_date, COALESCE(amount, 0)
	FROM
		generate_series(
			(date_trunc('month', CURRENT_DATE - interval '%d day') - interval '5 month' + interval '%d day')::timestamp, 
			(date_trunc('month', CURRENT_DATE - interval '%d day') + interval '%d day')::timestamp, 
			interval '1 month') AS payment_date
	LEFT JOIN
	(
		SELECT 
			date_trunc('month', payment_date - interval '%d day') + interval '%d day' as month_15,
			SUM(amount) AS amount
		FROM payments
		WHERE payment_type_id = 1
		GROUP BY payment_type_id, month_15
		ORDER BY payment_type_id, month_15
	) 
	AS y ON payment_date=y.month_15
	ORDER BY payment_date DESC;
	`
	paydayoffset := config.Budget2Config.Payday - 1
	sql = fmt.Sprintf(sql, paydayoffset, paydayoffset, paydayoffset, paydayoffset, paydayoffset, paydayoffset)

	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]*MonthlySummary, 0)
	for rows.Next() {
		summary := new(MonthlySummary)
		err := rows.Scan(&summary.PaymentDate, &summary.Amount)
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
