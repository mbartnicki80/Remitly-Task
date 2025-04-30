CREATE TABLE IF NOT EXISTS swift_codes (
    country_iso2_code CHAR(2) NOT NULL,
    swift_code VARCHAR(11) PRIMARY KEY,
    code_type CHAR(5) NOT NULL,
    bank_name TEXT NOT NULL,
    address TEXT,
    town_name TEXT NOT NULL,
    country_name TEXT NOT NULL,
    time_zone TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_country_iso2  ON swift_codes(country_iso2_code);
