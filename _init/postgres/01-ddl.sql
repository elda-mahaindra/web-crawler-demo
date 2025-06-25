-- Schema definitions

CREATE SCHEMA IF NOT EXISTS "ibdwh";

-- Table definitions

CREATE TABLE ibdwh.emas (
	emas_id SERIAL PRIMARY KEY,
	jual numeric NULL,
	beli numeric NULL,
	created_at timestamp NULL,
	avg_bpkh numeric NULL
);