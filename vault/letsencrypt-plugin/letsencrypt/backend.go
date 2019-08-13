package letsencrypt

import (
	"context"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
	"github.com/pkg/errors"
)

type tls struct {
	Cert string `json:"cert"`
	Key string `json:"key"`
	Ttl int `json:"ttl"`
}

// Factory creates a new usable instance of this secrets engine.
func Factory(ctx context.Context, c *logical.BackendConfig) (logical.Backend, error) {
	b := Backend(c)
	if err := b.Setup(ctx, c); err != nil {
		return nil, errors.Wrap(err, "failed to create factory")
	}
	return b, nil
}

// backend is the actual backend.
type backend struct {
	*framework.Backend
}

// Backend creates a new backend.
func Backend(c *logical.BackendConfig) *backend {
	var b backend

	b.Backend = &framework.Backend{
		BackendType: logical.TypeLogical,
		Help: `

The letsencrypt secrets acts as a front-end for generating and retrieving certbot
certificates.

		`,
		Paths: []*framework.Path{
			&framework.Path{
				Pattern:      "info",
				HelpSynopsis: "Display information about this plugin",
				HelpDescription: `

Displays information about the plugin, such as the plugin version and git commit.

`,
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation: b.pathInfo,
				},
			},

			&framework.Path{
				Pattern:      "cert/" + framework.GenericNameRegex("domain"),
				HelpSynopsis: "Retrieve a letsencrypt cert",
				HelpDescription: `

Return the letsencrypt cert for the given domain, generating it if it is not present.

`,
				Fields: map[string]*framework.FieldSchema{
					"domain": {
						Type:		framework.TypeString,
						Description:	"Domain for the cert",
					},
				},
				Callbacks: map[logical.Operation]framework.OperationFunc{
					logical.ReadOperation: b.pathCert,
				},
			},
		},
	}

	return &b
}
