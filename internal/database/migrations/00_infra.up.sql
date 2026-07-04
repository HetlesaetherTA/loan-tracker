DO $$
BEGIN
  IF NOT EXISTS(
    SELECT
    FROM pg_catalog.pg_roles
    WHERE rolname = 'app_user_loan_tracker') THEN
  CREATE USER app_user_loan_tracker WITH PASSWORD 'thisisatemporarypassword';
END IF;
END
$$;

ALTER USER app_user_loan_tracker NOSUPERUSER NOCREATEDB NOCREATEROLE;

CREATE SCHEMA IF NOT EXISTS app_loan_tracker;

