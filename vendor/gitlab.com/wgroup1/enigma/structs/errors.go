package structs

//ErrorMessage struct for general error
type ErrorMessage struct {
	Message        string `json:"message,omitempty"`
	Data           string `json:"data,omitempty"`
	SysMessage     string `json:"system_message,omitempty"`
	Code           int    `json:"code,omitempty"`
	RespMessage    string `json:"response_from_vendor,omitempty"`
	ReqMessage     string `json:"request_to_vendor,omitempty"`
	ReqSesame      string `json:"request_from_sesame,omitempty"`
	MessageID      string `json:"id,omitempty"`
	ConversationID string `json:"conversationId,omitempty"`
	ChannelID      string `json:"channelId,omitempty"`
	Source         Source `json:"source,omitempty"`
}

type ErrorMessageSparkPost struct {
	Status  int    `json:"status,omitempty"`
	Data    string `json:"data,omitempty"`
	Message string `json:"message,omitempty"`
}

type RequestEncrypt struct {
	Payload string `json:"payload,omitempty"`
}

type ResponseEncrypt struct {
	Data   string `json:"data,omitempty"`
	Status string `json:"status,omitempty"`
	Code   int    `json:"code,omitempty"`
}

type APIResponse struct {
	Message string      `json:"message,omitempty"`
	Code    int         `json:"code,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type HTTPRequest struct {
	Url            string `json:"url,omitempty"`
	RequestBody    string `json:"request_body,omitempty"`
	Method         string `json:"method,omitempty"`
	RequestHeader  string `json:"request_header,omitempty"`
	ResponseStatus int    `json:"response_status,omitempty"`
	ResponseHeader string `json:"response_header,omitempty"`
	ResponseBody   []byte `json:"response_body,omitempty"`
	Outbound       []byte `json:"outbound,omitempty"`
}

//var Message and system message
var (
	ErrNotFound   = "The data can't be found"
	Success       = "success"
	Unauthorized  = "Unauthorized, please login"
	AuthNotFound  = "Auth header is not found"
	TokenInv      = "token is invalid"
	GenTokenErr   = "generate token is error"
	UserNotFound  = "user is not found"
	IncorrectPass = "password is not match"
	QueryErr      = "query error"
	DBErr         = "DB Connection Error"
	PrepStmtErr   = "Prepared statement error"
	RowsAffErr    = "error while getting rows affected"
	LastIDErr     = "error whilte getting last inserted id"
	Validate      = "error when validating the data"
	NomorInd      = "nomor induk is exist for that institutions"
	Email         = "email is already exists"
	Family        = "Family ID is already exists"
	EmptyID       = "ID can't be empty"
	ClientErr     = "Client code doesn't exist"
	SuccessUpdate = "success updated data"
	SuccessInsert = "success insert data"
)
