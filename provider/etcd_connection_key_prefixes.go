package provider

import (
	"context"
	"errors"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
)

func (diff KeysDiff) IsEmpty() bool {
	return len(diff.Upserts) == 0 && len(diff.Deletions) == 0
}

func (conn *EtcdConnection) diffPrefixesWithRetries(srcPrefix string, dstPrefix string, retries int) (KeysDiff, error) {
	srcKeys, srcErr := conn.getKeyRangeWithRetries(srcPrefix, clientv3.GetPrefixRangeEnd(srcPrefix), retries)
	if srcErr != nil {
		return KeysDiff{}, srcErr
	}

	dstKeys, dstErr := conn.getKeyRangeWithRetries(dstPrefix, clientv3.GetPrefixRangeEnd(dstPrefix), retries)
	if dstErr != nil {
		return KeysDiff{}, dstErr
	}

	return GetKeysDiff(srcKeys, srcPrefix, dstKeys, dstPrefix), nil
}

func (conn *EtcdConnection) DiffPrefixes(srcPrefix string, dstPrefix string) (KeysDiff, error) {
	return conn.diffPrefixesWithRetries(srcPrefix, dstPrefix, conn.Retries)
}

func (conn *EtcdConnection) applyDiffToPrefixWithRetries(prefix string, diff KeysDiff, retries int) error {
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

func (conn *EtcdConnection) ApplyDiffToPrefix(prefix string, diff KeysDiff) error {
	return conn.applyDiffToPrefixWithRetries(prefix, diff, conn.Retries)
}

func (conn *EtcdConnection) DiffPrefixWithInput(prefix string, inputKeys map[string]KeyInfo, inputKeysPrefix string, inputIsSource bool) (KeysDiff, error) {
	prefixKeys, err := conn.getKeyRangeWithRetries(prefix, clientv3.GetPrefixRangeEnd(prefix), conn.Retries)
	if err != nil {
		return KeysDiff{}, err
	}

	if inputIsSource {
		return GetKeysDiff(inputKeys, inputKeysPrefix, prefixKeys, prefix), nil
	}

	return GetKeysDiff(prefixKeys, prefix, inputKeys, inputKeysPrefix), nil
}