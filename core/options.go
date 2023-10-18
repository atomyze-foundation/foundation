package core

// ContractOptions
// TxTTL - Время жизни транзакции в секундах. По умолчанию 0 - вечная жизнь.
// Проверяется при исполнении батча. В US равно 30 секунд.
// BatchPrefix - префик с которым в hlf сохраняются преимаджи. По умолчанию "batchTransactions"
// US задает свой более короткий префикс из одного или двух символов
// NonceTTL - время в секундах для nonce. Если пытаемся выполнить в батче транзакцию,
// которая старее максимального нонса (на данный моменд времени) более чем на NonceTTL,
// то мы ее не исполним с ошибкой. В US равно 50 секунд.
// Если NonceTTL = 0, то проверка происходит "по старому" при добавлении преимаджа.
// IsOtherNoncePrefix - исторически сложилось, что для нонсов в atomyze-us используется другой префикс.
// Поддержать разные префиксы мы обязаны, но плодить их не стоит. Поэтому только флаг.
type ContractOptions struct {
	DisabledFunctions  []string
	DisableSwaps       bool
	DisableMultiSwaps  bool
	TxTTL              uint
	BatchPrefix        string
	NonceTTL           uint
	IsOtherNoncePrefix bool
}
