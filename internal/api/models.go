package api

import (
	"time"

	"github.com/shopspring/decimal"
	"hetlesaether.com/loan-tracker/internal/database"
)

type Loan struct {
	ID                   string `json:"id,omitempty"`
	UserID               string `json:"user_id,omitempty"`
	OriginalPrincipal    string `json:"original_principal,omitempty"`
	Principal            string `json:"principal,omitempty"`
	YearlyInterest       string `json:"yearly_interest,omitempty"`
	InterestCalculatedAt string `json:"interest_calculated_at,omitempty"`
	DueAt                string `json:"due_at,omitempty"`
	PaymentFrequency     string `json:"payment_frequency,omitempty"`
	NextPaymentAt        string `json:"next_payment_at,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
	CurrentVersion       int32  `json:"current_version,omitempty"`
	Status               string `json:"status,omitempty"`
	Description          string `json:"description,omitempty"`
}

type Transaction struct {
	TransctionID string `json:"transction_id,omitempty"`
	LoanID       string `json:"loan_id,omitempty"`
	SenderUserID string `json:"sender_user_id,omitempty"`
	Amount       string `json:"amount,omitempty"`
	LoanVersion  int32  `json:"loan_version,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	Description  string `json:"description,omitempty"`
}

func parseLoan(loan database.AppLoanTrackerUserLoanView) Loan {
	return Loan{
		ID:                   loan.ID.String(),
		UserID:               loan.UserID.String(),
		OriginalPrincipal:    pgNumericToDecimal(loan.OriginalPrincipal).String(),
		Principal:            pgNumericToDecimal(loan.Principal).String(),
		YearlyInterest:       pgNumericToDecimal(loan.YearlyInterest).Mul(decimal.NewFromInt(100)).String(),
		InterestCalculatedAt: loan.InterestCalculatedAt.Time.Format(time.RFC3339),
		DueAt:                loan.DueAt.Time.Format(time.RFC3339),
		PaymentFrequency:     string(loan.PaymentFrequency),
		NextPaymentAt:        loan.NextPaymentAt.Time.Format(time.RFC3339),
		UpdatedAt:            loan.UpdatedAt.Time.Format(time.RFC3339),
		CurrentVersion:       loan.CurrentVersion,
		Status:               string(loan.Status),
		Description:          *loan.Description,
	}
}

func parseTransaction(transaction database.AppLoanTrackerUserLedgerView) Transaction {
	return Transaction{
		TransctionID: transaction.TransctionID.String(),
		LoanID:       transaction.LoanID.String(),
		SenderUserID: transaction.SenderUserID.String(),
		Amount:       pgNumericToDecimal(transaction.Amount).String(),
		LoanVersion:  transaction.LoanVersion,
		CreatedAt:    transaction.CreatedAt.Time.Format(time.RFC3339),
		Description:  *transaction.Description,
	}
}
