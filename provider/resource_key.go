package provider

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceKey() *schema.Resource {
	return &schema.Resource{
		Description: "Key value for etcd.",
		Create: resourceKeyCreate,
		Read:   resourceKeyRead,
		Delete: resourceKeyDelete,
		Update: resourceKeyUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Key to set.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"value": {
				Description: "Value to store in the key.",
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
	conn := meta.(EtcdConnection)

	err := conn.PutKey(key.Key, key.Value)
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value for key '%s': %s", key.Key, err.Error()))
	}

	d.SetId(key.Key)
	return resourceKeyRead(d, meta)
}

func resourceKeyRead(d *schema.ResourceData, meta interface{}) error {
	key := d.Id()
	conn := meta.(EtcdConnection)

	val, err := conn.GetKey(key)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving key '%s' for reading: %s", key, err.Error()))
	}

	d.Set("key", key)
	d.Set("value", val)

	return nil
}

func resourceKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	key := keySchemaToModel(d)
	conn := meta.(EtcdConnection)

	err := conn.PutKey(key.Key, key.Value)
	if err != nil {
		return errors.New(fmt.Sprintf("Error setting value for key '%s': %s", key.Key, err.Error()))
	}

	return resourceKeyRead(d, meta)
}

func resourceKeyDelete(d *schema.ResourceData, meta interface{}) error {
	key := keySchemaToModel(d)
	conn := meta.(EtcdConnection)

	err := conn.DeleteKey(key.Key)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting key '%s': %s", key.Key, err.Error()))
	}

	return nil
}
