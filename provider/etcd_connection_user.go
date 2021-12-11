package provider

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
)

func (conn *EtcdConnection) listUsersWithRetries(retries int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	res, err := conn.Client.UserList(ctx)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return []string{}, err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return []string{}, err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.listUsersWithRetries(retries - 1)
	}

	return res.Users, nil
}

func (conn *EtcdConnection) ListUsers() ([]string, error) {
	return conn.listUsersWithRetries(conn.Retries)
}

func (conn *EtcdConnection) addUserWithRetries(username string, password string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.UserAdd(ctx, username, password)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.addUserWithRetries(username, password, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) AddUser(username string, password string) error {
	return conn.addUserWithRetries(username, password, conn.Retries)
}

func (conn *EtcdConnection) getUserRolesWithRetries(username string, retries int) ([]string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	res, err := conn.Client.UserGet(ctx, username)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return []string{}, false, err
		}

		if etcdErr.Error() == rpctypes.ErrorDesc(rpctypes.ErrGRPCUserNotFound) {
			return []string{}, false, nil
		}

		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return []string{}, false, err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.getUserRolesWithRetries(username, retries - 1)
	}

	return res.Roles, true, nil
}

func (conn *EtcdConnection) GetUserRoles(username string) ([]string, bool, error) {
	return conn.getUserRolesWithRetries(username, conn.Retries)
}

func (conn *EtcdConnection) changeUserPasswordWithRetries(username string, password string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.UserChangePassword(ctx, username, password)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.changeUserPasswordWithRetries(username, password, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) ChangeUserPassword(username string, password string) error {
	return conn.changeUserPasswordWithRetries(username, password, conn.Retries)
}

func (conn *EtcdConnection) grantUserRoleWithRetries(username string, role string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.UserGrantRole(ctx, username, role)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.grantUserRoleWithRetries(username, role, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) GrantUserRole(username string, role string) error {
	return conn.grantUserRoleWithRetries(username, role, conn.Retries)
}

func (conn *EtcdConnection) revokeUserRoleWithRetries(username string, role string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.UserRevokeRole(ctx, username, role)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.revokeUserRoleWithRetries(username, role, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) RevokeUserRole(username string, role string) error {
	return conn.revokeUserRoleWithRetries(username, role, conn.Retries)
}

func (conn *EtcdConnection) deleteUserWithRetries(username string, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	_, err := conn.Client.UserDelete(ctx, username)
	if err != nil {
		etcdErr, ok := err.(rpctypes.EtcdError)
		if !ok {
			return err
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return err
		}

		time.Sleep(100 * time.Millisecond)
		return conn.deleteUserWithRetries(username, retries - 1)
	}

	return nil
}

func (conn *EtcdConnection) DeleteUser(username string) error {
	return conn.deleteUserWithRetries(username, conn.Retries)
}