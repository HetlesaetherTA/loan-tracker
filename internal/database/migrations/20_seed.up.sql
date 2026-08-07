-- AI generated seed for testing
SET search_path TO app_loan_tracker;

INSERT INTO loans(id, user_id, currency, original_principal, principal, yearly_interest, due_at, payment_frequency, next_payment_at, status, description)
  VALUES ('11111111-1111-1111-1111-111111111111', '00000000-0000-0000-0000-000000000000', '€', 10000.0000, 10000.0000, 0.05500, -- 5.5% interest
    CURRENT_TIMESTAMP + INTERVAL '1 year', 'MONTHLY', CURRENT_TIMESTAMP + INTERVAL '1 month', 'PENDING', 'Initial pending small business loan request');

INSERT INTO loans(id, user_id, currency, original_principal, principal, yearly_interest, due_at, payment_frequency, next_payment_at, current_version, status, description)
  VALUES ('22222222-2222-2222-2222-222222222222', '00000000-0000-0000-0000-000000000000', '€', 5000.0000, 4200.0000, 0.08250, -- 8.25% interest, $4,200 remaining balance
    CURRENT_TIMESTAMP + INTERVAL '6 months', 'BI_WEEKLY', CURRENT_TIMESTAMP + INTERVAL '2 weeks', 3, -- Reflects that updates/payments have happened
    'ACTIVE', 'Active personal loan for home improvements');

INSERT INTO loans(id, user_id, currency, original_principal, principal, yearly_interest, due_at, payment_frequency, next_payment_at, current_version, status, description)
  VALUES ('33333333-3333-3333-3333-333333333333', '00000000-0000-0000-0000-000000000000', 'NOK', 1200.0000, 0.0000, 0.12000, -- 12% interest, $0 remaining balance
    CURRENT_TIMESTAMP - INTERVAL '1 month', 'WEEKLY', CURRENT_TIMESTAMP - INTERVAL '1 month', 2, 'PAID_OFF', 'Short term emergency loan - fully settled');

-- ===========================================================================
-- POPULATE LEDGER (Linking back to the Active and Paid Off loans)
-- ===========================================================================
-- Transactions for the ACTIVE Loan ('22222222-...')
INSERT INTO ledger(loan_id, sender_user_id, amount, loan_version, description)
  VALUES ('22222222-2222-2222-2222-222222222222', '00000000-0000-0000-0000-000000000000', 400.0000, 2, 'First bi-weekly principal payment'),
('22222222-2222-2222-2222-222222222222', '00000000-0000-0000-0000-000000000000', 400.0000, 3, 'Second bi-weekly principal payment');

-- Transaction for the PAID_OFF Loan ('33333333-...')
INSERT INTO ledger(loan_id, sender_user_id, amount, loan_version, description)
  VALUES ('33333333-3333-3333-3333-333333333333', '00000000-0000-0000-0000-000000000000', 1200.0000, 2, 'Final lump-sum payoff transaction');

