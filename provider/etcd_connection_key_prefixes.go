package provider

import (
	"context"
	"errors"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
)

type PrefixesDiff struct {
	Upserts map[string]string
	Deletions []string
}

func (diff PrefixesDiff) IsEmpty() bool {
	return len(diff.Upserts) == 0 && len(diff.Deletions) == 0
}

func (conn *EtcdConnection) diffPrefixesWithRetries(srcPrefix string, dstPrefix string, retries int) (PrefixesDiff, error) {
	diffs := PrefixesDiff{
		Upserts: make(map[string]string),
		Deletions: []string{},
	}

	srcKeys, srcErr := conn.getKeyRangeWithRetries(srcPrefix, clientv3.GetPrefixRangeEnd(srcPrefix), retries)
	if srcErr != nil {
		return diffs, srcErr
	}

	dstKeys, dstErr := conn.getKeyRangeWithRetries(dstPrefix, clientv3.GetPrefixRangeEnd(dstPrefix), retries)
	if dstErr != nil {
		return diffs, dstErr
	}

	for key, _ := range dstKeys {
		suffix := strings.TrimPrefix(key, dstPrefix)
		if _, ok := srcKeys[srcPrefix + suffix]; !ok {
			diffs.Deletions = append(diffs.Deletions, suffix)
		}
	}

	for key, srcVal := range srcKeys {
		suffix := strings.TrimPrefix(key, srcPrefix)
		dstVal, ok := dstKeys[dstPrefix + suffix]
		if (!ok) || dstVal.Value != srcVal.Value {
			diffs.Upserts[suffix] = srcVal.Value
		}
	}

	return diffs, nil
}

func (conn *EtcdConnection) DiffPrefixes(srcPrefix string, dstPrefix string) (PrefixesDiff, error) {
	return conn.diffPrefixesWithRetries(srcPrefix, dstPrefix, conn.Retries)
}

func (conn *EtcdConnection) applyDiffToPrefixWithRetries(prefix string, diff PrefixesDiff, retries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(conn.Timeout)*time.Second)
	defer cancel()

	ops := []clientv3.Op{}

	for _, key := range diff.Deletions {
		ops = append(ops, clientv3.OpDelete(prefix + key))
	}

	for key, val := range diff.Upserts {
		ops = append(ops, clientv3.OpPut(prefix + key, val))
	}
	
	tx := conn.Client.Txn(ctx).Then(ops...)

	resp, txErr := tx.Commit()
	if txErr != nil {
		etcdErr, ok := txErr.(rpctypes.EtcdError)
		if !ok {
			return txErr
		}
		
		if etcdErr.Code() != codes.Unavailable || retries <= 0 {
			return txErr
		}

		time.Sleep(100 * time.Millisecond)
		return conn.applyDiffToPrefixWithRetries(prefix, diff, retries - 1)
	}

	if !resp.Succeeded {
		return errors.New("Transaction failed")
	}

	return nil
}

func (conn *EtcdConnection) ApplyDiffToPrefix(prefix string, diff PrefixesDiff) error {
	return conn.applyDiffToPrefixWithRetries(prefix, diff, conn.Retries)
}