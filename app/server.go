package app

import (
	"log"
	"time"

	etcd "github.com/coreos/etcd/client"
)

type APIServer struct {
	sysID       string
	swarmAddr   string
	keysAPI     etcd.KeysAPI
}

type PortMap struct {
	SrcPort string  `json:"srcport"`
	DstPort string  `json:"dstport"`
}

// input parameter for container creation
type Param struct {
    Image       string      `json:"image"`
	Constraints []string    `json:"labels"`
	PortMaps    []PortMap   `json:"portmaps"`
	Volumes     []string    `json:"volumes"`
	NetworkMode string      `json:"netmode"`
}

// container update condition
type Condition struct {
	Labels      []string    `json:"labels"`
	Image       string      `json:"image"`
}

// container info
type ContInfo struct {
	//name    string      `json:"name"`
	ID      string      `json:"id"`
	//image   string      `json:"image"`
}

// container set
type ContSet struct {
	Containers []ContInfo  `json:"containers"`
}

// service configuration stored in etcd
type ServConf struct {
	Node    string      `json:"node"`
	Service string      `json:"service"`
	Conf    interface{} `json:"config"`
}

func NewServer() (*APIServer) {
	etcdCfg := etcd.Config{
		Endpoints:  []string{"http://16.27.200.244:4001"},
		Transport:  etcd.DefaultTransport,
		HeaderTimeoutPerRequest:    time.Second,
	}
	etcdClient, err := etcd.New(etcdCfg)
	if err != nil {
		log.Println("Failed to connect to etcd:", err)
		return nil
	}
	return &APIServer{
		swarmAddr: "tcp://16.27.200.249:2376",
		keysAPI: etcd.NewKeysAPI(etcdClient),
	}
}


