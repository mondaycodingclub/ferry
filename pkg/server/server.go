package server

import (
	"encoding/json"
	"ferry/api"
	"ferry/util"
	"fmt"
	"github.com/google/uuid"
	"io"
	"net"
	"net/http"
	"time"
)

//var nodes = []api.Node{
//	{
//		Meta: api.Meta{
//			Name: "node1",
//		},
//		Spec:   api.NodeSpec{},
//		Status: api.NodeStatus{},
//	},
//}

var proxies = map[string]ProxyInfo{
	"proxy1": {
		Listener: nil,
		Proxy: api.Proxy{
			Meta: api.Meta{
				Name: "proxy1",
			},
			Spec: api.ProxySpec{
				NodeName:  "node1",
				InnerHost: "",
				InnerPort: "8081",
				OuterHost: "123.207.28.113",
				OuterPort: "8081",
			},
			Status: api.ProxyStatus{},
		},
	},
	"proxy2": {
		Listener: nil,
		Proxy: api.Proxy{
			Meta: api.Meta{
				Name: "proxy2",
			},
			Spec: api.ProxySpec{
				NodeName:  "node1",
				InnerHost: "192.168.31.158",
				InnerPort: "8080",
				OuterHost: "123.207.28.113",
				OuterPort: "8080",
			},
			Status: api.ProxyStatus{},
		},
	},
}

type NodeInfo struct {
	Conn net.Conn `json:"conn"`
	Node api.Node `json:"node"`
}

type ProxyInfo struct {
	Listener net.Listener `json:"listener"`
	Proxy    api.Proxy    `json:"proxy"`
}

type TunnelInfo struct {
	Src    net.Conn   `json:"src"`
	Dst    net.Conn   `json:"dst"`
	Tunnel api.Tunnel `json:"tunnel"`
}

type Server struct {
	nodes   map[string]NodeInfo
	proxies map[string]ProxyInfo
	tunnels map[string]TunnelInfo
}

func NewServer() *Server {

	return &Server{
		nodes:   make(map[string]NodeInfo),
		proxies: proxies,
		tunnels: make(map[string]TunnelInfo),
	}
}

func (s *Server) Run() {
	go s.nodeManager()
	go s.proxyManager()
	go s.tunnelManager()
	s.APIManager()
}

func (s *Server) nodeManager() {
	listen, err := net.Listen(util.TCP, fmt.Sprintf(":%d", 8020))
	if err != nil {
		return
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			return
		}
		content, err := util.Read(conn, util.DELIMITER)
		if err != nil {
			continue
		}
		var event api.Event
		err = json.Unmarshal([]byte(content), &event)
		if event.Type != "Node" || event.Action != api.ADD {
			conn.Close()
			fmt.Println("not RegistryNode")
			continue
		}
		node := event.Node
		fmt.Println("on RegistryNode", node.Name)

		info, ok := s.nodes[node.Name]
		if ok {
			info.Conn.Close()
			delete(s.nodes, node.Name)
		}

		s.nodes[node.Name] = NodeInfo{
			Conn: conn,
			Node: node,
		}

		go util.KeepAlive(conn)
	}
}

func (s *Server) proxyManager() {
	for {
		for _, proxyInfo := range s.proxies {
			time.Sleep(time.Second * 3)
			nodeName := proxyInfo.Proxy.Spec.NodeName
			nodeInfo, ok := s.nodes[nodeName]
			if !ok {
				fmt.Println("Node not found")
				continue
			}
			conn := nodeInfo.Conn
			event := api.Event{
				Type:   "Proxy",
				Action: api.ADD,
				Proxy:  proxyInfo.Proxy,
			}
			content, err := json.Marshal(event)
			if err != nil {
				return
			}
			_, err = util.Write(conn, string(content))
			if err != nil {
				return
			}
			listener := proxyInfo.Listener
			if listener != nil {
				return
			}

			listen, err := net.Listen(util.TCP, fmt.Sprintf(":%s", proxyInfo.Proxy.Spec.OuterPort))
			if err != nil {
				return
			}

			proxyInfo.Listener = listen

			go func(listen net.Listener, nodeName, proxyName string) {
				for {
					conn, err := listen.Accept()
					if err != nil {
						return
					}
					fmt.Println("outer connected", nodeName, proxyName, conn.RemoteAddr().String())

					tunnelName := uuid.New().String()
					s.tunnels[tunnelName] = TunnelInfo{
						Src: conn,
						Tunnel: api.Tunnel{
							Meta: api.Meta{
								Name: tunnelName,
							},
							Spec: api.TunnelSpec{
								NodeName:  nodeName,
								ProxyName: proxyName,
							},
							Status: api.TunnelStatus{},
						},
					}
					go s.invokeTunnel(tunnelName)
				}
			}(listen, nodeInfo.Node.Name, proxyInfo.Proxy.Name)

		}
	}
}

func (s *Server) tunnelManager() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", 8021))
	if err != nil {
		return
	}

	fmt.Println("Tunnel Listened")

	for {
		conn, err := listen.Accept()
		if err != nil {
			continue
		}

		content, err := util.Read(conn, util.DELIMITER)
		if err != nil {
			continue
		}
		var event api.Event
		err = json.Unmarshal([]byte(content), &event)

		if event.Type != "Tunnel" {
			conn.Close()
			continue
		}
		tunnel := event.Tunnel

		switch event.Action {
		case api.MODIFY:
			fmt.Println(tunnel.Name, "modified")
			tunnelInfo := s.tunnels[tunnel.Name]
			tunnelInfo.Dst = conn
			util.Copy(tunnelInfo.Src, tunnelInfo.Dst)
		case api.DELETE:
			tunnelInfo := s.tunnels[tunnel.Name]
			tunnelInfo.Src.Close()
			delete(s.tunnels, tunnel.Name)
			conn.Close()
		default:
			conn.Close()
		}
	}
}

func (s *Server) invokeTunnel(tunnelName string) {
	tunnelInfo := s.tunnels[tunnelName]
	ctrl := api.Event{
		Type:   "Tunnel",
		Action: api.ADD,
		Tunnel: tunnelInfo.Tunnel,
	}
	content, err := json.Marshal(ctrl)
	if err != nil {
		return
	}
	conn := s.nodes[tunnelInfo.Tunnel.Spec.NodeName].Conn
	_, err = util.Write(conn, string(content))
	if err != nil {
		return
	}
}

func (s *Server) APIManager() {
	http.HandleFunc("/all", s.listAll)
	http.ListenAndServe(":8023", nil)
}

func (s *Server) listAll(response http.ResponseWriter, request *http.Request) {
	data := make(map[string]interface{})
	data["nodes"] = s.nodes
	data["proxies"] = s.proxies
	data["tunnels"] = s.tunnels
	bytes, err := json.Marshal(data)
	if err != nil {
		io.WriteString(response, err.Error())
	}
	io.WriteString(response, string(bytes))
}

func (s *Server) addProxy(proxy api.Proxy) {
	exist, ok := s.proxies[proxy.Name]
	if ok {
		exist.Proxy = proxy
		return
	}
	s.proxies[proxy.Name] = ProxyInfo{
		Listener: nil,
		Proxy:    proxy,
	}
}
