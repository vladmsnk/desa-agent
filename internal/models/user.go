package models

type User struct {
	UserHash string               `json:"user_hash"`
	Status   UserStatus           `json:"status"`
	IdpType  IdentityProviderType `json:"idp_type"`
	PII      *UserPII             `json:"pii,omitempty"`
}

type UserStatus int

const (
	UserStatusUnspecified UserStatus = iota
	UserStatusActive
	UserStatusDisabled
)

type IdentityProviderType int

const (
	IdentityProviderTypeUnspecified IdentityProviderType = iota
	IdentityProviderTypeActiveDirectory
	IdentityProviderTypeLDAP
)

type UserPII struct {
	SourceID    string      `json:"source_id,omitempty"`
	Username    string      `json:"username,omitempty"`
	Email       string      `json:"email,omitempty"`
	DisplayName string      `json:"display_name,omitempty"`
	FirstName   string      `json:"first_name,omitempty"`
	LastName    string      `json:"last_name,omitempty"`
	Phone       string      `json:"phone,omitempty"`
	Department  string      `json:"department,omitempty"`
	Title       string      `json:"title,omitempty"`
	ManagerID   string      `json:"manager_id,omitempty"`
	EmployeeID  string      `json:"employee_id,omitempty"`
	Location    string      `json:"location,omitempty"`
	Attributes  []Attribute `json:"attributes,omitempty"`
}

type Attribute struct {
	Key   AttributeKey `json:"key"`
	Value string       `json:"value"`
}

type AttributeKey int

const (
	AttributeKeyUnspecified AttributeKey = iota
)

type Filter struct {
	Usernames []string `json:"usernames,omitempty"`
}
