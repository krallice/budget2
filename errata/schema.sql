CREATE DATABASE budget2;

\c budget2;

CREATE TABLE payment_types (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL
);

CREATE TABLE payments (
	id SERIAL PRIMARY KEY,
	payment_type_id INTEGER REFERENCES payment_types(id),
	payment_date DATE NOT NULL,
	amount NUMERIC(12,2) NOT NULL
);

CREATE TABLE payment_goals (
	id SERIAL PRIMARY KEY,
	aggregate BOOLEAN,
	payment_type_id INTEGER REFERENCES payment_types(id),
	goal_date DATE NOT NULL,
	amount NUMERIC(12,2) NOT NULL
);
