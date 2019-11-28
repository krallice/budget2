package models

type PaymentType struct {
	Id int			`json:"id"`
	Name string		`json:"name"`
}

func AllPaymentTypes() ([]*PaymentType, error) {

	sql := `
	SELECT * FROM payment_types
	`
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pts := make([]*PaymentType, 0)
	for rows.Next() {
		pt := new(PaymentType)
		err := rows.Scan(&pt.Id, &pt.Name)
		if err != nil {
			return nil, err
		}
		pts = append(pts, pt)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return pts, nil
}
