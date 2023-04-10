package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	schema.DescriptionKind = schema.StringMarkdown
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Description: "Name of the etcd user that will be used to access etcd. Can alternatively be set with the ETCDCTL_USERNAME environment variable. Can also be omitted if tls certificate authentication will be used instead as the username will be infered from the certificate.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_USERNAME", ""),
			},
			"password": &schema.Schema{
				Description: "Password of the etcd user that will be used to access etcd. Can alternatively be set with the ETCDCTL_PASSWORD environment variable. Can also be omitted if tls certificate authentication will be used instead.",
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_PASSWORD", ""),
			},
			"ca_cert": &schema.Schema{
				Description: "File that contains the CA certificate that signed the etcd servers' certificates. Can alternatively be set with the ETCDCTL_CACERT environment variable. Can also be omitted.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_CACERT", ""),
			},
			"cert": &schema.Schema{
				Description: "File that contains the client certificate used to authentify the user. Can alternatively be set with the ETCDCTL_CERT environment variable. Can be omitted if password authentication is used.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_CERT", ""),
			},
			"key": &schema.Schema{
				Description: "File that contains the client encryption key used to authentify the user. Can alternatively be set with the ETCDCTL_KEY environment variable. Can be omitted if password authentication is used.",
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_KEY", ""),
			},
			"endpoints": &schema.Schema{
				Description: "Endpoints of the etcd servers. The entry of each server should follow the ip:port format and be coma separated. Can alternatively be set with the ETCDCTL_ENDPOINTS environment variable.",
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ETCDCTL_ENDPOINTS", ""),
			},
			"connection_timeout": &schema.Schema{
				Description: "Timeout to establish the etcd servers connection as a duration. Defaults to 10s.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "10s",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					_, err := time.ParseDuration(v)
					if err != nil {
						return []string{}, []error{errors.New("connection_timeout must be a value golang duration string value")}
					}

					return []string{}, []error{}
				},
			},
			"request_timeout": &schema.Schema{
				Description: "Timeout for individual requests the provider makes on the etcd servers as a duration. Defaults to 10s.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "10s",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					_, err := time.ParseDuration(v)
					if err != nil {
						return []string{}, []error{errors.New("request_timeout must be a value golang duration string value")}
					}

					return []string{}, []error{}
				},
			},
			"retry_interval": &schema.Schema{
				Description: "Duration to wait after a failing etcd request before retrying. Defaults to 100ms.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "100ms",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(string)
					_, err := time.ParseDuration(v)
					if err != nil {
						return []string{}, []error{errors.New("retry_interval must be a value golang duration string value")}
					}

					return []string{}, []error{}
				},
			},
			"retries": &schema.Schema{
				Description: "Number of times operations that result in retriable errors should be re-attempted. Defaults to 10.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"etcd_role":                      resourceRole(),
			"etcd_user":                      resourceUser(),
			"etcd_key":                       resourceKey(),
			"etcd_range_scoped_state":        resourceRangeScopedState(),
			"etcd_synchronized_key_prefixes": resourceSynchronizedKeyPrefixes(),
			"etcd_synchronized_directory":    resourceSynchronizedDirectory(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"etcd_prefix_range_end": dataSourcePrefixRangeEnd(),
			"etcd_key_range":        dataSourceKeyRange(),
		},
		ConfigureFunc: providerConfigure,
		//Should implement close once this issue is resolved: https://github.com/hashicorp/terraform-plugin-sdk/issues/63
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	endpoints, _ := d.Get("endpoints").(string)
	username, _ := d.Get("username").(string)
	password, _ := d.Get("password").(string)
	caCert, _ := d.Get("ca_cert").(string)
	cert, _ := d.Get("cert").(string)
	key, _ := d.Get("key").(string)
	connectionTimeout, _ := d.Get("connection_timeout").(string)
	requestTimeout, _ := d.Get("request_timeout").(string)
	retryInterval, _ := d.Get("retry_interval").(string)
	retries, _ := d.Get("retries").(int)

	connectionTimeoutDuration, _ := time.ParseDuration(connectionTimeout)
	requestTimeoutDuration, _ := time.ParseDuration(requestTimeout)
	retryIntervalDuration, _ := time.ParseDuration(retryInterval)

	cli, cliErr := client.Connect(context.Background(), client.EtcdClientOptions{
		EtcdEndpoints:     strings.Split(endpoints, ","),
		Username:          username,
		Password:          password,
		ClientCertPath:    cert,
		ClientKeyPath:     key,
		CaCertPath:        caCert,
		ConnectionTimeout: connectionTimeoutDuration,
		RequestTimeout:    requestTimeoutDuration,
		RetryInterval:     retryIntervalDuration,
		Retries:           uint64(retries),
	})

	if cliErr != nil {
		return nil, errors.New(fmt.Sprintf("Failed to connect to etcd servers: %s", cliErr.Error()))
	}

	return cli, nil
}
