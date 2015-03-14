package builder

import (
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
)

type NetworkBuilder struct {
	c *lxc.Container
}

func NewNetworkBuilder(c *lxc.Container) *NetworkBuilder {
	return &NetworkBuilder{
		c: c,
	}
}

// simple way to add a piped interface using a veth named eth0
func (b *NetworkBuilder) AddPInterfaceVeth(name string) *NetworkBuilder {
	veth := DefaultVethType()
	veth.Name = name
	b.AddInterface(veth)
	return b
}

type InterfaceBuilder interface {
	Configure(b *NetworkBuilder) error
}

func (b *NetworkBuilder) AddInterface(ifdef InterfaceBuilder) *NetworkBuilder {
	err := ifdef.Configure(b)
	if err != nil {
		panic(err)
	}
	return b
}
func (b *NetworkBuilder) SetConfigItem(key, value string) error {
	log.Debugf("SetConfig %s = %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("SetConfig %s = %s: %s", key, value, err)
	}
	return err
}

// handles setting common network parameters
type NetworkConfig struct {
	// CIDR format
	IPv4 IPv4Config
	IPv6 IPv6Config
}
type IPv4Config struct {
	// IP in CDIR
	IP string

	// gateway
	Gateway string
}
type IPv6Config struct {
	IP      string
	Gateway string
}

func (b *NetworkBuilder) WithNetworkConfig(networkConfig *NetworkConfig) *NetworkBuilder {
	err := networkConfig.Configure(b)
	if err != nil {
		panic(err)
	}
	return b
}

func (t *NetworkConfig) Configure(b *NetworkBuilder) error {
	v4 := t.IPv4
	if len(v4.IP) > 0 {
		b.SetConfigItem("lxc.network.ipv4", v4.IP)
	}
	if len(v4.Gateway) > 0 {
		b.SetConfigItem("lxc.network.ipv4.gateway", v4.Gateway)
	}
	v6 := t.IPv6
	if len(v6.IP) > 0 {
		b.SetConfigItem("lxc.network.ipv6", v6.IP)
	}
	if len(v6.Gateway) > 0 {
		b.SetConfigItem("lxc.network.ipv6.gateway", v6.Gateway)
	}
	return nil
}

// share hosts network - lxc.network.type = none
type NoneType struct{}

func (t *NoneType) Configure(b *NetworkBuilder) error {
	b.SetConfigItem("lxc.network.type", "none")
	return nil
}

// empty network
type EmptyType struct {
}

func (t *EmptyType) Configure(b *NetworkBuilder) error {
	b.SetConfigItem("lxc.network.type", "empty")
	return nil
}

// veth network
type VethType struct {
	Link   string
	Flags  string
	Name   string
	Hwaddr string
}

func DefaultVethType() *VethType {
	return &VethType{
		Link:   "lxcbr0",
		Flags:  "up",
		Name:   "eth0",
		Hwaddr: "00:16:3e:xx:xx:xxeth0",
	}
}
func (t *VethType) Configure(b *NetworkBuilder) error {
	b.SetConfigItem("lxc.network.type", "veth")
	b.SetConfigItem("lxc.network.link", t.Link)
	b.SetConfigItem("lxc.network.flags", t.Flags)
	b.SetConfigItem("lxc.network.name", t.Name)
	b.SetConfigItem("lxc.network.hwaddr", t.Hwaddr)
	return nil
}

// vlan network
type VlanType struct {
	Link string
	Vlan int
}

func (t *VlanType) Configure(b *NetworkBuilder) error {
	b.SetConfigItem("lxc.network.type", "vlan")
	b.SetConfigItem("lxc.network.link", t.Link)
	b.SetConfigItem("lxc.network.vlan.id", fmt.Sprintf("%d", t.Vlan))
	return nil
}

// macvlan network
type MacVlanType struct {
	Link string
	Mode string
}

func (t *MacVlanType) BridgeMode() *MacVlanType {
	t.Mode = "bridge"
	return t
}
func (t *MacVlanType) PrivateMode() *MacVlanType {
	t.Mode = "private"
	return t
}
func (t *MacVlanType) VepaMode() *MacVlanType {
	t.Mode = "vepa"
	return t
}

func (t *MacVlanType) Configure(b *NetworkBuilder) error {
	b.SetConfigItem("lxc.network.type", "macvlan")
	b.SetConfigItem("lxc.network.link", t.Link)
	b.SetConfigItem("lxc.network.macvlan.mode", t.Mode)
	return nil
}

// phys network
type PhysType struct {
	Link string
}

func (t *PhysType) Configure(b *NetworkBuilder) error {
	b.SetConfigItem("lxc.network.type", "phys")
	b.SetConfigItem("lxc.network.link", t.Link)
	return nil
}
