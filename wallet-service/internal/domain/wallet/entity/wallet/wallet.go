package wallet

// Wallet represents a user's wallet entity
type Wallet struct {
	UserID   int64 // User identifier
	Balance  int64 // Available balance in cents/kopecks
	Reserved int64 // Reserved funds in cents/kopecks
}
