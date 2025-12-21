package models

type User struct {
	ID        int64
	FirstName string
	LastName  string
	Email     string
	Balance   float64
	Role      string // TODO: implement role-based access control
	PassHash  []byte
}
