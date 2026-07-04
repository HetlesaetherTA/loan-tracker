CREATE TYPE app_loan_tracker.loan_status_enum AS ENUM(
  'PENDING',
  'ACTIVE',
  'PAID_OFF',
  'DEFAULTED'
);

CREATE TYPE app_loan_tracker.frequency_type_enum AS ENUM(
  'WEEKLY',
  'BI_WEEKLY',
  'MONTHLY',
  'QUARTERLY',
  'YEARLY'
);

CREATE TABLE app_loan_tracker.loans(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL,
  original_principal numeric(16, 4) NOT NULL,
  principal numeric(16, 4) NOT NULL,
  yearly_interest numeric(6, 5) NOT NULL,
  interest_calculated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  due_at timestamp with time zone NOT NULL,
  payment_frequency app_loan_tracker.frequency_type_enum NOT NULL,
  next_payment_at timestamp with time zone NOT NULL,
  updated_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  current_version int NOT NULL DEFAULT 1,
  status app_loan_tracker.loan_status_enum NOT NULL DEFAULT 'PENDING',
  description text,
  metadata jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE app_loan_tracker.ledger(
  transction_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  loan_id uuid NOT NULL,
  sender_user_id uuid NOT NULL,
  amount numeric(16, 4) NOT NULL,
  loan_version int NOT NULL,
  created_at timestamp with time zone NOT NULL DEFAULT CURRENT_TIMESTAMP,
  description text,
  metadata jsonb NOT NULL DEFAULT '{}',
  CONSTRAINT fk_ledger_loan FOREIGN KEY (loan_id) REFERENCES app_loan_tracker.loans(id) ON DELETE CASCADE
);

-- Views:
CREATE VIEW app_loan_tracker.user_loan_view AS
SELECT id,
  original_principal,
  principal,
  yearly_interest,
  interest_calculated_at,
  due_at,
  payment_frequency,
  next_payment_at,
  updated_at,
  current_version,
  status,
  description
FROM app_loan_tracker.loans;

CREATE VIEW app_loan_tracker.user_ledger_view AS
SELECT transction_id,
  loan_id,
  sender_user_id,
  amount,
  loan_version,
  created_at,
  description
FROM app_loan_tracker.ledger;

-- Indexis:
CREATE INDEX idx_loans_user_id ON app_loan_tracker.loans(user_id);

CREATE INDEX idx_ledger_loan_id ON app_loan_tracker.ledger(loan_id);

-- access:
GRANT USAGE ON SCHEMA app_loan_tracker TO app_user_loan_tracker;

GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA app_loan_tracker TO app_user_loan_tracker;

ALTER DEFAULT PRIVILEGES IN SCHEMA app_loan_tracker GRANT
SELECT, INSERT, UPDATE, DELETE ON TABLES TO app_user_loan_tracker;

