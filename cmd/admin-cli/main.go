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
	// db is a connection with sqlx which let the admin-cli run operations that should only be ran in admin context.
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
		// very crude, quick solution. Probably TODO: clean up
		case "add_loan":
			for {
				var userID string
				var title string
				var currency string
				var amount float64
				var interest float64
				var frequency string
				var due time.Time

				// email
				err = huh.NewInput().Title("Enter user UserID").Value(&userID).Run()

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
				_, err := db.Exec(
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
		case "add_transaction":

		default:
			// Dont do anything
		}

		fmt.Println("\nPress Enter to return to menu...")
		fmt.Scanln()
	}
}
