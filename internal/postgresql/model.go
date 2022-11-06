package postgresql

import (
	"encoding/json"
	"time"
)

type MapHandlerJSON = map[string]HandlerJSON

type Balances struct {
	Number    string  `json:"number"`
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
	Total     float64 `json:"total"`
}

type TotalBalance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Order struct {
	Number     string    `json:"orderID"`
	User       string    `json:"userID"`
	CreatedAt  time.Time `json:"createdAt"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	FailedAt   time.Time `json:"failedAt"`
	Status     string    `json:"status"`
}

type User struct {
	Name     string `json:"login"`
	Password string `json:"password"`
}

type OrderWithdraw struct {
	Order    string  `json:"order"`
	Withdraw float64 `json:"sum"`
}

type OrdersAccrual struct {
	OrderAccrual []OrderAccrual
}

type OrderAccrual struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" format:"RFC333"`
}

type Withdraws struct {
	Withdraw []Withdraw
}

type Withdraw struct {
	Order       string    `json:"order"`
	Withdraw    float64   `json:"sum"`
	DateAccrual time.Time `json:"processed_at" format:"RFC333"`
	Current     float64   `json:"current,omitempty"`
}

type HandlerJSON interface {
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

func (o *OrdersAccrual) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(&o.OrderAccrual, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *OrdersAccrual) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o.OrderAccrual); err != nil {
		return err
	}
	return nil
}

func (o *Order) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(&o, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Order) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o); err != nil {
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

func (o *User) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *User) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, o); err != nil {
		return err
	}
	return nil
}

func (o *Balances) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Balances) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o); err != nil {
		return err
	}
	return nil
}

func (o *Withdraws) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o.Withdraw, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Withdraws) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o.Withdraw); err != nil {
		return err
	}
	return nil
}

func (o *Withdraw) Marshal() ([]byte, error) {
	strJSON, err := json.MarshalIndent(o, "", " ")
	if err != nil {
		return nil, err
	}
	return strJSON, nil
}

func (o *Withdraw) Unmarshal(byte []byte) error {
	if err := json.Unmarshal(byte, &o); err != nil {
		return err
	}
	return nil
}
