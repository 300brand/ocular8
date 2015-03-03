package etcd

import (
	"github.com/coreos/go-etcd/etcd"
	"path/filepath"
)

type Item struct {
	Key     string
	Default string
	Desc    string
	Value   string
}

type Client struct {
	*etcd.Client
}

func New(machine string) (c *Client) {
	c = &Client{etcd.NewClient([]string{machine})}
	return
}

// Get value at key.
// If it does not exist, set with default value and return default value.
func (c *Client) GetDefault(key, defaultValue, description string) (value string, err error) {
	resp, err := c.Client.Get(key, false, false)
	if e, ok := err.(*etcd.EtcdError); ok {
		switch e.ErrorCode {
		case 100:
			if resp, err = c.Client.Set(key, defaultValue, 0); err != nil {
				return
			}
			dir, base := filepath.Split(key)
			defaultKey := filepath.Join(dir, "_"+base+".default")
			descKey := filepath.Join(dir, "_"+base+".desc")
			if _, err = c.Client.Set(defaultKey, defaultValue, 0); err != nil {
				return
			}
			if _, err = c.Client.Set(descKey, description, 0); err != nil {
				return
			}
		}
	}
	if err != nil {
		return
	}
	value = resp.Node.Value
	return
}

func (c *Client) GetAll(items []*Item) (err error) {
	for _, item := range items {
		if item.Value, err = c.GetDefault(item.Key, item.Default, item.Desc); err != nil {
			return
		}
	}
	return
}
