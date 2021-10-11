terraform {
  required_version = ">= 0.13"
  required_providers {
    etcd = {
      source  = "ferlab/etcd"
      version = "1.0.0"
    }
  }
}