-- Add uuid column as varchar
ALTER TABLE users ADD COLUMN uuid VARCHAR(36) AFTER id;

-- Populate existing records with UUID()
SET SQL_SAFE_UPDATES = 0;
UPDATE users SET uuid = UUID() WHERE uuid IS NULL;
SET SQL_SAFE_UPDATES = 1;

-- Make uuid NOT NULL and UNIQUE
ALTER TABLE users MODIFY COLUMN uuid VARCHAR(36) NOT NULL, ADD UNIQUE INDEX idx_users_uuid (uuid);
