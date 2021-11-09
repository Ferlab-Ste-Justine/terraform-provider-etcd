package provider

import (
	"context"
	"time"

    clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdConnection struct {
	Client *clientv3.Client
	Timeout int
	Retries int
}

/* Key Values */

func (conn *EtcdConnection) PutKey(key string, val string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.Put(ctx, key, val)
	return err
}

func (conn *EtcdConnection) GetKey(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	getRes, err := conn.Client.Get(ctx, key)

    if err != nil {
		return "", err
	}

	return string(getRes.Kvs[0].Value), nil
}

func (conn *EtcdConnection) DeleteKey(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.Delete(ctx, key)
	return err
}

/* Users */

func (conn *EtcdConnection) ListUsers() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	res, err := conn.Client.UserList(ctx)
	if err != nil {
		return []string{}, err
	}

	return res.Users, nil
}

func (conn *EtcdConnection) AddUser(username string, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.UserAdd(ctx, username, password)
	return err
}

func (conn *EtcdConnection) GetUserRoles(username string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	res, err := conn.Client.UserGet(ctx, username)
	if err != nil {
		return []string{}, err
	}

	return res.Roles, nil
}

func (conn *EtcdConnection) ChangeUserPassword(username string, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.UserChangePassword(ctx, username, password)
	return err
}


func (conn *EtcdConnection) GrantUserRole(username string, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.UserGrantRole(ctx, username, role)
	return err
}

func (conn *EtcdConnection) RevokeUserRole(username string, role string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.UserRevokeRole(ctx, username, role)
	return err
}

func (conn *EtcdConnection) DeleteUser(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.UserDelete(ctx, username)
	return err
}

/* Roles */

func permissionToEnum(permission string) clientv3.PermissionType {
	if permission == "readwrite" {
		return clientv3.PermissionType(clientv3.PermReadWrite)
	} else if permission == "read" {
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

func (conn *EtcdConnection) ListRoles() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	res, err := conn.Client.RoleList(ctx)
	if err != nil {
		return []string{}, err
	}

	return res.Roles, nil
}

func (conn *EtcdConnection) AddRole(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.RoleAdd(ctx, name)
	return err
}

func (conn *EtcdConnection) GrantRolePermission(name string, key string, rangeEnd string, permission string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.RoleGrantPermission(ctx, name, key, rangeEnd, permissionToEnum(permission))
	return err
}

func (conn *EtcdConnection) RevokeRolePermission(name string, key string, rangeEnd string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.RoleRevokePermission(ctx, name, key, rangeEnd)
	return err
}

func (conn *EtcdConnection) GetRolePermissions(name string) ([]EtcdRolePermission, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	res, err := conn.Client.RoleGet(ctx, name)
	if err != nil {
		return []EtcdRolePermission{}, err
	}

	result := make([]EtcdRolePermission, len(res.Perm))
	for idx, _ := range res.Perm {
		perm := EtcdRolePermission{
			Permission: permissionEnumToPerm(clientv3.PermissionType(res.Perm[idx].PermType)), 
			Key: string(res.Perm[idx].Key), 
			RangeEnd: string(res.Perm[idx].RangeEnd),
		}
		result[idx] = perm
	}

	return result, nil
}

func (conn *EtcdConnection) DeleteRole(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout) * time.Second)
	defer cancel()

	_, err := conn.Client.RoleDelete(ctx, name)
	return err
}