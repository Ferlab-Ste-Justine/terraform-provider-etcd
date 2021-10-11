package provider

import (
	"context"
	"errors"
	"fmt"

    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
    clientv3 "go.etcd.io/etcd/client/v3"
)

func resourceUser() *schema.Resource {
    return &schema.Resource{
        Create: resourceUserCreate,
        Read: resourceUserRead,
        Delete: resourceUserDelete,
        Update: resourceUserUpdate,
        Importer: &schema.ResourceImporter{
            State: schema.ImportStatePassthrough,
        },
        Schema: map[string]*schema.Schema{
            "username": {
                Type: schema.TypeString,
                Required: true,
                ForceNew: true,
                ValidateFunc: validation.StringIsNotEmpty,
            },
            "password": {
                Type: schema.TypeString,
                Sensitive: true,
                Optional: true,
                ForceNew: false,
                ValidateFunc: validation.StringIsNotEmpty,
            },
            "roles": {
                Type: schema.TypeSet,
                Optional: true,
                ForceNew: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
        },
    }
}

type EtcdUser struct {
	Username string
	Password string
	Roles    []string
}

func userSchemaToModel(d *schema.ResourceData) EtcdUser {
    model := EtcdUser{Username: "", Password: "", Roles: []string{}}
	
	username, _ := d.GetOk("username")
	model.Username = username.(string)

	password, passwordExist := d.GetOk("password")
	if passwordExist {
		model.Password = password.(string)
	}

	roles, rolesExist := d.GetOk("roles")
	if rolesExist {
		for _, val := range (roles.(*schema.Set)).List() {
			role := val.(string)
			model.Roles = append(model.Roles, role)
		}
	}

	return model
}

func insertUser(cli *clientv3.Client, user EtcdUser) error {
	_, err := cli.UserAdd(context.Background(), user.Username, user.Password)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating new user '%s': %s", user.Username, err.Error()))
	}

	for _, role := range user.Roles {
		_, err := cli.UserGrantRole(context.Background(), user.Username, role)
		if err != nil {
			return errors.New(fmt.Sprintf("Error adding role '%s' to user '%s': %s", role, user.Username, err.Error()))
		}
	}

	return nil
}

func updateUser(cli *clientv3.Client, user EtcdUser) error {
	userRes, userErr := cli.UserGet(context.Background(), user.Username)
	if userErr != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing user '%s' for update: %s", user.Username, userErr.Error()))
	}

	_, passErr := cli.UserChangePassword(context.Background(), user.Username, user.Password)
	if passErr != nil {
		return errors.New(fmt.Sprintf("Error updating password of user '%s': %s", user.Username, passErr.Error()))
	}

	for _, role := range user.Roles {
		add := true
		for _, resRole := range userRes.Roles {
			if role == resRole {
				add = false
			}
		}

		if add {
			_, err := cli.UserGrantRole(context.Background(), user.Username, role)
			if err != nil {
				return errors.New(fmt.Sprintf("Error adding role '%s' to user '%s': %s", role, user.Username, err.Error()))
			}
		}
	}

	for _, resRole := range userRes.Roles {
		remove := true
		for _, role := range user.Roles {
			if resRole == role {
				remove = false
			}
		}

		if remove {
			_, err := cli.UserRevokeRole(context.Background(), user.Username, resRole)
			if err != nil {
				return errors.New(fmt.Sprintf("Error removing role '%s' from user '%s': %s", resRole, user.Username, err.Error()))
			}
		}
	}

	return nil
}

func upsertUser(cli *clientv3.Client, user EtcdUser) error {
	res, err := cli.UserList(context.Background())
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing users list: %s", err.Error()))
	}

	if isStringInSlice(user.Username, res.Users) {
		return updateUser(cli, user)
	}
	
	return insertUser(cli, user)
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	cli := meta.(*clientv3.Client)

	err := upsertUser(cli, user)
	if err != nil {
		return err
	}

	d.SetId(user.Username)
    return resourceUserRead(d, meta)
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	username := d.Id()
	cli := meta.(*clientv3.Client)
	
	userRes, userErr := cli.UserGet(context.Background(), username)
	if userErr != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing user '%s' for reading: %s", username, userErr.Error()))
	}

	d.Set("username", username)
	d.Set("roles", userRes.Roles)

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	cli := meta.(*clientv3.Client)

	err := upsertUser(cli, user)
	if err != nil {
		return err
	}

    return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	cli := meta.(*clientv3.Client)
	
	_, err := cli.UserDelete(context.Background(), user.Username)
    if err != nil {
		return errors.New(fmt.Sprintf("Error deleting user '%s': %s", user.Username, err.Error()))
	}

    return nil
}