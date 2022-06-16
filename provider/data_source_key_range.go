package provider

import (
	"fmt"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceKeyRange() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieves information about the keys contained in a given range.",
		Read: dataSourceKeyRangeRead,
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Key specifying the beginning of the key range.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"range_end": {
				Description: "Key specifying the end of the key range (exclusive). To you set it to the value of the key scopes the range to a single key. If you would like the range to be anything prefixed by the key, you can use the etcd_prefix_range_end data helper.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"results": &schema.Schema{
				Description: "List of keys that were read. Note that numerical values returned by etcd are in int64 format which might cause problems in int32 platforms.",
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "key": {
							Description: "key",
                            Type: schema.TypeString,
							Computed: true,
                        },
                        "value": {
							Description: "Value of the key",
                            Type: schema.TypeString,
							Computed: true,
                        },
						"version": {
							Description: "Current version of the key. Note that version is reset to 0 on deletion",
							Type: schema.TypeInt,
							Computed: true,
						},
						"create_revision": {
							Description: "Revision of the etcd keystore when the key was created",
							Type: schema.TypeInt,
							Computed: true,
						},
						"mod_revision": {
							Description: "Revision of the etcd keystore when the key was last modified",
							Type: schema.TypeInt,
							Computed: true,
						},
						"lease": {
							Description: "Id of the lease that the key is attached to. Will be 0 if the key is not attached to a lease.",
							Type: schema.TypeInt,
							Computed: true,
						},
                    },
				},
			},
		},
	}
}

func dataSourceKeyRangeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(EtcdConnection)
	key := d.Get("key").(string)
	rangeEnd := d.Get("range_end").(string)
	
	d.SetId(KeyRangeId{key, rangeEnd}.Serialize())

	keyInfos, err := conn.GetKeyRange(key, rangeEnd)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving key range (key='%s', range_end='%s'): %s", key, rangeEnd, err.Error()))
	}
	
	dataKeyInfos := make([]interface{}, len(keyInfos))

	idx := 0
	for _, keyInfo := range keyInfos {
		dataKeyInfo := make(map[string]interface{})

		dataKeyInfo["key"] = keyInfo.Key
		dataKeyInfo["value"] = keyInfo.Value
		dataKeyInfo["version"] = keyInfo.Version
		dataKeyInfo["create_revision"] = keyInfo.CreateRevision
		dataKeyInfo["mod_revision"] = keyInfo.ModRevision
		dataKeyInfo["lease"] = keyInfo.Lease

		dataKeyInfos[idx] = dataKeyInfo
		idx++
	}

	d.Set("results", dataKeyInfos)

	return nil
}