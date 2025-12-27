package models

type User struct {
	UserHash string
	Status   UserStatus
	IdpType  IdentityProviderType
	PII      *UserPII
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
	SourceID    string
	Username    string
	Email       string
	DisplayName string
	FirstName   string
	LastName    string
	Phone       string
	Department  string
	Title       string
	ManagerID   string
	EmployeeID  string
	Location    string
	Attributes  []Attribute
}

type Attribute struct {
	Key   AttributeKey
	Value string
}

type AttributeKey int

const (
	AttributeKeyUnspecified AttributeKey = iota
)
