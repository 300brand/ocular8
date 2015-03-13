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
	CanEdit bool
	Changed chan bool `json:"-"`
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

func (c *Client) GetList() (items []*Item, err error) {
	resp, err := c.Get("/", true, true)
	if err != nil {
		return
	}
	items = make([]*Item, 0, 64)
	nodes := []*etcd.Node{resp.Node}
	for i := 0; i < len(nodes); i++ {
		if nodes[i].Dir {
			// recursion without the insanity
			nodes = append(nodes, nodes[i].Nodes...)
			continue
		}

		item := &Item{
			Key:   nodes[i].Key,
			Value: nodes[i].Value,
		}
		items = append(items, item)

		dir, base := filepath.Split(nodes[i].Key)
		// Pull out super secret hidden default value
		defaultKey := filepath.Join(dir, "_"+base+".default")
		if resp, err = c.Get(defaultKey, false, false); err != nil {
			if e, ok := err.(*etcd.EtcdError); ok && e.ErrorCode == 100 {
				err = nil
				continue
			}
			return
		}
		item.Default = resp.Node.Value

		// Pull out super secret hidden description
		descKey := filepath.Join(dir, "_"+base+".desc")
		if resp, err = c.Get(descKey, false, false); err != nil {
			if e, ok := err.(*etcd.EtcdError); ok && e.ErrorCode == 100 {
				err = nil
				continue
			}
			return
		}
		item.Desc = resp.Node.Value
		item.CanEdit = true
	}
	return
}

func (c *Client) GetAll(kv map[string]*string) (err error) {
	for key, value := range kv {
		resp, err := c.Get(key, false, false)
		if err != nil {
			return err
		}
		*value = resp.Node.Value
	}
	return
}
