package provider

import (
    "crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    clientv3 "go.etcd.io/etcd/client/v3"
)


func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "username": &schema.Schema{
                Type: schema.TypeString,
                Optional: true,
				Sensitive: true,
                DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_USERNAME", ""),
            },
            "password": &schema.Schema{
                Type: schema.TypeString,
                Optional: true,
				Sensitive: true,
                DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_PASSWORD", ""),
            },
            "ca_cert": &schema.Schema{
                Type: schema.TypeString,
                Optional: true,
                DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_CACERT", ""),
            },
            "cert": &schema.Schema{
                Type: schema.TypeString,
                Optional: true,
                DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_CERT", ""),
            },
            "key": &schema.Schema{
                Type: schema.TypeString,
                Optional: true,
                DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_KEY", ""),
            },
            "endpoints": &schema.Schema{
                Type: schema.TypeString,
                Required: true,
                DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_ENDPOINTS", ""),
            },
        },
        ResourcesMap: map[string]*schema.Resource{
			"etcd_role": resourceRole(),
			"etcd_user": resourceUser(),
			"etcd_key": resourceKey(),
        },
        ConfigureFunc: providerConfigure,
		//Should implement close once this issue is resolved: https://github.com/hashicorp/terraform-plugin-sdk/issues/63
    }
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
    endpoints, _ := d.Get("endpoints").(string)
    username, _ :=  d.Get("username").(string)
    password, _ := d.Get("password").(string)
    caCert, _ := d.Get("ca_cert").(string)
    cert, _ := d.Get("cert").(string)
    key, _ := d.Get("key").(string)
    tlsConf := &tls.Config{}

    if cert != "" {
        certData, err := tls.LoadX509KeyPair(cert, key)
        if err != nil {
            return nil, err
        }
        (*tlsConf).Certificates = []tls.Certificate{certData}
        (*tlsConf).InsecureSkipVerify = false
    }

	if caCert != "" {
		caCertContent, err := ioutil.ReadFile(caCert) 
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Failed to read root certificate file: %s", err.Error()))
		}
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM(caCertContent)
		if !ok {
			return nil, errors.New("Failed to parse root certificate authority")
		}
		(*tlsConf).RootCAs = roots
	}

    return clientv3.New(clientv3.Config{
        Endpoints: strings.Split(endpoints, ","),
        Username: username,
        Password: password,
        TLS: tlsConf,
    })
}