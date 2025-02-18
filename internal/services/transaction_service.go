package services

type TransactionService interface {
	Create()
}

func NewTransactionService(ws WalletService) TransactionService {
	return &transactionService{
		ws: ws,
	}
}

type transactionService struct {
	ws WalletService
}

func (s *transactionService) Create() {
	// create a transaction from here.
	// tx := btcutil.NewTx(wire.NewMsgTx(wire.TxVersion))

}
