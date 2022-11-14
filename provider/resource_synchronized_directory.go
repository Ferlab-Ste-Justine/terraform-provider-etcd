package provider

import (
    "encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
	"github.com/Ferlab-Ste-Justine/etcd-sdk/keymodels"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func DiffPrefixWithInput(cli *client.EtcdClient, prefix string, inputKeys map[string]keymodels.KeyInfo, inputKeysPrefix string, inputIsSource bool) (keymodels.KeysDiff, error) {
	prefixKeys, _, err := cli.GetPrefix(prefix)
	if err != nil {
		return keymodels.KeysDiff{}, err
	}

	if inputIsSource {
		return keymodels.GetKeysDiff(inputKeys, inputKeysPrefix, prefixKeys, prefix), nil
	}

	return keymodels.GetKeysDiff(prefixKeys, prefix, inputKeys, inputKeysPrefix), nil
}

func resourceSynchronizedDirectory() *schema.Resource {
	return &schema.Resource{
		Description: "Synchronizes the content of an key prefix and directory. Note that etcd is has a default max object size of 1.5MiB and is most suitable for keys that are bounded to a small size like configurations. Use another solution for larger files. Also, currently, only file systems following the unix convention are supported.",
		Create: resourceSynchronizedDirectoryCreate,
		Read:   resourceSynchronizedDirectoryRead,
		Delete: resourceSynchronizedDirectoryDelete,
		Update: resourceSynchronizedDirectoryUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key_prefix": {
				Description: "Key prefix to synchronize with the directory.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"directory": {
				Description: "Directory to synchronize with the key prefix.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"source": {
				Description: "Authoritative source of data during the sync (data will move from the source to the destination). Can be one of: directory, key-prefix",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc:  func(val interface{}, key string) (warns []string, errs []error) {
					source := val.(string)
					if source != "directory" && source != "key-prefix" {
						return []string{}, []error{errors.New("The source field must be one of the following: directory, key-prefix")}
					}
					return []string{}, []error{}
				},
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
			"files_permission": &schema.Schema{
				Description: "Permission of generated files in the case where the directory is the destination.",
				Type:     schema.TypeString,
				Optional: true,
				Default: "0700",
				ForceNew: false,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					permission := val.(string)
					iPermission, err := strconv.ParseInt(permission, 8, 32)
					if err != nil || iPermission < 0 || iPermission > 511 {
						return []string{}, []error{errors.New("The files_permission field must constitute a valid unix value for file permissions")}
					}
					return []string{}, []error{}
				},
			},
			"directory_permission": &schema.Schema{
				Description: "Permission of generated directories if the directory is the destination and missing.",
				Type:     schema.TypeString,
				Optional: true,
				Default: "0700",
				ForceNew: false,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					permission := val.(string)
					iPermission, err := strconv.ParseInt(permission, 8, 32)
					if err != nil || iPermission < 0 || iPermission > 511 {
						return []string{}, []error{errors.New("The directory_permission field must constitute a valid unix value for file permissions")}
					}
					return []string{}, []error{}
				},
			},
		},
	}
}

type SynchronizedDirectory struct {
	KeyPrefix           string
	Directory           string
	Source              string
	Recurrence          string
	FilesPermission     int32
	DirectoryPermission int32
}

func synchronizedDirectorySchemaToModel(d *schema.ResourceData) SynchronizedDirectory {
	model := SynchronizedDirectory{}

	model.KeyPrefix = d.Get("key_prefix").(string)
	model.Source = d.Get("source").(string)
	model.Recurrence = d.Get("recurrence").(string)

	fPermission, _ := strconv.ParseInt(d.Get("files_permission").(string), 8, 32)
	model.FilesPermission = int32(fPermission)

	dPermission, _ := strconv.ParseInt(d.Get("directory_permission").(string), 8, 32)
	model.DirectoryPermission = int32(dPermission)

	directory, _ := filepath.Abs(d.Get("directory").(string))
	if directory[len(directory)-1:] != "/" {
		directory = directory + "/"
	}
	model.Directory = directory

	return model
}

type SynchronizedDirectoryId struct {
	KeyPrefix string
	Directory string
}

func (state SynchronizedDirectory) GetId() SynchronizedDirectoryId {
	return SynchronizedDirectoryId{state.KeyPrefix, state.Directory}
}

//Needed to absolutely ensure it is deterministic
func (id SynchronizedDirectoryId) MarshalJSON() ([]byte, error) {
    mKeyPrefix, _ := json.Marshal(id.KeyPrefix)
    mDirectory, _ := json.Marshal(id.Directory)
    return []byte(fmt.Sprintf("{\"KeyPrefix\":%s,\"Directory\":%s}", string(mKeyPrefix), string(mDirectory))), nil
}

func (id SynchronizedDirectoryId) Serialize() string {
    out, _ := json.Marshal(id)
    return string(out)
}

func DeserializeSynchronizedDirectoryId(id string) (SynchronizedDirectoryId, error) {
    var synchronizedDirectoryId SynchronizedDirectoryId
    err := json.Unmarshal([]byte(id), &synchronizedDirectoryId)
    return synchronizedDirectoryId, err
}

func resourceSynchronizedDirectoryCreate(d *schema.ResourceData, meta interface{}) error {
	synchronizedDirectory := synchronizedDirectorySchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	if synchronizedDirectory.Source == "key-prefix" {
		EnsureDirectoryExists(synchronizedDirectory.Directory, synchronizedDirectory.DirectoryPermission)
	}

	dirKeys, dirErr := GetDirectoryContent(synchronizedDirectory.Directory)
	if dirErr != nil {
		return errors.New(fmt.Sprintf("Error getting differential of prefix %s and directory %s: %s", synchronizedDirectory.KeyPrefix, synchronizedDirectory.Directory, dirErr.Error()))
	}

	inputIsSource := synchronizedDirectory.Source == "directory"
	diffs, err := DiffPrefixWithInput(cli, synchronizedDirectory.KeyPrefix, dirKeys, synchronizedDirectory.Directory, inputIsSource)
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting differential of prefix %s and directory %s: %s", synchronizedDirectory.KeyPrefix, synchronizedDirectory.Directory, err.Error()))
	}

	if !diffs.IsEmpty() {
		if synchronizedDirectory.Source == "directory" {
			err := cli.ApplyDiffToPrefix(synchronizedDirectory.KeyPrefix, diffs)
			if err != nil {
				return errors.New(fmt.Sprintf("Error synchronizing changes to key prefix %s: %s", synchronizedDirectory.KeyPrefix, err.Error()))
			}
		} else {
			err := ApplyDiffToDirectory(synchronizedDirectory.Directory, diffs, synchronizedDirectory.FilesPermission, synchronizedDirectory.DirectoryPermission)
			if err != nil {
				return errors.New(fmt.Sprintf("Error synchronizing changes to directory %s: %s", synchronizedDirectory.Directory, err.Error()))
			}
		}
	}

	d.SetId(synchronizedDirectory.GetId().Serialize())
	return nil
}

func resourceSynchronizedDirectoryRead(d *schema.ResourceData, meta interface{}) error {
	synchronizedDirectory := synchronizedDirectorySchemaToModel(d)
	cli := meta.(*client.EtcdClient)

	if synchronizedDirectory.Source == "key-prefix" {
		EnsureDirectoryExists(synchronizedDirectory.Directory, synchronizedDirectory.DirectoryPermission)
	}

	if synchronizedDirectory.Recurrence == "once" {
		return nil
	}

	if synchronizedDirectory.Recurrence == "always" {
		d.SetId("")
		return nil
	}

	dirKeys, dirErr := GetDirectoryContent(synchronizedDirectory.Directory)
	if dirErr != nil {
		return errors.New(fmt.Sprintf("Error getting differential of prefix %s and directory %s: %s", synchronizedDirectory.KeyPrefix, synchronizedDirectory.Directory, dirErr.Error()))
	}

	//input is always defined as source for the direction, but it doesn't matter, we just want to see if the diff is empty
	diffs, err := DiffPrefixWithInput(cli, synchronizedDirectory.KeyPrefix, dirKeys, synchronizedDirectory.Directory, true)
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting differential of prefix %s and directory %s: %s", synchronizedDirectory.KeyPrefix, synchronizedDirectory.Directory, err.Error()))
	}

	if !diffs.IsEmpty() {
		d.SetId("")
	}

	return nil
}

func resourceSynchronizedDirectoryUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceSynchronizedDirectoryRead(d, meta)
}

func resourceSynchronizedDirectoryDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}