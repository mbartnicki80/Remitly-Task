CREATE TABLE IF NOT EXISTS swift_codes (
    country_iso2_code CHAR(2) NOT NULL,
    swift_code VARCHAR(11) PRIMARY KEY,
    bank_name TEXT NOT NULL,
    address TEXT,
    town_name TEXT NOT NULL,
    country_name TEXT NOT NULL,
    is_headquarter BOOLEAN NOT NULL
);

CREATE TABLE IF NOT EXISTS branches (
    swift_code VARCHAR(11) PRIMARY KEY,
    headquarter VARCHAR(11) NOT NULL,
    FOREIGN KEY (swift_code) REFERENCES swift_codes(swift_code) ON DELETE CASCADE,
    FOREIGN KEY (headquarter) REFERENCES swift_codes(swift_code) ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS idx_country_iso2  ON swift_codes(country_iso2_code);
