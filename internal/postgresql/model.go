package postgresql

import (
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Cfg struct {
	Token string
	*pgxpool.Pool
	Key string `json:"key"`
}

type User struct {
	Name     string `json:"login"`
	Password string `json:"password"`
}

type Account struct {
	User
	*Cfg
}

type Order struct {
	Number string
	*Cfg
}

type OrderWithdraw struct {
	Order    string  `json:"order"`
	Withdraw float64 `json:"sum"`
	*Cfg
}

type orderDB struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC333"`
}

type BalansDB struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
	Total     float64 `json:"total"`
}

type withdrawDB struct {
	Order       int       `json:"order"`
	Withdraw    float64   `json:"sum"`
	DateAccrual time.Time `json:"processed_at" format:"RFC333"`
	Current     float64   `json:"current,omitempty"`
}

type FullScoringSystem struct {
	ScoringSystem *ScoringSystem
	HTTPStatus    int
}

type ScoringSystem struct {
	Order   string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
