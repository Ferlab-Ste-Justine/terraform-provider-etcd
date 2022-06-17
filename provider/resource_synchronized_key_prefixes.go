package provider

import (
    "encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceSynchronizedKeyPrefixes() *schema.Resource {
	return &schema.Resource{
		Description: "Synchronizes a source key prefix with a destination key prefix either once when the resources is created or whenever the resource is applied. Note that the resource assumes that the destination is not being written to during synchronization.",
		Create: resourceSynchronizedKeyPrefixesCreate,
		Read:   resourceSynchronizedKeyPrefixesRead,
		Delete: resourceSynchronizedKeyPrefixesDelete,
		Update: resourceSynchronizedKeyPrefixesUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"source_prefix": {
				Description: "Source key prefix to synchronize from.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"destination_prefix": {
				Description: "Destination key prefix to synchronize to.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"recurrence": &schema.Schema{
				Description: "Defines when the resource should be recreated to trigger a resync. Can be set to once, onchange or always. Note that onchange looks for change during the plan phase only so consider setting it to always if another terraform resource in your script changes the source.",
				Type:     schema.TypeString,
				Optional: true,
				Default: "always",
				ForceNew: false,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					recurrence := val.(string)
					if recurrence != "once" && recurrence != "onchange" && recurrence != "always" {
						return []string{}, []error{errors.New("The recurrence field must be one of the following: once, onchange, always")}
					}
					return []string{}, []error{}
				},
			},
		},
	}
}

type SynchronizedKeyPrefixes struct {
	SourcePrefix           string
	DestinationPrefix      string
	Recurrence             string
}

func synchronizedKeyPrefixesSchemaToModel(d *schema.ResourceData) SynchronizedKeyPrefixes {
	model := SynchronizedKeyPrefixes{}

	model.SourcePrefix = d.Get("source_prefix").(string)
	model.DestinationPrefix  = d.Get("destination_prefix").(string)
	model.Recurrence = d.Get("recurrence").(string)

	return model
}

type SynchronizedKeyPrefixesId struct {
	SourcePrefix      string
	DestinationPrefix string
}

func (state SynchronizedKeyPrefixes) GetId() SynchronizedKeyPrefixesId {
	return SynchronizedKeyPrefixesId{state.SourcePrefix, state.DestinationPrefix}
}

//Needed to absolutely ensure it is deterministic
func (id SynchronizedKeyPrefixesId) MarshalJSON() ([]byte, error) {
    mSourcePrefix, _ := json.Marshal(id.SourcePrefix)
    mDesginationPrefix, _ := json.Marshal(id.DestinationPrefix)
    return []byte(fmt.Sprintf("{\"SourcePrefix\":%s,\"DestinationPrefix\":%s}", string(mSourcePrefix), string(mDesginationPrefix))), nil
}

func (id SynchronizedKeyPrefixesId) Serialize() string {
    out, _ := json.Marshal(id)
    return string(out)
}

func DeserializeSynchronizedKeyPrefixesId(id string) (SynchronizedKeyPrefixesId, error) {
    var synchronizedKeyPrefixesId SynchronizedKeyPrefixesId
    err := json.Unmarshal([]byte(id), &synchronizedKeyPrefixesId)
    return synchronizedKeyPrefixesId, err
}

func resourceSynchronizedKeyPrefixesCreate(d *schema.ResourceData, meta interface{}) error {
	synchronizedKeyPrefixes := synchronizedKeyPrefixesSchemaToModel(d)
	conn := meta.(EtcdConnection)

	diffs, err := conn.DiffPrefixes(synchronizedKeyPrefixes.SourcePrefix, synchronizedKeyPrefixes.DestinationPrefix)
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting differential of prefix %s from prefix %s: %s", synchronizedKeyPrefixes.DestinationPrefix, synchronizedKeyPrefixes.SourcePrefix, err.Error()))
	}

	if !diffs.IsEmpty() {
		err := conn.ApplyDiffToPrefix(synchronizedKeyPrefixes.DestinationPrefix, diffs)
		if err != nil {
			return errors.New(fmt.Sprintf("Error applying differential to prefix %s: %s", synchronizedKeyPrefixes.DestinationPrefix, err.Error()))
		}
	}

	d.SetId(synchronizedKeyPrefixes.GetId().Serialize())
	return nil
}

func resourceSynchronizedKeyPrefixesRead(d *schema.ResourceData, meta interface{}) error {
	synchronizedKeyPrefixes := synchronizedKeyPrefixesSchemaToModel(d)
	conn := meta.(EtcdConnection)

	if synchronizedKeyPrefixes.Recurrence == "once" {
		return nil
	}

	if synchronizedKeyPrefixes.Recurrence == "always" {
		d.SetId("")
		return nil
	}

	diffs, err := conn.DiffPrefixes(synchronizedKeyPrefixes.SourcePrefix, synchronizedKeyPrefixes.DestinationPrefix)
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting differential of prefix %s from prefix %s: %s", synchronizedKeyPrefixes.DestinationPrefix, synchronizedKeyPrefixes.SourcePrefix, err.Error()))
	}

	if !diffs.IsEmpty() {
		d.SetId("")
	}

	return nil
}

func resourceSynchronizedKeyPrefixesUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceRangeScopedStateRead(d, meta)
}

func resourceSynchronizedKeyPrefixesDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}