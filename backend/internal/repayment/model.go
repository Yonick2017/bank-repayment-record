package repayment

import "time"

type Record struct {
	ID          int64
	CardName    string
	Currency    string
	AmountCents int64
	RepaymentAt time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Filters struct {
	CardName string
	Currency string
}
