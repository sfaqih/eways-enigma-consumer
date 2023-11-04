package structs

import "time"

//User for UserLogin
type User struct {
	Username string `json:"username" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	ClientID int32  `json:"client_id" validate:"required"`
	RoleID   int    `json:"role_id" validate:"required"`
	HashKey  string `json:"hash_key,omitempty"`
}

//Auth struct for responding back the json
type Auth struct {
	Message string   `json:"message"`
	Data    UserData `json:"data"`
	Client  Client   `json:"client_data"`
	Roles   []Role   `json:"roles"`
}

//Client struct for nested auth
type Client struct {
	ClientID   int32  `json:"client_id"`
	ClientCode string `json:"client_code,omitempty"`
	ClientName string `json:"client_name"`
	Status     int    `json:"status"`
}

//Role struct for nested auth
type Role struct {
	RoleID   int32  `json:"role_id" db:"role_id"`
	RoleName string `json:"role_name" db:"role_name"`
}

//UserData is parent struct for auth
type UserData struct {
	AccessToken string    `json:"access_token"`
	ExpireAt    time.Time `json:"expire_at"`
	UserID      int32     `json:"user_id" db:"id"`
	Username    string    `json:"username" db:"username"`
	HashKey     string    `json:"hash_key,omitempty"`
}

//Address is the struct for all of the entities that relates to human
type Address struct {
	ID            int    `json:"id,omitempty"`
	ProvinceID    int    `json:"province_id,omitempty"`
	ProvinceName  string `json:"province_name,omitempty"`
	CityID        int    `json:"city_id,omitempty"`
	CityName      string `json:"city_name,omitempty"`
	KecamatanID   int    `json:"kecamatan_id,omitempty"`
	KecamatanName string `json:"kecamatan_name,omitempty"`
	KelurahanID   int    `json:"kelurahan_id,omitempty"`
	KelurahanName string `json:"kelurahan_name,omitempty"`
	ZipCode       string `json:"zipcode,omitempty"`
}
