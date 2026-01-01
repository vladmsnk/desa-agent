package ldap

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"

	"desa-agent/internal/config"
	"desa-agent/internal/models"

	"github.com/go-ldap/ldap/v3"
)

type Adapter struct {
	cfg  config.IDPConfig
	conn *ldap.Conn
}

func New(cfg config.IDPConfig) (*Adapter, error) {
	a := &Adapter{cfg: cfg}

	if err := a.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}

	return a, nil
}

func (a *Adapter) connect() error {
	address := fmt.Sprintf("%s:%d", a.cfg.Host, a.cfg.Port)

	var conn *ldap.Conn
	var err error

	if a.cfg.UseTLS {
		conn, err = ldap.DialTLS("tcp", address, &tls.Config{
			InsecureSkipVerify: true, // For testing; in production use proper certs
		})
	} else {
		conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		return fmt.Errorf("failed to dial LDAP server: %w", err)
	}

	if err := conn.Bind(a.cfg.BindDN, a.cfg.BindPass); err != nil {
		conn.Close()
		return fmt.Errorf("failed to bind to LDAP: %w", err)
	}

	a.conn = conn
	return nil
}

func (a *Adapter) ensureConnected() error {
	if a.conn == nil || a.conn.IsClosing() {
		return a.connect()
	}
	return nil
}

func (a *Adapter) getUsersDN() string {
	if a.cfg.UsersDN != "" {
		return a.cfg.UsersDN
	}
	return a.cfg.BaseDN
}

func (a *Adapter) GetUser(ctx context.Context, userID string) (*models.User, error) {
	if err := a.ensureConnected(); err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		a.cfg.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,
		0,
		false,
		fmt.Sprintf("(uid=%s)", ldap.EscapeFilter(userID)),
		getUserAttributes(),
		nil,
	)

	result, err := a.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("user not found: %s", userID)
	}

	return a.entryToUser(result.Entries[0]), nil
}

func (a *Adapter) ListUsers(ctx context.Context) ([]models.User, error) {
	if err := a.ensureConnected(); err != nil {
		return nil, err
	}

	searchRequest := ldap.NewSearchRequest(
		a.getUsersDN(),
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		"(objectClass=inetOrgPerson)",
		getUserAttributes(),
		nil,
	)

	result, err := a.conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("LDAP search failed: %w", err)
	}

	users := make([]models.User, 0, len(result.Entries))
	for _, entry := range result.Entries {
		users = append(users, *a.entryToUser(entry))
	}

	return users, nil
}

func (a *Adapter) Close() error {
	if a.conn != nil {
		a.conn.Close()
		a.conn = nil
	}
	return nil
}

func (a *Adapter) entryToUser(entry *ldap.Entry) *models.User {
	uid := entry.GetAttributeValue("uid")
	managerDN := entry.GetAttributeValue("manager")

	// Extract manager UID from DN (e.g., "uid=manager,ou=People,dc=desaidentity,dc=com" -> "manager")
	managerID := extractUIDFromDN(managerDN)

	return &models.User{
		UserHash: uid, // Using uid as hash for now
		Status:   models.UserStatusActive,
		IdpType:  models.IdentityProviderTypeLDAP,
		PII: &models.UserPII{
			SourceID:    entry.DN,
			Username:    uid,
			Email:       entry.GetAttributeValue("mail"),
			DisplayName: entry.GetAttributeValue("cn"),
			FirstName:   entry.GetAttributeValue("givenName"),
			LastName:    entry.GetAttributeValue("sn"),
			Phone:       entry.GetAttributeValue("telephoneNumber"),
			Title:       entry.GetAttributeValue("title"),
			Department:  entry.GetAttributeValue("departmentNumber"),
			ManagerID:   managerID,
			EmployeeID:  entry.GetAttributeValue("employeeNumber"),
			Location:    entry.GetAttributeValue("l"),
		},
	}
}

func getUserAttributes() []string {
	return []string{
		"uid",
		"cn",
		"sn",
		"givenName",
		"mail",
		"telephoneNumber",
		"title",
		"departmentNumber",
		"manager",
		"employeeNumber",
		"l",
	}
}

func extractUIDFromDN(dn string) string {
	if dn == "" {
		return ""
	}

	// Parse DN like "uid=manager,ou=People,dc=desaidentity,dc=com"
	parts := strings.Split(dn, ",")
	if len(parts) == 0 {
		return ""
	}

	// First part should be "uid=value"
	firstPart := parts[0]
	if strings.HasPrefix(strings.ToLower(firstPart), "uid=") {
		return strings.TrimPrefix(firstPart, "uid=")
	}

	return ""
}
