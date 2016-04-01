package app

import (
	"encoding/json"
	"log"
	"net/http"
	"io/ioutil"
	"golang.org/x/net/context"
)


// path in etcd to store container parameter
const  containerConfd string = "/app/container-param/"

// store container config to etcd
func (s *APIServer)SetContainerCfg(id string, param *Param) {
	value, _ := json.Marshal(*param)
	_, err := s.keysAPI.Set(context.Background(),
		containerConfd + id,
		string(value),
		nil,
	)
	if err != nil {
		log.Println("write etcd error:", err)
	}
}

// get container config from etcd
func (s *APIServer)ContainerCfg(id string) (*Param) {
	resp, err := s.keysAPI.Get(context.Background(),
		containerConfd + id,
		nil,
	)
	if err != nil {
		log.Println("read etcd error:", err)
		return nil
	}
	var param Param
	json.Unmarshal([]byte(resp.Node.Value), &param)
	return &param
}

// delete container config in etcd
func (s *APIServer)RmvContainerCfg(id string) {
	_, err := s.keysAPI.Delete(context.Background(),
		containerConfd + id,
		nil,
	)
	if err != nil {
		log.Println("delete in etcd error:", err)
	}
}

// service configuration request handle, update service config in etcd
func (s *APIServer) SetConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var cf ServConf
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{})
		log.Println("read http request error:", err)
		return
	}
	json.Unmarshal(b, &cf)
	value, _ := json.Marshal(cf.Conf)
	_, err = s.keysAPI.Set(context.Background(),
			"/app/"+cf.Node+"/"+cf.Service,
			string(value),
			nil,
	)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Println("write etcd error:", err)
	}else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write(b)
}

// remove config record from etcd
func (s *APIServer) RmvConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var cf ServConf
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte{})
		log.Println("read http request error:", err)
		return
	}
	json.Unmarshal(b, &cf)
	_, err = s.keysAPI.Delete(context.Background(),
		"/app/"+cf.Node+"/"+cf.Service,
		nil,
	)
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		log.Println("delete etcd record error:", err)
	}else {
		w.WriteHeader(http.StatusOK)
	}
	w.Write(b)
}



