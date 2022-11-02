package channel

//type FullScoringOrder struct {
//	ScoringOrder   *ScoringOrder
//	ResponseStatus constants.Answer
//}

type ScoringOrder struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

//func NewFullScoringService() *FullScoringOrder {
//	return &FullScoringOrder{
//		ScoringOrder:   new(ScoringOrder),
//		ResponseStatus: constants.AnswerSuccessfully,
//	}
//}

func NewScoringOrder() *ScoringOrder {
	return &ScoringOrder{
		Order:   "",
		Accrual: 0.00,
		Status:  "",
	}
}
