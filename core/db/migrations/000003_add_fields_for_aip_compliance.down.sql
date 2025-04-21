ALTER TABLE
    users
ADD
    COLUMN name VARCHAR(255) NOT NULL;

-- Since we renamed the column to first_name in the up migration, we need to restore the original name
ALTER TABLE
    users RENAME COLUMN first_name TO name;

ALTER TABLE
    users DROP COLUMN last_name;
