package cloudflare

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

type Client struct {
	api       *cloudflare.API
	accountID string
	listID    string
}

func NewClient(apiToken, accountID, listID string) (*Client, error) {
	api, err := cloudflare.NewWithAPIToken(apiToken)
	if err != nil {
		return nil, err
	}
	return &Client{
		api:       api,
		accountID: accountID,
		listID:    listID,
	}, nil
}

func (c *Client) BlockIP(ctx context.Context, ip, comment string) error {
	_, err := c.api.CreateIPListItems(ctx, c.accountID, c.listID, []cloudflare.IPListItemCreateRequest{
		{
			IP:      ip,
			Comment: comment,
		},
	})
	return err
}
