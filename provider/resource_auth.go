package provider

import (
	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAuth() *schema.Resource {
	return &schema.Resource{
		Description: "Controls the authentication status (enabled or disabled) of the etcd cluster.",
		Create:      resourceAuthUpsert,
		Read:        resourceAuthRead,
		Delete:      resourceAuthDelete,
		Update:      resourceAuthUpsert,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"enabled": {
				Description: "Enable or disable auth on etcd.",
				Type:        schema.TypeBool,
				Required:    true,
			},
		},
	}
}

func resourceAuthRead(d *schema.ResourceData, meta interface{}) error {
	cli := meta.(*client.EtcdClient)

	enabled, err := cli.GetAuthStatus()
	if err != nil {
		return err
	}

	d.Set("enabled", enabled)

	return nil
}

func resourceAuthUpsert(d *schema.ResourceData, meta interface{}) error {
	enabledVar, _ := d.GetOk("enabled")
	enabled := enabledVar.(bool)

	cli := meta.(*client.EtcdClient)

	err := cli.SetAuthStatus(enabled)
	if err != nil {
		return err
	}

	d.SetId("auth")
	return resourceAuthRead(d, meta)
}

func resourceAuthDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
