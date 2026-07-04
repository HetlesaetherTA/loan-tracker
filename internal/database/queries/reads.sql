-- name: GetUserLoans :many
SELECT *
FROM app_loan_tracker.user_loan_view
WHERE user_id = $1;

-- name: GetUserLoanLedger :many
SELECT ledger.*
FROM app_loan_tracker.user_ledger_view AS ledger
  JOIN app_loan_tracker.user_loan_view ON id = ledger.id
WHERE loan_id = $1
  AND user_id = $2;

-- name: GetUserLoanByID :one
SELECT *
FROM app_loan_tracker.user_loan_view
WHERE id = $1
  AND user_id = $2;

