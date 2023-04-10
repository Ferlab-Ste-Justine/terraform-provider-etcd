package provider

import (
	"errors"
	"fmt"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKey() *schema.Resource {
	return &schema.Resource{
		Description: "Key value for etcd.",
		Create:      resourceKeyCreate,
		Read:        resourceKeyRead,
		Delete:      resourceKeyDelete,
		Update:      resourceKeyUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description:  "Key to set.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"value": {
				Description:  "Value to store in the key.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

type EtcdKey struct {
	Key   string
	Value string
}

func keySchemaToModel(d *schema.ResourceData) EtcdKey {
	model := EtcdKey{Key: "", Value: ""}

	key, _ := d.GetOk("key")
	model.Key = key.(string)

	value, _ := d.GetOk("value")
	model.Value = value.(string)

	return model
}

func resourceKeyCreate(d *schema.ResourceData, meta interface{}) error {
	key := keySchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	err := cli.PutKey(key.Key, key.Value)
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value for key '%s': %s", key.Key, err.Error()))
	}

	d.SetId(key.Key)
	return resourceKeyRead(d, meta)
}

func resourceKeyRead(d *schema.ResourceData, meta interface{}) error {
	key := d.Id()
	cli := meta.(*client.EtcdClient)

	val, exists, err := cli.GetKey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving key '%s' for reading: %s", key, err.Error()))
	}

	if !exists {
		d.SetId("")
		return nil
	}

	d.Set("key", key)
	d.Set("value", val)

	return nil
}

func resourceKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	key := keySchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	err := cli.PutKey(key.Key, key.Value)
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value for key '%s': %s", key.Key, err.Error()))
	}

	return resourceKeyRead(d, meta)
}

func resourceKeyDelete(d *schema.ResourceData, meta interface{}) error {
	key := keySchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	err := cli.DeleteKey(key.Key)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting key '%s': %s", key.Key, err.Error()))
	}

	return nil
}
