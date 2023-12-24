package api

type Meta struct {
	Name string
}

type NodeSpec struct {
}

type NodeStatus struct {
}

type Node struct {
	Meta
	Spec   NodeSpec
	Status NodeStatus
}

type ProxySpec struct {
	NodeName  string
	InnerHost string
	InnerPort string
	OuterHost string
	OuterPort string
}

type ProxyStatus struct {
}

type Proxy struct {
	Meta
	Spec   ProxySpec
	Status ProxyStatus
}

type TunnelSpec struct {
	NodeName  string
	ProxyName string
}

type TunnelStatus struct {
}

type Tunnel struct {
	Meta
	Spec   TunnelSpec
	Status TunnelStatus
}

type CtrlType string

const (
	RegistryNode       CtrlType = "RegistryNode"
	SyncNode           CtrlType = "SyncNode"
	SyncProxy          CtrlType = "SyncProxy"
	InvokeTunnel       CtrlType = "InvokeTunnel"
	CreateTunnel       CtrlType = "CreateTunnel"
	CreateTunnelFailed CtrlType = "CreateTunnelFailed"
)

type Ctrl struct {
	Type      CtrlType
	Node      Node
	Proxy     Proxy
	NodeName  string
	ProxyName string
	TunnelID  string
}

type Action string

const (
	ADD    Action = "Add"
	MODIFY Action = "Modify"
	DELETE Action = "Delete"
)

type Event struct {
	Type   string
	Action Action
	Node   Node
	Proxy  Proxy
	Tunnel Tunnel
}
