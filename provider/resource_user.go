package provider

import (
	"errors"
	"fmt"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceUser() *schema.Resource {
	return &schema.Resource{
		Description: "User that can access etcd.",
		Create: resourceUserCreate,
		Read:   resourceUserRead,
		Delete: resourceUserDelete,
		Update: resourceUserUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Description: "Name of the user. Changing this will delete the user and create a new one.",
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"password": {
				Description: "Password of the user. Can be omitted for a user that you wish to authenticate strictly with tls certificate authentication.",
				Type:         schema.TypeString,
				Sensitive:    true,
				Optional:     true,
				ForceNew:     false,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"roles": {
				Description: "Roles of the user, to define his access.",
				Type:     schema.TypeSet,
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

func insertUser(conn EtcdConnection, user EtcdUser) error {
	err := conn.AddUser(user.Username, user.Password)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating new user '%s': %s", user.Username, err.Error()))
	}

	for _, role := range user.Roles {
		err := conn.GrantUserRole(user.Username, role)
		if err != nil {
			return errors.New(fmt.Sprintf("Error adding role '%s' to user '%s': %s", role, user.Username, err.Error()))
		}
	}

	return nil
}

func updateUser(conn EtcdConnection, user EtcdUser) error {
	resRoles, userRolesErr := conn.GetUserRoles(user.Username)
	if userRolesErr != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing user '%s' for update: %s", user.Username, userRolesErr.Error()))
	}

	passErr := conn.ChangeUserPassword(user.Username, user.Password)
	if passErr != nil {
		return errors.New(fmt.Sprintf("Error updating password of user '%s': %s", user.Username, passErr.Error()))
	}

	for _, role := range user.Roles {
		add := true
		for _, resRole := range resRoles {
			if role == resRole {
				add = false
			}
		}

		if add {
			err := conn.GrantUserRole(user.Username, role)
			if err != nil {
				return errors.New(fmt.Sprintf("Error adding role '%s' to user '%s': %s", role, user.Username, err.Error()))
			}
		}
	}

	for _, resRole := range resRoles {
		remove := true
		for _, role := range user.Roles {
			if resRole == role {
				remove = false
			}
		}

		if remove {
			err := conn.RevokeUserRole(user.Username, resRole)
			if err != nil {
				return errors.New(fmt.Sprintf("Error removing role '%s' from user '%s': %s", resRole, user.Username, err.Error()))
			}
		}
	}

	return nil
}

func upsertUser(conn EtcdConnection, user EtcdUser) error {
	users, err := conn.ListUsers()
	if err != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing users list: %s", err.Error()))
	}

	if isStringInSlice(user.Username, users) {
		return updateUser(conn, user)
	}

	return insertUser(conn, user)
}

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	conn := meta.(EtcdConnection)

	err := upsertUser(conn, user)
	if err != nil {
		return err
	}

	d.SetId(user.Username)
	return resourceUserRead(d, meta)
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	username := d.Id()
	conn := meta.(EtcdConnection)

	resRoles, userRolesErr := conn.GetUserRoles(username)
	if userRolesErr != nil {
		etcdErr, ok := userRolesErr.(rpctypes.EtcdError)
		if !ok {
			return errors.New(fmt.Sprintf("Error retrieving existing user '%s' for reading: %s", username, userRolesErr.Error()))
		}
		
		if etcdErr.Error() == rpctypes.ErrorDesc(rpctypes.ErrGRPCUserNotFound) {
			d.SetId("")
			return nil
		}

		return errors.New(fmt.Sprintf("Error retrieving existing user '%s' for reading: %s", username, userRolesErr.Error()))
	}

	d.Set("username", username)
	d.Set("roles", resRoles)

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	conn := meta.(EtcdConnection)

	err := upsertUser(conn, user)
	if err != nil {
		return err
	}

	return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	conn := meta.(EtcdConnection)

	err := conn.DeleteUser(user.Username)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting user '%s': %s", user.Username, err.Error()))
	}

	return nil
}
