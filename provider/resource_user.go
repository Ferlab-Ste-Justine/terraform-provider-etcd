package provider

import (
	"errors"
	"fmt"

	"github.com/Ferlab-Ste-Justine/etcd-sdk/client"
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

func userSchemaToModel(d *schema.ResourceData) client.EtcdUser {
	model := client.EtcdUser{Username: "", Password: "", Roles: []string{}}

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

func resourceUserCreate(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	cli := meta.(client.EtcdClient)

	err := cli.UpsertUser(user)
	if err != nil {
		return err
	}

	d.SetId(user.Username)
	return resourceUserRead(d, meta)
}

func resourceUserRead(d *schema.ResourceData, meta interface{}) error {
	username := d.Id()
	cli := meta.(client.EtcdClient)

	resRoles, userExists, userRolesErr := cli.GetUserRoles(username)
	if userRolesErr != nil {
		return errors.New(fmt.Sprintf("Error retrieving existing user '%s' for reading: %s", username, userRolesErr.Error()))
	}

	if !userExists {
		d.SetId("")
		return nil
	}

	d.Set("username", username)
	d.Set("roles", resRoles)

	return nil
}

func resourceUserUpdate(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	cli := meta.(client.EtcdClient)

	err := cli.UpsertUser(user)
	if err != nil {
		return err
	}

	return resourceUserRead(d, meta)
}

func resourceUserDelete(d *schema.ResourceData, meta interface{}) error {
	user := userSchemaToModel(d)
	cli := meta.(client.EtcdClient)

	err := cli.DeleteUser(user.Username)
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting user '%s': %s", user.Username, err.Error()))
	}

	return nil
}
