package channel

type ScoringOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

func NewScoringOrder() *ScoringOrder {
	return &ScoringOrder{
		Order:   "",
		Accrual: 0.00,
		Status:  "",
	}
}
