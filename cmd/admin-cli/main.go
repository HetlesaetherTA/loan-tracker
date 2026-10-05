package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"hetlesaether.com/loan-tracker/internal/database"

	"github.com/charmbracelet/huh"
)

var dbURL string = os.Getenv("DATABASE_URL")

func main() {
	// db and database are different...
	// Database is the sqlc instance which handles migrations and public queries.
	// db is a connection with sqlx which let the admin-cli run operations that should only be run in admin context.
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
		return
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
		return
	}

	for {
		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Loans domain control panel").
					Options(
						huh.NewOption("Accrue Interest", "accrue_interest"),
						huh.NewOption("Add loan", "add_loan"),
						huh.NewOption("Add transaction", "add_transaction"),
						huh.NewOption("Database UP", "db_up"),
						huh.NewOption("Database DOWN", "db_down"),
						huh.NewOption("EXIT CLI", "exit"),
					).
					Value(&action),
			),
		)

		if err := form.Run(); err != nil {
			log.Fatal(err)
		}

		if action == "exit" {
			break
		}

		ctx := context.Background()

		switch action {
		case "db_up":
			err := database.Up(ctx, db)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Database UP successfully!")
			}
		case "db_down":
			err := database.Down(ctx, db)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Database DOWN successfully!")
			}
		case "add_loan":
			addLoan(db)
		case "add_transaction":
			addTransaction(ctx, db)
		case "accrue_interest":
			accrueInterest(ctx, db)
		default:
			// Do nothing
		}

		fmt.Println("\nPress Enter to return to menu...")
		fmt.Scanln()
	}
}

func addLoan(db *sql.DB) {
	for {
		var userID string
		var title string
		var currency string
		var amount float64
		var interest float64
		var frequency string
		var due time.Time

		// email
		err := huh.NewInput().Title("Enter user UserID").Value(&userID).Run()
		if err != nil {
			continue
		}

		// title
		err = huh.NewInput().Title("Enter Title").Value(&title).Run()
		if err != nil {
			continue
		}

		// currency

		err = huh.NewInput().Title("Enter currency").Value(&currency).Run()
		if err != nil {
			continue
		}

		// amount
		var tmp string
		err = huh.NewInput().Title("Enter Amount").Value(&tmp).Run()
		if err != nil {
			continue
		}

		amount, err = strconv.ParseFloat(tmp, 64)
		if err != nil {
			continue
		}

		// interest
		err = huh.NewInput().Title("Enter interest (ex: 0.04 = 4%)").Value(&tmp).Run()
		tmp = strings.ToUpper(tmp)

		if err != nil {
			continue
		}

		interest, err = strconv.ParseFloat(tmp, 64)
		if err != nil {
			continue
		}

		// frequency
		var enumStr string
		var frequencyEnum []string

		err = db.QueryRow(`
					SELECT string_agg(e.enumlabel, ',' ORDER BY e.enumsortorder)
					FROM pg_enum e
					JOIN pg_type t ON e.enumtypid = t.oid
					JOIN pg_namespace n ON t.typnamespace = n.oid
					WHERE n.nspname = 'app_loan_tracker' AND t.typname = 'frequency_type_enum'
				`).Scan(&enumStr)
		if err != nil {
			continue
		}

		if enumStr != "" {
			frequencyEnum = strings.Split(enumStr, ",")
		}

		err = huh.NewInput().Title(fmt.Sprintf("Enter frequency (%s)", strings.Join(frequencyEnum, ", "))).Value(&frequency).Run()
		if err != nil {
			continue
		}

		if !slices.Contains(frequencyEnum, frequency) {
			continue
		}

		// due
		err = huh.NewInput().Title("Enter due, RFC3339 date (ex: '2026-08-13T15:00:00Z' (UTC))").Value(&tmp).Run()
		if err != nil {
			continue
		}

		due, err = time.Parse(time.RFC3339, tmp)
		if err != nil {
			continue
		}

		query := `
					INSERT INTO app_loan_tracker.loans (
						user_id, 
						currency, 
						original_principal, 
						principal, 
						yearly_interest, 
						due_at, 
						payment_frequency, 
						next_payment_at, 
						current_version, 
						status, 
						description
					) VALUES (
						$1, 
						$2, 
						$3, 
						$3,
						$4, 
						$5, 
						$6, 
						CURRENT_TIMESTAMP + (
							CASE $6::app_loan_tracker.frequency_type_enum
								WHEN 'WEEKLY' THEN INTERVAL '1 week'
								WHEN 'BI_WEEKLY' THEN INTERVAL '2 weeks'
								WHEN 'MONTHLY' THEN INTERVAL '1 month'
								WHEN 'QUARTERLY' THEN INTERVAL '3 months'
								WHEN 'YEARLY' THEN INTERVAL '1 year'
								ELSE INTERVAL '1 week'
							END
						), 
						1, 
						'ACTIVE', 
						$7
					)
				`
		_, err = db.Exec(
			query,
			userID,
			currency,
			fmt.Sprintf("%.4f", amount),
			fmt.Sprintf("%.5f", interest),
			due,
			frequency,
			title,
		)
		if err != nil {
			fmt.Println(err)
		}

		break
	}
}

// newTransaction handles the database operations for creating a transaction and updating loan principal.
func newTransaction(ctx context.Context, db *sql.DB, loanID string, senderID string, amount float64, desc string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var currentVersion int
	var currentPrincipal float64

	err = tx.QueryRowContext(ctx, `
		SELECT current_version, principal
		FROM app_loan_tracker.loans
		WHERE id = $1
	`, loanID).Scan(&currentVersion, &currentPrincipal)
	if err != nil {
		return fmt.Errorf("fetch loan state: %w", err)
	}

	var txID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO app_loan_tracker.ledger (
			loan_id,
			sender_user_id,
			amount,
			loan_version,
			description
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING transction_id
	`, loanID, senderID, amount, currentVersion+1, desc).Scan(&txID)
	if err != nil {
		return fmt.Errorf("insert ledger: %w", err)
	}

	newPrincipal := currentPrincipal + amount

	_, err = tx.ExecContext(ctx, `
		UPDATE app_loan_tracker.loans
		SET principal = $1,
			current_version = $2,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, newPrincipal, currentVersion+1, loanID)
	if err != nil {
		return fmt.Errorf("update loan: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// addTransaction collects inputs from the user and delegates creation to newTransaction.
func addTransaction(ctx context.Context, db *sql.DB) {
	var loanID string
	var senderID string
	var amount float64
	var desc string

	for {
		// 1. Collect Loan ID
		if err := huh.NewInput().Title("Enter loan ID").Value(&loanID).Run(); err != nil {
			continue
		}

		// 2. Collect Sender ID
		if err := huh.NewInput().Title("Enter sender ID (empty = default)").Value(&senderID).Run(); err != nil {
			continue
		}
		if strings.TrimSpace(senderID) == "" {
			senderID = "00000000-0000-0000-0000-000000000000"
		}

		// 3. Collect Amount
		var tmp string
		if err := huh.NewInput().Title("Enter amount").Value(&tmp).Run(); err != nil {
			continue
		}
		var err error
		amount, err = strconv.ParseFloat(tmp, 64)
		if err != nil {
			continue
		}

		// 4. Collect Description
		if err := huh.NewInput().Title("Enter description").Value(&desc).Run(); err != nil {
			continue
		}

		// 5. Execute DB Transaction
		if err := newTransaction(ctx, db, loanID, senderID, amount, desc); err != nil {
			// Failed DB operation restarts loop to re-prompt or handle error
			continue
		}

		break
	}
}

// accrueInterestForLoan calculates unpaid interest since the last calculation timestamp,
// creates an interest ledger entry, and updates the loan's principal and calculation timestamp.
func accrueInterestForLoan(ctx context.Context, db *sql.DB, loanID string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var yearlyInterestRate float64 // stored as a decimal rate (e.g. 0.05 for 5%) or percentage (e.g. 5.0)
	var interestCalculatedAt time.Time
	var currentPrincipal float64
	var currentVersion int

	// Lock the loan row for update to prevent concurrent race conditions
	err = tx.QueryRowContext(ctx, `
		SELECT yearly_interest, interest_calculated_at, principal, current_version
		FROM app_loan_tracker.loans
		WHERE id = $1
		FOR UPDATE
	`, loanID).Scan(&yearlyInterestRate, &interestCalculatedAt, &currentPrincipal, &currentVersion)
	if err != nil {
		return fmt.Errorf("fetch loan details: %w", err)
	}

	now := time.Now()
	elapsedSeconds := now.Sub(interestCalculatedAt).Seconds()

	if elapsedSeconds <= 0 {
		return nil // No time has elapsed since last accrual
	}

	// Constants for time unit conversion
	const secondsPerYear = 365.25 * 24 * 60 * 60 // 31,557,600 seconds

	// If yearlyInterestRate is stored as a percentage (e.g., 5 for 5%), convert it:
	// rate := yearlyInterestRate / 100.0
	// If it's already stored as a decimal rate (e.g., 0.05), use yearlyInterestRate directly.
	rate := yearlyInterestRate
	if rate > 1.0 {
		rate = rate / 100.0 // automatically handles integer percentages like 5 -> 0.05
	}

	// Interest = Principal * Annual Rate * (Elapsed Time / Time in Year)
	interestAmount := currentPrincipal * rate * (elapsedSeconds / secondsPerYear)

	if interestAmount <= 0 {
		return nil
	}

	systemSenderID := "00000000-0000-0000-0000-000000000000"

	var txID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO app_loan_tracker.ledger (
			loan_id,
			sender_user_id,
			amount,
			loan_version,
			description
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING transction_id
	`, loanID, systemSenderID, interestAmount, currentVersion+1, "Interest").Scan(&txID)
	if err != nil {
		return fmt.Errorf("insert interest ledger: %w", err)
	}

	newPrincipal := currentPrincipal + interestAmount

	_, err = tx.ExecContext(ctx, `
		UPDATE app_loan_tracker.loans
		SET principal = $1,
		    current_version = $2,
		    interest_calculated_at = $3,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`, newPrincipal, currentVersion+1, now, loanID)
	if err != nil {
		return fmt.Errorf("update loan state: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// accrueInterest prompts the user for input and triggers the interest accrual process.
func accrueInterest(ctx context.Context, db *sql.DB) {
	var loanID string

	for {
		err := huh.NewInput().Title("Enter Loan ID").Value(&loanID).Run()
		if err != nil {
			continue
		}

		if strings.TrimSpace(loanID) == "" {
			continue
		}

		if err := accrueInterestForLoan(ctx, db, loanID); err != nil {
			// Failed execution restarts prompt or error handling loop
			continue
		}

		break
	}
}
