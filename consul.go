package main

import (
	"github.com/hashicorp/consul/api"
	"github.com/thehivecorporation/log"
	"strconv"
)

func putConsulKey(config *Config, key string, value []byte) {
	consulConfig := &api.Config{}
	consulConfig.Address = config.ConsulHost + ":" + strconv.FormatInt(config.ConsulPort, 10)
	// Get a new client
	client, err := api.NewClient(consulConfig)
	if err != nil {
		log.WithError(err).Fatal("Error instantiating consul client")
		panic(err)
	}
	// Get a handle to the KV API
	kv := client.KV()
	// PUT a  KV pair
	p := &api.KVPair{Key: key, Value: []byte(value)}
	_, err = kv.Put(p, nil)
	if err != nil {
		log.WithError(err).Fatal("Error saving key value pair into consul KV")
		panic(err)
	}
	log.Debug("Consul saved " + key + ":" + string(value))
}
