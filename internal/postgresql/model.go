package postgresql

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/andynikk/gofermart/internal/constants"
	"github.com/andynikk/gofermart/internal/environment"
)

type AnswerBD struct {
	Answer constants.Answer
	JSON   []byte
}

type Order struct {
	*OrderUser
	ResponseStatus constants.Answer
}

type Account struct {
	*User
	ResponseStatus constants.Answer
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

type OrdersDB struct {
	OrderDB        []OrderDB
	ResponseStatus constants.Answer
}

type OrderDB struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC333"`
}

type Balance struct {
	BalanceDB
	ResponseStatus constants.Answer
}

type Balances struct {
	TotalBalanceDB totalBalanceDB
	ResponseStatus constants.Answer
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
type Withdraws struct {
	WithdrawDB     []withdrawDB
	ResponseStatus constants.Answer
}

type withdrawDB struct {
	Order       string    `json:"order"`
	Withdraw    float64   `json:"sum"`
	DateAccrual time.Time `json:"processed_at" format:"RFC333"`
	Current     float64   `json:"current,omitempty"`
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

type MapResult = map[string]Result

type Result interface {
	InJSON() ([]byte, error)
	FromJSON([]byte) error
}

func (o *OrdersDB) InJSON() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.OrderDB, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *OrdersDB) FromJSON(byte []byte) error {
	if err := json.Unmarshal(byte, &o.OrderDB); err != nil {
		return err
	}
	return nil
}

func (o *OrderWithdraw) InJSON() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *OrderWithdraw) FromJSON(byte []byte) error {
	if err := json.Unmarshal(byte, &o); err != nil {
		return err
	}
	return nil
}

func (o *User) InJSON() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *User) FromJSON(byte []byte) error {
	if err := json.Unmarshal(byte, &o); err != nil {
		return err
	}
	return nil
}

func (o *Balances) InJSON() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.TotalBalanceDB, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Balances) FromJSON(byte []byte) error {
	if err := json.Unmarshal(byte, &o.TotalBalanceDB); err != nil {
		return err
	}
	return nil
}

func (o *Withdraws) InJSON() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.WithdrawDB, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Withdraws) FromJSON(byte []byte) error {
	if err := json.Unmarshal(byte, &o.WithdrawDB); err != nil {
		return err
	}
	return nil
}
