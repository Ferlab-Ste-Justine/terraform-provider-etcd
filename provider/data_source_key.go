package provider

import (
	"errors"
	"fmt"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceKey() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about a key.",
		Read:        dataSourceKeyRead,
		Schema: map[string]*schema.Schema{
			"key": {
				Description:  "Key to retrieve.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"must_exist": &schema.Schema{
				Description: "Whether to cause an error if the key is not found.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"value": &schema.Schema{
				Description: "Value of the key.",
				Type:        schema.TypeString,
				Computed:    true,
			},
			"version": {
				Description: "Current version of the key. Note that version is reset to 0 on deletion",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"create_revision": {
				Description: "Revision of the etcd keystore when the key was created",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"mod_revision": {
				Description: "Revision of the etcd keystore when the key was last modified",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"lease": {
				Description: "Id of the lease that the key is attached to. Will be 0 if the key is not attached to a lease.",
				Type:        schema.TypeInt,
				Computed:    true,
			},
			"found": &schema.Schema{
				Description: "Whether the key was found.",
				Type:        schema.TypeBool,
				Computed:    true,
			},
		},
	}
}

func dataSourceKeyRead(d *schema.ResourceData, meta interface{}) error {
	cli := meta.(*client.EtcdClient)
	key := d.Get("key").(string)
	mustExist := d.Get("must_exist").(bool)

	d.SetId(key)

	keyInfo, err := cli.GetKey(key, client.GetKeyOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving key '%s': %s", key, err.Error()))
	}

	if !keyInfo.Found() {
		if mustExist {
			return errors.New(fmt.Sprintf("Error retrieving key '%s': it was not found", key))
		}

		d.Set("found", false)
		return nil
	}

	d.Set("value", keyInfo.Value)
	d.Set("version", keyInfo.Version)
	d.Set("create_revision", keyInfo.CreateRevision)
	d.Set("mod_revision", keyInfo.ModRevision)
	d.Set("lease", keyInfo.Lease)
	d.Set("found", true)

	return nil
}