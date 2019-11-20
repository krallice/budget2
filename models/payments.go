package models

type Payment struct {
	Id int					`json:"id"`
	Payment_Type_Id int		`json:"payment_type_id"`
	Payment_Date string		`json:"payment_date"`
	Amount float32			`json:"amount"`
}

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
