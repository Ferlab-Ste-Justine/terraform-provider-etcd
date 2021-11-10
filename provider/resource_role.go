package provider

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: "User role for etcd to define access control.",
		Create: resourceRoleCreate,
		Read:   resourceRoleRead,
		Delete: resourceRoleDelete,
		Update: resourceRoleUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the role. Changing this will delete the role and create a new one.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"permissions": {
				Description: "Permissions to grant to the role on various etcd key ranges.",
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"permission": {
							Description: "Permissions to grant to the role on the given key range. Can be: read, write or readwrite",
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: false,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)
								if v != "read" && v != "write" && v != "readwrite" {
									return []string{}, []error{errors.New("Permission value for role can only be one of the followings: read, write, readwrite")}
								}

								return []string{}, []error{}
							},
						},
						"key": {
							Description: "Key specifying the beginning of the key range.",
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     false,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"range_end": {
							Description: "Key specifying the end of the key range (exclusive). To you set it to the value of the key to grant permission on a single key. If you would like the range to be anything prefixed by the key, you can use the etcd_prefix_range_end data helper.",
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     false,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
		},
	}
}

type EtcdRolePermission struct {
	Permission string
	Key        string
	RangeEnd   string
}

type EtcdRole struct {
	Name        string
	Permissions []EtcdRolePermission
}

func roleSchemaToModel(d *schema.ResourceData) EtcdRole {
	model := EtcdRole{Name: "", Permissions: []EtcdRolePermission{}}

	name, _ := d.GetOk("name")
	model.Name = name.(string)

	permissions, permissionsExist := d.GetOk("permissions")
	if permissionsExist {
		for _, val := range (permissions.(*schema.Set)).List() {
			permission := val.(map[string]interface{})
			model.Permissions = append(model.Permissions, EtcdRolePermission{Permission: permission["permission"].(string), Key: permission["key"].(string), RangeEnd: permission["range_end"].(string)})
		}
	}

	return model
}

func insertRole(conn EtcdConnection, role EtcdRole) error {
	err := conn.AddRole(role.Name)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating new role '%s': %s", role.Name, err.Error()))
	}

	for _, permission := range role.Permissions {
		err := conn.GrantRolePermission(role.Name, permission.Key, permission.RangeEnd, permission.Permission)
		if err != nil {
			return errors.New(fmt.Sprintf("Error adding role permission (key='%s', range_end='%s', permission='%s') for role '%s': %s", permission.Key, permission.RangeEnd, permission.Permission, role.Name, err.Error()))
		}
	}

	return nil
}

func updateRole(conn EtcdConnection, role EtcdRole) error {

	resPermissions, err := conn.GetRolePermissions(role.Name)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing role '%s' for update: %s", role.Name, err.Error()))
	}

	for _, resPermission := range resPermissions {
		remove := true
		for _, permission := range role.Permissions {
			if resPermission.Permission == permission.Permission && resPermission.Key == permission.Key && resPermission.RangeEnd == permission.RangeEnd {
				remove = false
			}
		}
		if remove {
			err := conn.RevokeRolePermission(role.Name, resPermission.Key, resPermission.RangeEnd)
			if err != nil {
				return errors.New(fmt.Sprintf("Error removing role permission (key='%s', range_end='%s', permission='%s') for role '%s': %s", resPermission.Key, resPermission.RangeEnd, resPermission.Permission, role.Name, err.Error()))
			}
		}
	}

	for _, permission := range role.Permissions {
		add := true
		for _, resPermission := range resPermissions {
			if resPermission.Permission == permission.Permission && resPermission.Key == permission.Key && resPermission.RangeEnd == permission.RangeEnd {
				add = false
			}
		}
		if add {
			err := conn.GrantRolePermission(role.Name, permission.Key, permission.RangeEnd, permission.Permission)
			if err != nil {
				return errors.New(fmt.Sprintf("Error adding role permission (key='%s', range_end='%s', permission='%s') for role '%s': %s", permission.Key, permission.RangeEnd, permission.Permission, role.Name, err.Error()))
			}
		}
	}

	return nil
}

func upsertRole(conn EtcdConnection, role EtcdRole) error {
	roles, err := conn.ListRoles()
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing roles list: %s", err.Error()))
	}

	if isStringInSlice(role.Name, roles) {
		return updateRole(conn, role)
	}

	return insertRole(conn, role)
}

func resourceRoleCreate(d *schema.ResourceData, meta interface{}) error {
	role := roleSchemaToModel(d)
	conn := meta.(EtcdConnection)

	err := upsertRole(conn, role)
	if err != nil {
		return err
	}

	d.SetId(role.Name)
	return resourceRoleRead(d, meta)
}

func resourceRoleRead(d *schema.ResourceData, meta interface{}) error {
	roleName := d.Id()
	conn := meta.(EtcdConnection)

	resPermissions, err := conn.GetRolePermissions(roleName)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing role '%s' for reading: %s", roleName, err.Error()))
	}

	d.Set("name", roleName)
	permissions := make([]map[string]interface{}, 0)
	for _, resPermission := range resPermissions {
		permissions = append(permissions, map[string]interface{}{
			"permission": resPermission.Permission,
			"key":        resPermission.Key,
			"range_end":  resPermission.RangeEnd,
		})
	}
	d.Set("permissions", permissions)

	return nil
}

func resourceRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	role := roleSchemaToModel(d)
	conn := meta.(EtcdConnection)
	upsertRole(conn, role)
	return resourceRoleRead(d, meta)
}

func resourceRoleDelete(d *schema.ResourceData, meta interface{}) error {
	role := roleSchemaToModel(d)
	conn := meta.(EtcdConnection)

	err := conn.DeleteRole(role.Name)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting role '%s': %s", role.Name, err.Error()))
	}

	return nil
}
