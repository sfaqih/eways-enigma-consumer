package structs

type AuthVendor struct {
	Token     string  `json:"token,omitempty"`
	ExpiredAt string  `json:"expires_after,omitempty"`
	Users     []Users `json:"users,omitempty"`
	Alias     string  `json:"alias,omitempty"`
	ID        string  `json:"id,omitempty"`
	URL       string  `json:"url,omitempty"`
	Status    int     `json:"status,omitempty"`
}

type Users struct {
	Token     string `json:"token,omitempty"`
	ExpiredAt string `json:"expires_after,omitempty"`
}

// v.id, v.vendor_alias, v.url, v.uri_login, v.username, v.password , v.auth_type , v.header_prefix
type VendorLogin struct {
	ID           int    `json:"id"`
	Alias        string `json:"alias"`
	URL          string `json:"url"`
	URI          string `json:"uri_login"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Method       string `json:"method"`
	AuthType     string `json:"auth_type"`
	HeaderPrefix string `json:"header_prefix"`
}

type VendorService struct {
	ID                int    `json:"vendor_id,omitempty"`
	Alias             string `json:"alias,omitempty"`
	Channel           string `json:"channel,omitempty"`
	URL               string `json:"url,omitempty"`
	URILogin          string `json:"uri_login,omitempty"`
	Username          string `json:"username,omitempty"`
	Password          string `json:"password,omitempty"`
	LoginMethod       string `json:"login_method,omitempty"`
	LoginAuthType     string `json:"login_auth_type,omitempty"`
	LoginHeaderPrefix string `json:"login_header_prefix,omitempty"`
	URI               string `json:"uri,omitempty"`
	Method            string `json:"method,omitempty"`
	AuthType          string `json:"auth_type,omitempty"`
	HeaderPrefix      string `json:"header_prefix,omitempty"`
	Token             string `json:"token,omitempty"`
	ExpiredAt         string `json:"expired_at,omitempty"`
	Status            int    `json:"status,omitempty"`
	ClientID          int    `json:"client_id,omitempty"`
	FromSender        string `json:"from_sender,omitempty"`
	ReplyTo           string `json:"reply_to,omitempty"`
}

type CallBackAuth struct {

	//AuthType        string `json:"auth_type,omitempty" validate:"oneof=noauth apikey bearertoken basicauth"`
	AuthType        string `json:"auth_type,omitempty"`
	AuthHeader      string `json:"auth_header,omitempty"`
	AuthHeaderValue string `json:"auth_header_value,omitempty"`
	CallBackUrl     string `json:"call_back_url,omitempty"`
	//CallBackMethod  string `json:"method,omitempty" validate:"oneof=post get"`
	CallBackMethod string `json:"method,omitempty"`

	ApiKey      string `json:"key,omitempty"`
	ApiKeyValue string `json:"value,omitempty"`
	//ApiKeyAddTo string `json:"add_to,omitempty"`

	AuthToken    string `json:"token,omitempty"`
	AuthUserName string `json:"user_name,omitempty"`
	AuthPassword string `json:"password,omitempty"`
}

type Damcorp_Auth_Response struct {
	Success bool                       `json:"success,omitempty"`
	Code    int                        `json:"code,omitempty"`
	Message string                     `json:"message,omitempty"`
	Data    Damcorp_Auth_Response_Data `json:"data,omitempty"`
}

type Damcorp_Auth_Response_Data struct {
	Token string `json:"token,omitempty"`
}
