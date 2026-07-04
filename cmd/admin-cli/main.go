package main

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"hetlesaether.com/loan-tracker/internal/database"
	"log"
	"os"

	"github.com/charmbracelet/huh"
)

var dbURL string = os.Getenv("DATABASE_URL")

func main() {
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
		}

		fmt.Println("\nPress Enter to return to menu...")
		fmt.Scanln()
	}
}
