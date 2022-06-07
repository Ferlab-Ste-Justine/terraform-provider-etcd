package provider

import (
    "encoding/json"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceRangeScopedState() *schema.Resource {
	return &schema.Resource{
		Description: "Resource to manage the lifecycle of a state scoped by a key range.",
		Create: resourceRangeScopedStateCreate,
		Read:   resourceRangeScopedStateRead,
		Delete: resourceRangeScopedStateDelete,
		Update: resourceRangeScopedStateUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Key specifying the beginning of the key range.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"range_end": {
				Description: "Key specifying the end of the key range (exclusive). To you set it to the value of the key scopes the range to a single key. If you would like the range to be anything prefixed by the key, you can use the etcd_prefix_range_end data helper.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"clear_on_creation": &schema.Schema{
				Description: "Whether to clear all pre-existing keys in the range when the resource is created.",
				Type:     schema.TypeBool,
				Optional: true,
				Default: true,
				ForceNew: false,
			},
			"clear_on_deletion": &schema.Schema{
				Description: "Whether to clear all existing keys in the range when the resource is deleted.",
				Type:     schema.TypeBool,
				Optional: true,
				Default: true,
				ForceNew: false,
			},
		},
	}
}

type RangeScopedState struct {
	Key             string
	RangeEnd        string
	ClearOnCreation bool
	ClearOnDeletion bool
}

func (state RangeScopedState) GetId() RangeScopedStateID {
	return RangeScopedStateID{state.Key, state.RangeEnd}
}

func rangeScopedStateSchemaToModel(d *schema.ResourceData) RangeScopedState {
	model := RangeScopedState{Key: "", RangeEnd: "", ClearOnCreation: true, ClearOnDeletion: false}

	model.Key = d.Get("key").(string)
	model.RangeEnd = d.Get("range_end").(string)
	model.ClearOnCreation = d.Get("clear_on_creation").(bool)
	model.ClearOnDeletion = d.Get("clear_on_deletion").(bool)

	return model
}

type RangeScopedStateID struct {
	Key             string
	RangeEnd        string
}

//Needed to absolutely ensure it is deterministic
func (id RangeScopedStateID) MarshalJSON() ([]byte, error) {
    mKey, _ := json.Marshal(id.Key)
    mRangeEnd, _ := json.Marshal(id.RangeEnd)
    return []byte(fmt.Sprintf("{\"Key\":%s,\"RangeEnd\":%s}", string(mKey), string(mRangeEnd))), nil
}

func (id RangeScopedStateID) Serialize() string {
    out, _ := json.Marshal(id)
    return string(out)
}

func DeserializeRangeScopedStateId(id string) (RangeScopedStateID, error) {
    var rangeStateId RangeScopedStateID
    err := json.Unmarshal([]byte(id), &rangeStateId)
    return rangeStateId, err
}

func resourceRangeScopedStateCreate(d *schema.ResourceData, meta interface{}) error {
	rangeState := rangeScopedStateSchemaToModel(d)
	conn := meta.(EtcdConnection)

	if rangeState.ClearOnCreation {
		err := conn.DeleteKeyRange(rangeState.Key, rangeState.RangeEnd)
		if err != nil {
			return errors.New(fmt.Sprintf("Error deleting key range ['%s', '%s'): %s", rangeState.Key, rangeState.RangeEnd, err.Error()))
		}
	}

	d.SetId(rangeState.GetId().Serialize())
	return resourceRangeScopedStateRead(d, meta)
}

func resourceRangeScopedStateRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceRangeScopedStateUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceRangeScopedStateRead(d, meta)
}

func resourceRangeScopedStateDelete(d *schema.ResourceData, meta interface{}) error {
	rangeState := rangeScopedStateSchemaToModel(d)
	conn := meta.(EtcdConnection)

	if rangeState.ClearOnDeletion {
		err := conn.DeleteKeyRange(rangeState.Key, rangeState.RangeEnd)
		if err != nil {
			return errors.New(fmt.Sprintf("Error deleting key range ['%s', '%s'): %s", rangeState.Key, rangeState.RangeEnd, err.Error()))
		}
	}

	return nil
}

