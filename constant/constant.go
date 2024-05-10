package constant

const (
	URL                   = "http://10.0.200.203:8863/api/v1/pbz/statement"
	ACCOUNT_STATEMENT_URL = "http://172.20.1.113:7001/BanksESB_PE07EDC00/BanksMiddleware_PE07EDC00Port"
	ACCOUNT_PAYMENT_URL   = "http://10.0.200.203:8863/api/v1/pbz/payment"
	//ACCOUNT_STATEMENT_URL = "http://192.168.101.113:7001/BanksESB_PE07EDC00/BanksMiddleware_PE07EDC00Port"

	// maps of list of banks

)

var (
	BANKS_ACCOUNT = map[string]string{
		"0404003001": "0404003001",
		"0400714000": "0400714000",
	}

	DATASENT = make(map[string]string)
)
