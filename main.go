package main

import (
    "github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
    "ferlab/terraform-provider-etcd/provider"
)

func main() {
    plugin.Serve(&plugin.ServeOpts{
        ProviderFunc: provider.Provider,
    })
}