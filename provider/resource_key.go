package provider

import (
	"context"
	"errors"
	"fmt"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
    clientv3 "go.etcd.io/etcd/client/v3"
)

func resourceKey() *schema.Resource {
    return &schema.Resource{
        Create: resourceKeyCreate,
        Read: resourceKeyRead,
        Delete: resourceKeyDelete,
        Update: resourceKeyUpdate,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
            "key": {
                Type: schema.TypeString,
                Required: true,
                ForceNew: true,
                ValidateFunc: validation.StringIsNotEmpty,
            },
            "value": {
                Type: schema.TypeString,
                Required: true,
                ForceNew: false,
                ValidateFunc: validation.StringIsNotEmpty,
            },
        },
    }
}

type EtcdKey struct {
	Key string
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
	cli := meta.(*clientv3.Client)

	_, err := cli.Put(context.Background(), key.Key, key.Value)
    if err != nil {
		return errors.New(fmt.Sprintf("Error setting value for key '%s': %s", key.Key, err.Error()))
	}

	d.SetId(key.Key)
    return resourceKeyRead(d, meta)
}

func resourceKeyRead(d *schema.ResourceData, meta interface{}) error {
	key := d.Id()
	cli := meta.(*clientv3.Client)

	getRes, err := cli.Get(context.Background(), key)
    if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving key '%s' for reading: %s", key, err.Error()))
	}

	d.Set("key", key)
	d.Set("value", string(getRes.Kvs[0].Value))

	return nil
}

func resourceKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	key := keySchemaToModel(d)
	cli := meta.(*clientv3.Client)

	_, err := cli.Put(context.Background(), key.Key, key.Value)
    if err != nil {
		return errors.New(fmt.Sprintf("Error setting value for key '%s': %s", key.Key, err.Error()))
	}

    return resourceKeyRead(d, meta)
}

func resourceKeyDelete(d *schema.ResourceData, meta interface{}) error {
	key := keySchemaToModel(d)
	cli := meta.(*clientv3.Client)

	_, err := cli.Delete(context.Background(), key.Key)
    if err != nil {
		return errors.New(fmt.Sprintf("Error deleting key '%s': %s", key.Key, err.Error()))
	}

    return nil
}