ALTER TABLE users
    ADD COLUMN image_url VARCHAR(255),
    ADD COLUMN date_of_birth DATE,
    ADD COLUMN address VARCHAR(255),
    ADD COLUMN phone_number VARCHAR(20);