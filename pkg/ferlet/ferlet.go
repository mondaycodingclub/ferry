package ferlet

import (
	"encoding/json"
	"ferry/api"
	"ferry/util"
	"fmt"
	"net"
	"time"
)

type Ferlet struct {
	node    *api.Node
	proxies map[string]api.Proxy

	serverHost string
	serverPort string
	serverAddr string
	tunnelPort string
}

func NewFerlet(config Config) *Ferlet {
	node := &api.Node{
		Meta: api.Meta{
			Name: config.Name,
		},
	}

	return &Ferlet{
		node:    node,
		proxies: make(map[string]api.Proxy),

		serverHost: config.ServerHost,
		serverPort: config.ServerPort,
		serverAddr: net.JoinHostPort(config.ServerHost, config.ServerPort),
		tunnelPort: "8021",
	}
}

func (a *Ferlet) Run() {
	for {
		a.connectServer()
		time.Sleep(time.Second * 3)
	}

}

func (a *Ferlet) connectServer() {
	conn, err := net.Dial(util.TCP, a.serverAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	fmt.Println("server connected")
	go a.register(conn)
	for {
		content, err := util.Read(conn, util.DELIMITER)
		if err != nil {
			return
		}
		var ctrl api.Event
		err = json.Unmarshal([]byte(content), &ctrl)
		switch ctrl.Type {
		case "Proxy":
			switch ctrl.Action {
			case api.ADD:
				a.syncProxy(ctrl.Proxy)
				//case api.DELETE:
				//	delete(a.proxies, ctrl.Proxy.)
			}
		case "Tunnel":
			switch ctrl.Action {
			case api.ADD:
				a.invokeTunnel(ctrl.Tunnel)
			}

		default:
			fmt.Println(ctrl.Type)
		}
	}
}

func (a Ferlet) register(conn net.Conn) {
	ctrl := api.Event{
		Type:   "Node",
		Action: api.ADD,
		Node:   *a.node,
	}
	content, err := json.Marshal(ctrl)
	if err != nil {
		return
	}
	_, err = util.Write(conn, string(content))
	if err != nil {
		return
	}
}

func (a *Ferlet) syncProxy(proxy api.Proxy) {
	_, ok := a.proxies[proxy.Name]
	if ok {
		delete(a.proxies, proxy.Name)
	}
	a.proxies[proxy.Name] = proxy
	fmt.Println("proxy synced", proxy)
}

func (a *Ferlet) invokeTunnel(tunnel api.Tunnel) {
	proxy, ok := a.proxies[tunnel.Spec.ProxyName]
	if !ok {
		fmt.Println("proxy not found")
		a.closeTunnel(tunnel)
		return
	}
	a.createTunnel(proxy.Spec, tunnel)
}

func (a Ferlet) createTunnel(proxy api.ProxySpec, tunnel api.Tunnel) {
	src, err := net.Dial(util.TCP, fmt.Sprintf("%s:%s", a.serverHost, a.tunnelPort))
	if err != nil {
		fmt.Println("tunnel connect failed", err)
		return
	}
	fmt.Println("tunnel connected")

	ctrl := api.Event{
		Type:   "Tunnel",
		Action: api.MODIFY,
		Tunnel: tunnel,
	}
	content, err := json.Marshal(ctrl)
	if err != nil {
		return
	}
	_, err = util.Write(src, string(content))
	if err != nil {
		return
	}

	dst, err := net.Dial(util.TCP, fmt.Sprintf("%s:%s", proxy.InnerHost, proxy.InnerPort))
	if err != nil {
		return
	}
	fmt.Println("local connected")
	util.Copy(src, dst)
}

func (a Ferlet) closeTunnel(tunnel api.Tunnel) {
	src, err := net.Dial(util.TCP, fmt.Sprintf("%s:%s", a.serverHost, a.tunnelPort))
	if err != nil {
		return
	}

	ctrl := api.Event{
		Type:   "Tunnel",
		Action: api.DELETE,
		Tunnel: tunnel,
	}
	content, err := json.Marshal(ctrl)
	if err != nil {
		return
	}
	_, err = util.Write(src, string(content))
	if err != nil {
		return
	}
}
