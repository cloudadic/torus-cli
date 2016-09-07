package registry

import (
	"context"
	"errors"
	"log"
	"net/url"

	"github.com/arigatomachine/cli/envelope"
	"github.com/arigatomachine/cli/identity"
	"github.com/arigatomachine/cli/pathexp"
)

// CredentialTreeClient represents the `/credentialtree` registry endpoint,
// user for retrieving keyrings, keyring members, and credentials associated
// with claims.
type CredentialTreeClient struct {
	client *Client
}

// CredentialTree represents a Keyring, it's members, and associated
// Credentials.
type CredentialTree struct {
	Keyring     *envelope.Signed  `json:"keyring"`
	Members     []envelope.Signed `json:"members"`
	Credentials []envelope.Signed `json:"credentials"`
}

// Post creates a new CredentialTree on the registry.
//
// The CredentialTree includes the keyring, it's members, and credentials.
func (c *CredentialTreeClient) Post(ctx context.Context, t *CredentialTree) (*CredentialTree, error) {
	req, err := c.client.NewRequest("POST", "/credentialtree", nil, t)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := CredentialTree{}
	_, err = c.client.Do(ctx, req, &resp)
	if err != nil {
		log.Printf("Failed to create credential tree: %s", err)
		return nil, err
	}

	return &resp, nil
}

// List returns back all segments of the CredentialGraph (Keyring, Keyring
// Members, and Credentials) that match the given name, path, or path
// expression.
func (c *CredentialTreeClient) List(ctx context.Context, name, path string,
	pathExp *pathexp.PathExp, ownerID *identity.ID) ([]CredentialTree, error) {

	query := url.Values{}

	if path != "" && pathExp != nil {
		return nil, errors.New("cannot provide path and pathexp at the same time")
	}
	if path != "" {
		query.Set("path", path)
	}
	if pathExp != nil {
		query.Set("pathexp", pathExp.String())
	}
	if name != "" {
		query.Set("name", name)
	}
	if ownerID != nil {
		query.Set("owner_id", ownerID.String())
	}

	req, err := c.client.NewRequest("GET", "/credentialtree", &query, nil)
	if err != nil {
		log.Printf("Error building http request: %s", err)
		return nil, err
	}

	resp := []CredentialTree{}
	_, err = c.client.Do(ctx, req, &resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
