SET search_path TO app_loan_tracker;

-- ===========================================================================
-- DOWN SCRIPT: Cleans out the test seed data
-- ===========================================================================
-- 1. Clear the ledger first due to foreign key constraints
DELETE FROM ledger
WHERE sender_user_id = '00000000-0000-0000-0000-000000000000';

-- 2. Clear the test records from the loans table
DELETE FROM loans
WHERE user_id = '00000000-0000-0000-0000-000000000000';

