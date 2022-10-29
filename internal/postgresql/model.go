package postgresql

import (
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/environment"
)

type AnswerBD struct {
	Answer constants.Answer
	JSON   []byte
}

type User struct {
	Name     string `json:"login"`
	Password string `json:"password"`
}

type OrderUser struct {
	Number     string    `json:"orderID"`
	User       string    `json:"userID"`
	CreatedAt  time.Time `json:"createdAt"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	FailedAt   time.Time `json:"failedAt"`
	Status     string    `json:"status"`
}

type OrderWithdraw struct {
	Order    string  `json:"order"`
	Withdraw float64 `json:"sum"`
}

type orderDB struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC333"`
}

type BalanceDB struct {
	Number    string  `json:"number"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
	Total     float64 `json:"total"`
}

type totalBalanceDB struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type withdrawDB struct {
	Order       int       `json:"order"`
	Withdraw    float64   `json:"sum"`
	DateAccrual time.Time `json:"processed_at" format:"RFC333"`
}

type FullScoringSystem struct {
	ScoringSystem *ScoringSystem
	Answer        constants.Answer
}

type ScoringSystem struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}

type GoodOrderSS struct {
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

type OrderSS struct {
	Order       string        `json:"order"`
	GoodOrderSS []GoodOrderSS `json:"goods"`
}

type DBConnector struct {
	Pool *pgxpool.Pool
	Cfg  *environment.DBConfig
}
