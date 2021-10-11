#!/bin/sh
ROOT_USER=""
while [ "$ROOT_USER" != "root" ]; do
    sleep 1
    etcdctl user add --no-password --cacert=/opt/certs/ca.pem --endpoints=https://127.0.0.1:2379 --insecure-transport=false --cert=/opt/certs/root.pem --key=/opt/certs/root.key root
    ROOT_USER=$(etcdctl user list --cacert=/opt/certs/ca.pem --endpoints=https://127.0.0.1:2379 --insecure-transport=false --cert=/opt/certs/root.pem --key=/opt/certs/root.key | grep root)
done
ROOT_ROLES=""
while [ -z "$ROOT_ROLES" ]; do
    sleep 1
    etcdctl user grant-role --cacert=/opt/certs/ca.pem --endpoints=https://127.0.0.1:2379 --insecure-transport=false --cert=/opt/certs/root.pem --key=/opt/certs/root.key root root
    ROOT_ROLES=$(etcdctl user get --cacert=/opt/certs/ca.pem --endpoints=https://127.0.0.1:2379 --insecure-transport=false --cert=/opt/certs/root.pem --key=/opt/certs/root.key root | grep "Roles: root")
done
etcdctl auth enable --cacert=/opt/certs/ca.pem --endpoints=https://127.0.0.1:2379 --insecure-transport=false --cert=/opt/certs/root.pem --key=/opt/certs/root.key
while [ $? -ne 0 ]; do
    sleep 1
    etcdctl auth enable --cacert=/opt/certs/ca.pem --endpoints=https://127.0.0.1:2379 --insecure-transport=false --cert=/opt/certs/root.pem --key=/opt/certs/root.key
done