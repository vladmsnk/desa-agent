package adapters

import (
	"fmt"

	"desa-agent/internal/adapters/ad"
	"desa-agent/internal/adapters/ldap"
	"desa-agent/internal/config"
)

func NewIdentityProvider(cfg config.IDPConfig) (IdentityProvider, error) {
	switch cfg.Type {
	case config.IdentityProviderTypeActiveDirectory:
		return ad.New(cfg)
	case config.IdentityProviderTypeLDAP:
		return ldap.New(cfg)
	default:
		return nil, fmt.Errorf("unsupported identity provider type: %s", cfg.Type)
	}
}
