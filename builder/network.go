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

// share hosts network - lxc.network.type = none
type NoneType struct{}

func (t *NoneType) Configure(b *NetworkBuilder) error {
	b.c.SetConfigItem("lxc.network.type", "none")
	return nil
}

// empty network
type EmptyType struct {
}

func (t *EmptyType) Configure(b *NetworkBuilder) error {
	b.c.SetConfigItem("lxc.network.type", "empty")
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
	b.c.SetConfigItem("lxc.network.type", "veth")
	b.c.SetConfigItem("lxc.network.link", t.Link)
	b.c.SetConfigItem("lxc.network.flags", t.Flags)
	b.c.SetConfigItem("lxc.network.name", t.Name)
	b.c.SetConfigItem("lxc.network.hwaddr", t.Hwaddr)
	return nil
}

// vlan network
type VlanType struct {
	Link string
	Vlan int
}

func (t *VlanType) Configure(b *NetworkBuilder) error {
	b.c.SetConfigItem("lxc.network.type", "vlan")
	b.c.SetConfigItem("lxc.network.link", t.Link)
	b.c.SetConfigItem("lxc.network.vlan.id", fmt.Sprintf("%d", t.Vlan))
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
	b.c.SetConfigItem("lxc.network.type", "macvlan")
	b.c.SetConfigItem("lxc.network.link", t.Link)
	b.c.SetConfigItem("lxc.network.macvlan.mode", t.Mode)
	return nil
}

// phys network
type PhysType struct {
	Link string
}

func (t *PhysType) Configure(b *NetworkBuilder) error {
	b.c.SetConfigItem("lxc.network.type", "phys")
	b.c.SetConfigItem("lxc.network.link", t.Link)
	return nil
}
