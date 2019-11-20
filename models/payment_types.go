package models

type Payment_Type struct {
	Id int			`json:"id"`
	Name string		`json:"name"`
}

func AllPaymentTypes() ([]*Payment_Type, error) {

	rows, err := db.Query("SELECT * FROM payment_types")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	payment_types := make([]*Payment_Type, 0)
	for rows.Next() {
		payment_type := new(Payment_Type)
		err := rows.Scan(&payment_type.Id, &payment_type.Name)
		if err != nil {
			return nil, err
		}
		payment_types = append(payment_types, payment_type)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return payment_types, nil
}
