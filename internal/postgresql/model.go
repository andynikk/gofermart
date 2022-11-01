package postgresql

import (
	"encoding/json"
	"time"

	"github.com/andynikk/gofermart/internal/constants"
)

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
	*BalanceDB
	ResponseStatus constants.Answer
}

type Balances struct {
	TotalBalanceDB *totalBalanceDB
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

type MapHandlerJSON = map[string]HandlerJSON

type HandlerJSON interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func (o *OrdersDB) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.OrderDB, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *OrdersDB) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o.OrderDB); err != nil {
		return err
	}
	return nil
}

func (o *OrderWithdraw) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *OrderWithdraw) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o); err != nil {
		return err
	}
	return nil
}

func (o *Account) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.User, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Account) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o.User); err != nil {
		return err
	}
	return nil
}

func (o *Balances) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.TotalBalanceDB, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Balances) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o.TotalBalanceDB); err != nil {
		return err
	}
	return nil
}

func (o *Withdraws) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.WithdrawDB, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Withdraws) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o.WithdrawDB); err != nil {
		return err
	}
	return nil
}

// create object
func NewOrder() *Order {
	return &Order{
		OrderUser:      new(OrderUser),
		ResponseStatus: constants.AnswerSuccessfully,
	}
}

func NewAccount() *Account {
	return &Account{
		User:           new(User),
		ResponseStatus: constants.AnswerSuccessfully,
	}
}

func NewOrdersDB() *OrdersDB {
	return &OrdersDB{
		OrderDB:        []OrderDB{},
		ResponseStatus: constants.AnswerSuccessfully,
	}
}

func NewBalance() *Balance {
	return &Balance{
		BalanceDB:      new(BalanceDB),
		ResponseStatus: constants.AnswerSuccessfully,
	}
}

func NewBalances() *Balances {
	return &Balances{
		TotalBalanceDB: new(totalBalanceDB),
		ResponseStatus: constants.AnswerSuccessfully,
	}
}

func NewWithdraws() *Withdraws {
	return &Withdraws{
		WithdrawDB:     []withdrawDB{},
		ResponseStatus: constants.AnswerSuccessfully,
	}
}

func NewOrderWithdraw() *OrderWithdraw {
	return &OrderWithdraw{
		Order:    "",
		Withdraw: 0.00,
	}
}
