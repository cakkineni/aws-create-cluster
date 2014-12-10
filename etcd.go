package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"os"
)

var (
	etcdClient *etcd.Client
)

func init() {
	etcdIP := os.Getenv("ETCD_API")
	if etcdIP == "" {
		etcdIP = "172.17.42.1:4001"
	}
	etcdIP = fmt.Sprintf("http://%s", etcdIP)
	machines := []string{etcdIP}

	println(etcdIP)

	etcdClient = etcd.NewClient(machines)
}

func setEtcdKey(key string, value string) {
	println("Setting Etcd Key")
	_, err := etcdClient.Set(key, value, 0)

	if err != nil {
		fmt.Println(err)
	}
}
