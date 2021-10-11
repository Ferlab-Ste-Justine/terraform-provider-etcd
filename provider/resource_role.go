package provider

import (
	"context"
	"errors"
	"fmt"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
    clientv3 "go.etcd.io/etcd/client/v3"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
        Create: resourceRoleCreate,
        Read: resourceRoleRead,
        Delete: resourceRoleDelete,
        Update: resourceRoleUpdate,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
            "name": {
                Type: schema.TypeString,
                Required: true,
                ForceNew: true,
                ValidateFunc: validation.StringIsNotEmpty,
            },
            "permissions": {
                Type: schema.TypeSet,
                Optional: true,
                ForceNew: false,
                Elem: &schema.Resource{
                    Schema: map[string]*schema.Schema{
                        "permission": {
                            Type: schema.TypeString,
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
                            Type: schema.TypeString,
                            Required: true,
                            ForceNew: false,
                            ValidateFunc: validation.StringIsNotEmpty,
                        },
                        "range_end": {
                            Type: schema.TypeString,
                            Required: true,
                            ForceNew: false,
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
	Key string
	RangeEnd string
}

type EtcdRole struct {
	Name string
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

func permissionPermToEnum(permission EtcdRolePermission) clientv3.PermissionType {
	if permission.Permission == "readwrite" {
		return clientv3.PermissionType(clientv3.PermReadWrite)
	} else if permission.Permission == "read" {
		return clientv3.PermissionType(clientv3.PermRead)
	} else {
		return clientv3.PermissionType(clientv3.PermWrite)
	}
}

func permissionEnumToPerm(perm clientv3.PermissionType) string {
	if perm == clientv3.PermissionType(clientv3.PermReadWrite) {
		return "readwrite"
	} else if perm == clientv3.PermissionType(clientv3.PermRead) {
		return "read"
	} else {
		return "write"
	}
}

func insertRole(cli *clientv3.Client, role EtcdRole) error {
	_, err := cli.RoleAdd(context.Background(), role.Name)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating new role '%s': %s", role.Name, err.Error()))
	}

	for _, permission := range role.Permissions {
		_, err := cli.RoleGrantPermission(context.Background(), role.Name, permission.Key, permission.RangeEnd, permissionPermToEnum(permission))
		if err != nil {
			return errors.New(fmt.Sprintf("Error adding role permission (key='%s', range_end='%s', permission='%s') for role '%s': %s", permission.Key, permission.RangeEnd, permission.Permission, role.Name, err.Error()))
		}
	}

	return nil
}

func updateRole(cli *clientv3.Client, role EtcdRole) error {
	res, err := cli.RoleGet(context.Background(), role.Name)
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing role '%s' for update: %s", role.Name, err.Error()))
	}

	etcdPermissions := res.Perm

	for _, etcdPermission := range etcdPermissions {
		remove := true
		for _, permission := range role.Permissions {
			pType := permissionPermToEnum(permission)
			if clientv3.PermissionType(etcdPermission.PermType) == pType && string(etcdPermission.Key) == permission.Key && string(etcdPermission.RangeEnd) == permission.RangeEnd {
				remove = false
			}
		}
		if remove {
			_, err := cli.RoleRevokePermission(context.Background(), role.Name, string(etcdPermission.Key), string(etcdPermission.RangeEnd))
			if err != nil {
				permType := permissionEnumToPerm(clientv3.PermissionType(etcdPermission.PermType))
				return errors.New(fmt.Sprintf("Error removing role permission (key='%s', range_end='%s', permission='%s') for role '%s': %s", string(etcdPermission.Key), string(etcdPermission.RangeEnd), permType, role.Name, err.Error()))
			}
		}
	}

	for _, permission := range role.Permissions {
		add := true
		pType := permissionPermToEnum(permission)
		for _, etcdPermission := range etcdPermissions {
			if clientv3.PermissionType(etcdPermission.PermType) == pType && string(etcdPermission.Key) == permission.Key && string(etcdPermission.RangeEnd) == permission.RangeEnd {
				add = false
			}
		}
		if add {
			_, err := cli.RoleGrantPermission(context.Background(), role.Name, permission.Key, permission.RangeEnd, pType)
			if err != nil {
				return errors.New(fmt.Sprintf("Error adding role permission (key='%s', range_end='%s', permission='%s') for role '%s': %s", permission.Key, permission.RangeEnd, permission.Permission, role.Name, err.Error()))
			}
		}
	}

	return nil
}

func upsertRole(cli *clientv3.Client, role EtcdRole) error {
	res, err := cli.RoleList(context.Background())
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing roles list: %s", err.Error()))
	}

	if isStringInSlice(role.Name, res.Roles) {
		return updateRole(cli, role)
	}
	
	return insertRole(cli, role)
}

func resourceRoleCreate(d *schema.ResourceData, meta interface{}) error {
	role := roleSchemaToModel(d)
	cli := meta.(*clientv3.Client)
	
	err := upsertRole(cli, role)
	if err != nil {
		return err
	}
	
	d.SetId(role.Name)
    return resourceRoleRead(d, meta)
}

func resourceRoleRead(d *schema.ResourceData, meta interface{}) error {
	roleName := d.Id()
	cli := meta.(*clientv3.Client)
	
	res, err := cli.RoleGet(context.Background(), roleName)
    if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing role '%s' for reading: %s", roleName, err.Error()))
	}

	d.Set("name", roleName)
	permissions := make([]map[string]interface{}, 0)
	for _, v := range res.Perm {
		permissions = append(permissions, map[string]interface{}{
			"permission": permissionEnumToPerm(clientv3.PermissionType(v.PermType)),
			"key": v.Key,
			"range_end": v.RangeEnd,
		})
	}
	d.Set("permissions", permissions)
    
	return nil
}

func resourceRoleUpdate(d *schema.ResourceData, meta interface{}) error {
	role := roleSchemaToModel(d)
	cli := meta.(*clientv3.Client)
	upsertRole(cli, role)
    return resourceRoleRead(d, meta)
}

func resourceRoleDelete(d *schema.ResourceData, meta interface{}) error {
	role := roleSchemaToModel(d)
	cli := meta.(*clientv3.Client)
	
	_, err := cli.RoleDelete(context.Background(), role.Name)
    if err != nil {
		return errors.New(fmt.Sprintf("Error deleting role '%s': %s", role.Name, err.Error()))
	}

	return nil
}