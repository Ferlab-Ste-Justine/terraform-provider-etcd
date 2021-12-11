package provider

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

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

func (conn *EtcdConnection) listRolesWithRetries(retries int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	res, err := conn.Client.RoleList(ctx)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return []string{}, err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return []string{}, err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.listRolesWithRetries(retries - 1)
	}

	return res.Roles, nil
}

func (conn *EtcdConnection) ListRoles() ([]string, error) {
	return conn.listRolesWithRetries(conn.Retries)
}

func (conn *EtcdConnection) addRoleWithRetries(name string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.RoleAdd(ctx, name)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.addRoleWithRetries(name, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) AddRole(name string) error {
	return conn.addRoleWithRetries(name, conn.Retries)
}

func (conn *EtcdConnection) grantRolePermissionWithRetries(name string, key string, rangeEnd string, permission string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.RoleGrantPermission(ctx, name, key, rangeEnd, permissionToEnum(permission))
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.grantRolePermissionWithRetries(name, key, rangeEnd, permission, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) GrantRolePermission(name string, key string, rangeEnd string, permission string) error {
	return conn.grantRolePermissionWithRetries(name, key, rangeEnd, permission, conn.Retries)
}

func (conn *EtcdConnection) revokeRolePermissionWithRetries(name string, key string, rangeEnd string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.RoleRevokePermission(ctx, name, key, rangeEnd)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.revokeRolePermissionWithRetries(name, key, rangeEnd, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) RevokeRolePermission(name string, key string, rangeEnd string) error {
	return conn.revokeRolePermissionWithRetries(name, key, rangeEnd, conn.Retries)
}

func (conn *EtcdConnection) getRolePermissionsWithRetries(name string, retries int) ([]EtcdRolePermission, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	res, err := conn.Client.RoleGet(ctx, name)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return []EtcdRolePermission{}, false, err
		}

		if etcdErr.Error() == rpctypes.ErrorDesc(rpctypes.ErrGRPCRoleNotFound) {
			return []EtcdRolePermission{}, false, nil
		}

		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return []EtcdRolePermission{}, false, err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.getRolePermissionsWithRetries(name, retries - 1)
	}

	result := make([]EtcdRolePermission, len(res.Perm))
	for idx, _ := range res.Perm {
		perm := EtcdRolePermission{
			Permission: permissionEnumToPerm(clientv3.PermissionType(res.Perm[idx].PermType)),
			Key:        string(res.Perm[idx].Key),
			RangeEnd:   string(res.Perm[idx].RangeEnd),
		}
		result[idx] = perm
	}

	return result, true, nil
}

func (conn *EtcdConnection) GetRolePermissions(name string) ([]EtcdRolePermission, bool, error) {
	return conn.getRolePermissionsWithRetries(name, conn.Retries)
}

func (conn *EtcdConnection) deleteRoleWithRetries(name string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.RoleDelete(ctx, name)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.deleteRoleWithRetries(name, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) DeleteRole(name string) error {
	return conn.deleteRoleWithRetries(name, conn.Retries)
}
