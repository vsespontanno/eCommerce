package wallet

import (
	proto "github.com/vsespontanno/eCommerce/proto/wallet"
)

type WalletServer struct {
	proto.UnimplementedWalletServer
}

type Wallet interface {
	
}