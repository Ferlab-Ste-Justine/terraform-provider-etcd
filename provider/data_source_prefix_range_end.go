package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func dataSourcePrefixRangeEnd() *schema.Resource {
	return &schema.Resource{
		Description: "Helper to retrieve a range end that, combined with the key argument, constitutes a prefix of key.",
		Read: dataSourcePrefixRangeEndRead,
		Schema: map[string]*schema.Schema{
			"key": &schema.Schema{
				Description: "Key to get a prefix of.",
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"range_end": &schema.Schema{
				Description: "Computed range end that, combined with the key, constitutes a prefix of the key.",
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourcePrefixRangeEndRead(d *schema.ResourceData, meta interface{}) error {
	key := d.Get("key").(string)
	
	d.SetId(key)
	d.Set("range_end", clientv3.GetPrefixRangeEnd(key))
	
	return nil
}