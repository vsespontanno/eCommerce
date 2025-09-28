package entity

type OrderStatus string

const (
	StatusPending           OrderStatus = "pending"
	StatusProcessingPayment OrderStatus = "processing_payment"
	StatusPaid              OrderStatus = "paid"
	StatusCompleted         OrderStatus = "completed"
	StatusFailed            OrderStatus = "failed"
	StatusCancelled         OrderStatus = "cancelled" // After successful compensation
)
