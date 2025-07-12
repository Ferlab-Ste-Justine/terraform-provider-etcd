terraform {
  required_version = ">= 1.3.0"
  required_providers {
    etcd = {
      source  = "Ferlab-Ste-Justine/etcd"
      version = "1.0.0"
    }
  }
}