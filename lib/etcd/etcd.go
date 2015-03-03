package etcd

import (
	"github.com/coreos/go-etcd/etcd"
)

type Client struct {
	*etcd.Client
}

func New(machines []string) (c *Client) {
	c = &Client{etcd.NewClient(machines)}
	return
}

// Get value at key.
// If it does not exist, set with default value and return default value.
func (c *Client) GetDefault(key, defaultValue string) (value string, err error) {
	resp, err := c.Client.Get(key, false, false)
	if e, ok := err.(*etcd.EtcdError); ok {
		switch e.ErrorCode {
		case 100:
			resp, err = c.Client.Set(key, defaultValue, 0)
		}
	}
	if err != nil {
		return
	}
	value = resp.Node.Value
	return
}
