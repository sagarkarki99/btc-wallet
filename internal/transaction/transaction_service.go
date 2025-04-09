package transaction

type TransactionService interface {
	Create()
	GetTransaction()
}

func NewTransactionService() TransactionService {

	return &transactionService{}
}

type transactionService struct {
}

func (s *transactionService) Create() {
	// create a transaction from here.
	// fetch utxos for the wallet address.
	// tx := btcutil.NewTx(wire.NewMsgTx(wire.TxVersion))

}

func (s *transactionService) GetTransaction() {

}
