ALTER TABLE users
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS date_of_birth,
    DROP COLUMN IF EXISTS address,
    DROP COLUMN IF EXISTS phone_number;