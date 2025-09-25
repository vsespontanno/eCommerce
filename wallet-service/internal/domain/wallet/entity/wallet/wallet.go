package wallet

type Wallet struct {
	userID   int64
	balance  float64
	reserved float64
}

func (w Wallet) TopUp(amount float64) {
	w.balance += amount
}

func (w Wallet) Balance() float64 {
	return w.balance
}
