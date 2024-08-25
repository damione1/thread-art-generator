ALTER TABLE
    users
ADD
    COLUMN first_name VARCHAR(255) NOT NULL;

ALTER TABLE
    users
ADD
    COLUMN last_name VARCHAR(255);

-- Name is used as ressource name in the api. To avoid confusion, we will rename it to first_name
ALTER TABLE
    users DROP COLUMN name;
