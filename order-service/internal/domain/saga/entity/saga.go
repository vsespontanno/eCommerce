package entity

type Saga struct {
	ID          string
	Type        string
	Status      string
	CurrentStep int
	Payload     map[string]interface{}
}
