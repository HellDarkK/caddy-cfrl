package cloudflare

import (
	"context"
	"errors"

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

// RemoveIP removes a specific IP from the Cloudflare IP List.
func (c *Client) RemoveIP(ctx context.Context, ip string) error {
	items, err := c.api.ListIPListItems(ctx, c.accountID, c.listID)
	if err != nil {
		return err
	}
	var deleteReq cloudflare.IPListItemDeleteRequest
	for _, item := range items {
		if item.IP == ip {
			deleteReq.Items = append(deleteReq.Items, cloudflare.IPListItemDeleteItemRequest{ID: item.ID})
		}
	}
	if len(deleteReq.Items) == 0 {
		return errors.New("no matching IP found in Cloudflare IP list")
	}
	_, err = c.api.DeleteIPListItems(ctx, c.accountID, c.listID, deleteReq)
	return err
}
