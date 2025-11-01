package config

type Config struct {
	GRPCServerPort         int
	GRPCWalletClientPort   string
	HTTPProductsClientPort string
	KafkaHost              string
	PGUser                 string
	PGPassword             string
	PGName                 string
	PGHost                 string
	PGPort                 string
}
