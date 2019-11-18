CREATE TABLE payment_types (
	id SERIAL PRIMARY KEY,
	name VARCHAR(255) NOT NULL
);

CREATE TABLE payments (
	id SERIAL PRIMARY KEY,
	payment_type_id INTEGER REFERENCES payment_types(id),
	name NUMERIC(12,2) NOT NULL
);
