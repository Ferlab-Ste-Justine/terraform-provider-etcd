package provider

import (
	"errors"
	"fmt"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKeyPrefix() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to manage all the keys contained within a specified prefix.",
		Create:      resourceKeyPrefixCreate,
		Read:        resourceKeyPrefixRead,
		Delete:      resourceKeyPrefixDelete,
		Update:      resourceKeyPrefixUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"prefix": {
				Description:  "Prefix of keys to set.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"keys": {
				Description: "Keys to define in the prefix.",
				Type:     schema.TypeSet,
                Required: true,
                ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
			"clear_on_deletion": &schema.Schema{
				Description: "Whether to clear all existing keys with the prefix when the resource is deleted.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    false,
			},
		},
	}
}

type EtcdKeyPrefix struct {
	Prefix          string
	Keys            map[string]string
	ClearOnDeletion bool
}

func keyPrefixSchemaToModel(d *schema.ResourceData) EtcdKeyPrefix {
	model := EtcdKeyPrefix{Keys: make(map[string]string)}

	prefix, _ := d.GetOk("prefix")
	model.Prefix = prefix.(string)

	keys, keysExists := d.GetOk("retry")
	if keysExists {
		for _, elem := range (keys.(*schema.Set)).List() {
			elemMap := elem.(map[string]interface{})
			key, _ := elemMap["key"]
			val, _ := elemMap["value"]
			model.Keys[key.(string)] = val.(string)
		}
	}

	clearOnDeletion, _ := d.GetOk("clear_on_deletion")
	model.ClearOnDeletion = clearOnDeletion.(bool)

	return model
}

func resourceKeyPrefixCreate(d *schema.ResourceData, meta interface{}) error {
	keyPrefix := keyPrefixSchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	prefixKeys, prefixErr := cli.GetPrefix(keyPrefix.Prefix)
	if prefixErr != nil {
		return errors.New(fmt.Sprintf("Error getting keys under prefix '%s': %s", keyPrefix.Prefix, prefixErr.Error()))
	}

	diff := client.GetKeyDiff(keyPrefix.Keys, prefixKeys.Keys.ToValueMap(keyPrefix.Prefix))

	applyErr := cli.ApplyDiffToPrefix(keyPrefix.Prefix, diff)
	if applyErr != nil {
		return errors.New(fmt.Sprintf("Error applying key changes under prefix '%s': %s", keyPrefix.Prefix, applyErr.Error()))
	}

	d.SetId(keyPrefix.Prefix)
	return resourceKeyPrefixRead(d, meta)
}

func resourceKeyPrefixRead(d *schema.ResourceData, meta interface{}) error {
	keyPrefix := keyPrefixSchemaToModel(d)
	cli := meta.(*client.EtcdClient)
	
	prefixKeys, prefixErr := cli.GetPrefix(keyPrefix.Prefix)
	if prefixErr != nil {
		return errors.New(fmt.Sprintf("Error getting keys under prefix '%s': %s", keyPrefix.Prefix, prefixErr.Error()))
	}

    keys := make([]map[string]interface{}, 0)
    for _, v := range prefixKeys.Keys {
        keys = append(keys, map[string]interface{}{
            "key": v.Key,
            "value": v.Value,
        })
    }
    d.Set("keys", keys)

	return nil
}

func resourceKeyPrefixUpdate(d *schema.ResourceData, meta interface{}) error {
	keyPrefix := keyPrefixSchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	prefixKeys, prefixErr := cli.GetPrefix(keyPrefix.Prefix)
	if prefixErr != nil {
		return errors.New(fmt.Sprintf("Error getting keys under prefix '%s': %s", keyPrefix.Prefix, prefixErr.Error()))
	}

	diff := client.GetKeyDiff(keyPrefix.Keys, prefixKeys.Keys.ToValueMap(keyPrefix.Prefix))

	applyErr := cli.ApplyDiffToPrefix(keyPrefix.Prefix, diff)
	if applyErr != nil {
		return errors.New(fmt.Sprintf("Error applying key changes under prefix '%s': %s", keyPrefix.Prefix, applyErr.Error()))
	}

	return resourceKeyPrefixRead(d, meta)
}

func resourceKeyPrefixDelete(d *schema.ResourceData, meta interface{}) error {
	keyPrefix := keyPrefixSchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	if !keyPrefix.ClearOnDeletion {
		return nil
	}

	deleteErr := cli.DeletePrefix(keyPrefix.Prefix)
	if deleteErr != nil {
		return errors.New(fmt.Sprintf("Error deleting keys under prefix '%s': %s", keyPrefix.Prefix, deleteErr.Error()))
	}

	return nil
}
