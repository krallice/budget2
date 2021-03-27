package models

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

type DB struct {
	*sql.DB
}

type Datastore interface {
	AllPayments() ([]*Payment, error)

	AllPaymentTypes() ([]*PaymentType, error)
	GetPaymentTypeById(i int) (*PaymentType, error)

	InsertPayment(p *Payment) error
	GetMonthlySummary() ([]*MonthlySummary, error)
	GetBudgetSummary() (*BudgetSummary, error)
	GetPaymentSummary() ([]*PaymentSummary, error)
	GetRecentHouseHistory() ([]*MonthlySummary, error)

	GetPaymentGoals() ([]*PaymentGoal, error)
}

func InitDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Panic(err)
		return  nil, err
	}

	if err = db.Ping(); err != nil {
		log.Panic(err)
		return  nil, err
	}
	return &DB{db}, nil
}
