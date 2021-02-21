package component

import (
	. "github.com/liangdas/mqant-modules/component/rsync"
	"github.com/liangdas/mqant/log"
	"hash/crc32"
	"time"
)

const (
	FULL = iota
	PATCH
	NEWEST  //无需同步
)

type InterDataSync interface {
	ResetData() error
	UpdateData()
	SyncDate()
	Marshal(table interface{}) ([]byte, int, error)
}

type SyncBytes interface {
	Source(table interface{}) ([]byte, error) //数据源
}

type DataSync struct {
	sub        SyncBytes
	original   []byte //上一次同步数据
	blockSize  int    //数据块大小，根据实际情况调整，默认可设置为 24
	syncDate   int64  //数据同步给客户端的日期
	updateData int64  //数据更新日期
}

func (ds *DataSync) OnInitDataSync(Sub SyncBytes, blockSize int) error {
	ds.sub = Sub
	ds.blockSize = blockSize
	return nil
}

/**
重置补丁
*/
func (ds *DataSync) ResetData() error {
	ds.original = nil
	ds.updateData = time.Now().UnixNano()
	return nil
}
func (ds *DataSync) UpdateData() {
	ds.updateData = time.Now().UnixNano()
}
func (ds *DataSync) SyncDate() {
	ds.syncDate = time.Now().UnixNano()
}

/**
补丁数据
*/
func (ds *DataSync) Marshal(table interface{}) ([]byte, int, error) {
	if ds.updateData <= ds.syncDate {
		return nil, NEWEST, nil
	}
	modified, err := ds.sub.Source(table)
	if err != nil {
		return nil, 0, err
	}
	if ds.original == nil {
		ds.original = modified
		ds.SyncDate()
		log.TInfo(nil, "ds.original=modified")
		return modified, FULL, err
	}
	rs := &LRsync{
		BlockSize: ds.blockSize,
	}
	hashes := rs.CalculateBlockHashes(ds.original)
	opsChannel := rs.CalculateDifferences(modified, hashes)
	ieee := crc32.NewIEEE()
	ieee.Write(modified)
	s := ieee.Sum32()
	delta := rs.CreateDelta(opsChannel, len(modified), s)
	ds.original = modified
	ds.SyncDate()
	return delta, PATCH, err
}
