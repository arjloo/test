package app

import (
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"strings"
	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/docker/engine-api/types"
)

const labelkey string = "com.docker.swarm.constraints"

func (s *APIServer) NewSwarmClient() (*client.Client) {
	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
	c, err := client.NewClient(s.swarmAddr, "v1.22", nil, defaultHeaders)
	if err != nil {
		log.Println("new swarm client failed: ", err)
		return nil
	}
	return c
}

// transform input parameter into container config
func (s *APIServer) Parse (param *Param) (*container.Config, *container.HostConfig) {
	// gen container config
	cons := make([]string, len(param.Constraints), cap(param.Constraints))
	for i := range param.Constraints {
		cons[i] = "constraint:" + param.Constraints[i]
	}
	cfg := &container.Config{Image: param.Image}
	if param.Constraints != nil {
		cfg.Env = cons
	}

	// gen container host config
	hostCfg := new(container.HostConfig)
	if param.NetworkMode != "" {
		hostCfg.NetworkMode = container.NetworkMode(param.NetworkMode)
	}
	if param.PortMaps != nil {
		m := make(nat.PortMap)
		for _, it := range param.PortMaps {
			if v, ok := m[nat.Port(it.SrcPort)]; ok {
				m[nat.Port(it.SrcPort)] = append(v, nat.PortBinding{HostPort: it.DstPort})
			}else {
				m[nat.Port(it.SrcPort)] = []nat.PortBinding{nat.PortBinding{HostPort: it.DstPort}}
			}
		}
		hostCfg.PortBindings = m
	}
	if param.Volumes != nil {
		hostCfg.Binds = param.Volumes
	}

	return cfg, hostCfg
}

func (s *APIServer) CreateServHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cli := s.NewSwarmClient()
	if cli == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte{})
		return
	}

	var param Param
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{})
		log.Println(err)
		return
	}

	json.Unmarshal(b, &param)
	cfg, hostcfg := s.Parse(&param)
	c, err := cli.ContainerCreate(context.Background(), cfg, hostcfg, nil, "")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte{})
		log.Println(err)
		return
	}else {
		cli.ContainerStart(context.Background(), c.ID)
	}
	s.SetContainerCfg(c.ID, &param)
	cont := ContInfo{ID: c.ID}
	j, _ := json.Marshal(cont)
	w.Write(j)
	w.WriteHeader(http.StatusOK)
}

func (s *APIServer) DeleteServHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cli := s.NewSwarmClient()
	if cli == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte{})
		return
	}
	var param ContInfo
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{})
		log.Println(err)
		return
	}
	json.Unmarshal(b, &param)
	e := cli.ContainerRemove(context.Background(), types.ContainerRemoveOptions{ContainerID: param.ID, Force: true})
	if e != nil {
		log.Println("container remove failed:", param.ID)
		w.WriteHeader(http.StatusBadRequest)
	}else {
		s.RmvContainerCfg(param.ID)
		w.WriteHeader(http.StatusOK)
	}
	w.Write(b)
}

// get containers satisfying conditions
func Containers2Update(client *client.Client, cond *Condition) ([]string, error) {
	opts := types.ContainerListOptions{All: true}
	containers, err := client.ContainerList(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	var set []string
	img := strings.Split(cond.Image, ":")[0]
	for _, c := range containers {
		if !strings.Contains(c.Image, img) {
			continue
		}
		ok := true
		for _, l := range cond.Labels {
			str := c.Labels[labelkey]
			if !strings.Contains(str, l) {
				ok = false
				break
			}
		}
		if ok {
			set = append(set, c.ID)
		}
	}
	return set, nil
}

func (s *APIServer) UpdateServHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cli := s.NewSwarmClient()
	if cli == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte{})
		return
	}
	var cond Condition
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{})
		log.Println(err)
		return
	}
	json.Unmarshal(b, &cond)

	containers, err := Containers2Update(cli, &cond)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte{})
		return
	}

	var conts ContSet
	for _, id := range containers {
		err := cli.ContainerRemove(context.Background(), types.ContainerRemoveOptions{ContainerID: id, Force: true})
		if err != nil {
			log.Println("container remove failed:", id)
			continue
		}

		cfg4update := s.ContainerCfg(id)
		cfg4update.Image = cond.Image
		cfg, hostcfg := s.Parse(cfg4update)
		s.RmvContainerCfg(id)
		c, err := cli.ContainerCreate(context.Background(), cfg, hostcfg, nil, "")
		if err != nil {
			log.Println("container update failed:", id, err)
			continue
		}else {
			cli.ContainerStart(context.Background(), c.ID)
		}

		s.SetContainerCfg(c.ID, cfg4update)

		cont := ContInfo{ID: c.ID}
		conts.Containers = append(conts.Containers, cont)
	}
	j, _ := json.Marshal(conts)
	w.Write(j)
	w.WriteHeader(http.StatusOK)
}

