package provider

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdConnection struct {
	Client  *clientv3.Client
	Timeout int
	Retries int
}