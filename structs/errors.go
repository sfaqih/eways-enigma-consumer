package structs

//ErrorMessage struct for general error
type ErrorMessage struct {
	TrxID       int    `json:"trx_id,omitempty"`
	Code        int    `json:"code,omitempty"`
	Message     string `json:"message,omitempty"`
	Data        string `json:"data,omitempty"`
	SysMessage  string `json:"system_message,omitempty"`
	RespMessage string `json:"response_from_vendor,omitempty"`
	ExtRefID    string `json:"external_id_from_vendor,omitempty"`
	ReqMessage  string `json:"request_to_vendor,omitempty"`
	ReqSesame   string `json:"request_from_sesame,omitempty"`
}

//var Message and system message
var (
	ErrNotFound      = "The data can't be found"
	Success          = "success"
	Unauthorized     = "Unauthorized, please login"
	AuthNotFound     = "Auth header is not found"
	TokenInv         = "token is invalid"
	GenTokenErr      = "generate token is error"
	UserNotFound     = "user is not found"
	IncorrectPass    = "password is not match"
	QueryErr         = "query error"
	DBErr            = "DB Connection Error"
	PrepStmtErr      = "Prepared statement error"
	RowsAffErr       = "error while getting rows affected"
	LastIDErr        = "error while getting last inserted id"
	Validate         = "error when validating the data"
	Email            = "email is already exists"
	EmptyID          = "ID can't be empty"
	CustErr          = "Customer doesn't exist or inactive"
	DupCust          = "The existing customer is already exist, please check your data again! Duplicate with CustID:"
	CustFieldErr     = "One of these fields (phone_1, phone_2, email or member_id) must not be empty!"
	UserBalanceEmpty = "The customer doesn't have any balance or 0"
	OutofStock       = "the reward code requested is out of stock"
	RewardErr        = "the reward requested is not exist"
	RewardTypeErr    = "the reward type requested is not exist"
	MasterRewardErr  = "the master reward is nil"
	MasterProdErr    = "the master product is nil"
	TypeErr          = "reward type is not recognized"
	ProdErr          = "Product code is not recognized"
	BalanceIDErr     = "No balance ID matched with the configuration"
	ClientErr        = "Client code doesn't exist"
	MaxQty           = "Maximum qty of redemption is reached"
	QtyErr           = "Qty doesn't match with the configuration"
	ConfErr          = "Configuration doesn't exist"
	BalanceErr       = "Balance configuration doesn't exist"
	BankingConf      = "Banking failed, please check your configuration"
	RedeemConf       = "Redemption failed, please check your configuration"
	InsuffBalance    = "insufficient balance"
	MinBalance       = "the minimum balance doesn't meet the configuration"
	SuccessUpdate    = "success updated data"
	SuccessInsert    = "success insert data"
	SuccessBanking   = "banking is succeeded"
	SuccessRedeem    = "redeem is succeeded"
	InProgress       = "In Progress"
	BlockErr         = "Auto Block configuration doesn't exist"
	InvalidFormat    = "Invalid Format"
	PLNNumberErr     = "PLN Number is nil"
	PhoneNumberErr   = "Phone Number is nil"
	ETollNumberErr   = "EToll Number is nil"
	RedSuccessErr    = "errow while getting redemption_success_logs ID"
	VendorErr        = "Something error with 3rd Party, please try again later"
	SuccessRedirect  = "Success redirect to vendor"

	//Custom Message
	ErrNominal             = "Invalid Nominal"
	ErrQuantity            = "Invalid Quantity"
	ErrDate                = "Invalid Date"
	ErrNoRecordFound       = "No Record Found"
	ErrMaxQuantity         = "Invalid Max Quantity"
	ErrMinQuantity         = "Invalid Min Quantity"
	ErrVendorAliasNotFound = "vendor alias is not found"

	//IAK Message
	IAK_Success      = "SUCCESS"
	IAK_Process      = "PROCESS"
	IAK_Undef        = "UNDEFINED RESPONSE CODE"
	ErrTrxNotFound   = "TRANSACTION NOT FOUND"
	ErrFailed        = "Failed! Please try again"
	ErrCustBlocked   = "CUSTOMER NUMBER BLOCKED"
	ErrInvalidNumber = "INCORRECT DESTINATION NUMBER"
	ErrNumNotMatch   = "NUMBER NOT MATCH WITH OPERATOR"
	ErrInsuffDeposit = "INSUFFICIENT DEPOSIT"
	ErrCodeNotFound  = "CODE NOT FOUND"
	ErrInvalidIP     = "INVALID IP ADDRESS"
	ErrOutOfService  = "PRODUCT IS TEMPORARILY OUT OF SERVICE"
	ErrXML           = "ERROR IN XML FORMAT"
	ErrPageNotFound  = "PAGE NOT FOUND"
	ErrMaxNum1Day    = "MAXIMUM 1 NUMBER 1 TIME IN 1 DAY"
	ErrTooLong       = "NUMBER IS TOO LONG"
	ErrAuth          = "WRONG AUTHENTICATION"
	ErrComm          = "WRONG COMMAND"
	ErrNumBlocked    = "THIS DESTINATION NUMBER HAS BEEN BLOCKED"
	ErrMaxNum1DayAny = "MAXIMUM 1 NUMBER WITH ANY CODE 1 TIME IN 1 DAY"
)

var (
	CodeErrNotFound          = 600
	CodeSuccessRedirect      = 200
	CodeSuccess              = 200
	CodeUnauthorized         = 401
	CodeAuthNotFound         = 400
	CodeTokenInv             = 401
	CodeGenTokenErr          = 401
	CodeInvalidFormat        = 422
	CodeUserNotFound         = 601
	CodeIncorrectPass        = 602
	CodeQueryErr             = 603
	CodeDBErr                = 604
	CodePrepStmtErr          = 605
	CodeRowsAffErr           = 607
	CodeLastIDErr            = 608
	CodeValidate             = 609
	CodeEmail                = 610
	CodeEmptyID              = 611
	CodeCustErr              = 612
	CodeDupCust              = 613
	CodeCustFieldErr         = 614
	CodeUserBalanceEmpty     = 615
	CodeOutofStock           = 616
	CodeRewardErr            = 617
	CodeTypeErr              = 618
	CodeProdErr              = 619
	CodeBalanceIDErr         = 620
	CodeClientErr            = 621
	CodeMaxQty               = 622
	CodeQtyErr               = 623
	CodeConfErr              = 624
	CodeBalanceErr           = 625
	CodeBankingConf          = 626
	CodeRedeemConf           = 627
	CodeInsuffBalance        = 628
	CodeMinBalance           = 629
	CodeErrNominal           = 630
	CodeErrQuantity          = 631
	CodeErrDate              = 632
	CodeBlockBanking         = 633
	CodeBlockRedemption      = 634
	CodeBlockRegistration    = 635
	CodeBlockErr             = 636
	CodeMasterRewardErr      = 637
	CodeMasterProdErr        = 638
	CodePLNNumberErr         = 639
	CodePhoneNumberErr       = 640
	CodeRedSuccessErr        = 641
	CodeVendorErr            = 642
	CodeRewardTypeErr        = 643
	CodeRewardVendorAliasErr = 644
	CodeETollNumberErr       = 645
	CodeSuccessUpdate        = 200
	CodeSuccessInsert        = 200
	CodeSuccessBanking       = 200
	CodeSuccessRedeem        = 200
	CodeInProgress           = 201

	//IAK Code
	CodeIAK_Success      = 0
	CodeIAK_Process      = 39
	CodeIAK_Undef        = 201
	CodeErrTrxNotFound   = 6
	CodeErrFailed        = 7
	CodeErrCustBlocked   = 13
	CodeErrInvalidNumber = 14
	CodeErrNumNotMatch   = 16
	CodeErrInsuffDeposit = 17
	CodeErrCodeNotFound  = 20
	CodeErrInvalidIP     = 102
	CodeErrOutOfService  = 106
	CodeErrXML           = 107
	CodeErrPageNotFound  = 117
	CodeErrMaxNum1Day    = 202
	CodeErrTooLong       = 203
	CodeErrAuth          = 204
	CodeErrComm          = 205
	CodeErrNumBlocked    = 206
	CodeErrMaxNum1DayAny = 207

	//IAK Code
)
