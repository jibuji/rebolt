package rebolt

import (
	"errors"
	"fmt"
	"path"

	"github.com/boltdb/bolt"
)

var mydb *bolt.DB

//Configure configure bolt
func configureBolt(conf *BoltConfig) error {
	if len(conf.DBPath) <= 0 {
		return errors.New("bolt needs dbPath, please specified in the config")
	}
	mode := conf.OpenMode
	if mode == 0 {
		mode = 0600
	}

	db, err := bolt.Open(conf.DBPath, mode, conf.Options)
	if err != nil {
		return err
	}
	mydb = db
	return nil
}

func getBoltDB(dbIndex int) (IDB, error) {
	if mydb == nil {
		return nil, errors.New("You need call rebolt.InitDB to config rebolt ")
	}
	name := fmt.Sprintf("db%d", dbIndex)
	bktName := []byte(name)
	mydb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bktName)
		if err != nil {
			panic(err)
		}
		return nil
	})
	return boltDB{
		topBktName: bktName,
	}, nil
}

type boltDB struct {
	topBktName []byte
}

func (db boltDB) Update(f func(ITX) error) {
	mydb.Update(func(tx *bolt.Tx) error {
		return f(&boltTx{
			tx:         tx,
			topBktName: db.topBktName,
		})
	})
}

func (db boltDB) View(f func(ITX) error) {
	mydb.View(func(tx *bolt.Tx) error {
		return f(boltTx{
			tx:         tx,
			topBktName: db.topBktName,
		})
	})
}

type boltTx struct {
	tx         *bolt.Tx
	topBktName []byte
}

var singleKey = []byte("0")

func (btx boltTx) Set(key interface{}, val interface{}) {
	kbs := convertToByte(key)
	vbs := convertToByte(val)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt, _ := tbkt.CreateBucketIfNotExists(kbs)
	if bkt != nil {
		bkt.Put(singleKey, vbs)
	}
}

func (btx boltTx) Get(key interface{}) []byte {
	kbs := convertToByte(key)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt := tbkt.Bucket(kbs)
	if bkt == nil {
		return nil
	}
	return bkt.Get(singleKey)
}
func (btx boltTx) Del(key interface{}) {
	kbs := convertToByte(key)
	tbkt := btx.tx.Bucket(btx.topBktName)
	tbkt.DeleteBucket(kbs)
}

func (btx boltTx) Keys(pattern interface{}) [][]byte {
	pbs := convertToString(pattern)
	var ret [][]byte
	tbkt := btx.tx.Bucket(btx.topBktName)
	tbkt.ForEach(func(name []byte, val []byte) error {
		success, _ := path.Match(pbs, string(name))
		if success {
			ret = append(ret, name)
		}
		return nil
	})
	return ret
}

//
func (btx boltTx) SIsMember(key, field interface{}) bool {
	kbs := convertToByte(key)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt := tbkt.Bucket(kbs)
	if bkt == nil {
		return false
	}
	fbs := convertToByte(field)
	val := bkt.Get(fbs)
	return len(val) > 0
}

func (btx boltTx) SMembers(key interface{}) [][]byte {
	kbs := convertToByte(key)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt := tbkt.Bucket(kbs)
	if bkt == nil {
		return nil
	}
	var ret [][]byte
	bkt.ForEach(func(k []byte, v []byte) error {
		ret = append(ret, k)
		return nil
	})
	return ret
}

func (btx boltTx) SAdd(key, member interface{}) {
	kbs := convertToByte(key)
	mbs := convertToByte(member)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt, _ := tbkt.CreateBucketIfNotExists(kbs)
	if bkt != nil {
		bkt.Put(mbs, singleKey)
	}
}

func (btx boltTx) SRem(key, member interface{}) {
	kbs := convertToByte(key)
	mbs := convertToByte(member)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt := tbkt.Bucket(kbs)
	if bkt != nil {
		bkt.Delete(mbs)
	}
}

func (btx boltTx) HGet(key, field interface{}) []byte {
	kbs := convertToByte(key)
	fbs := convertToByte(field)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt := tbkt.Bucket(kbs)
	if bkt == nil {
		return nil
	}
	return bkt.Get(fbs)
}

func (btx boltTx) HSet(key, field, value interface{}) {
	kbs := convertToByte(key)
	fbs := convertToByte(field)
	vbs := convertToByte(value)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt, _ := tbkt.CreateBucketIfNotExists(kbs)
	if bkt != nil {
		bkt.Put(fbs, vbs)
	}
}

// //params:key, field1, val1,  ....
func (btx boltTx) HMSet(kfrs ...interface{}) {
	kbs := convertToByte(kfrs[0])
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt, _ := tbkt.CreateBucketIfNotExists(kbs)
	if bkt == nil {
		return
	}
	length := len(kfrs)
	for i := 2; i < length; i += 2 {
		field := convertToByte(kfrs[i-1])
		value := convertToByte(kfrs[i])
		// if len(field) > 0 && len(value) > 0 {
		//when put zero value of [value], bolt would behave abnormally
		bkt.Put(field, value)
		// }
	}
}

// //params:key, field1, field2,  ....
func (btx boltTx) HMGet(kfields ...interface{}) [][]byte {
	kbs := convertToByte(kfields[0])
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt := tbkt.Bucket(kbs)
	if bkt == nil {
		return nil
	}
	length := len(kfields)
	var ret [][]byte
	for i := 1; i < length; i++ {
		field := convertToByte(kfields[i])
		value := bkt.Get(field)
		ret = append(ret, value)
	}
	return ret
}

func (btx boltTx) HGetAll(key interface{}) [][]byte {
	kbs := convertToByte(key)
	tbkt := btx.tx.Bucket(btx.topBktName)
	bkt := tbkt.Bucket(kbs)
	if bkt == nil {
		return nil
	}
	var ret [][]byte
	bkt.ForEach(func(k []byte, v []byte) error {
		ret = append(ret, k, v)
		return nil
	})
	return ret
}

func (btx boltTx) Watch(keys ...interface{}) {
	//Do nothing, because transaction is done by db.Update or db.View

}

func (btx boltTx) Multi() {
	//Do nothing, because transaction is done by db.Update or db.View

}

func (btx boltTx) Exec() error {
	//Do nothing, because transaction is done by db.Update or db.View
	return nil
}

func convertToByte(ib interface{}) []byte {
	if bs, ok := ib.([]byte); ok {
		return bs
	}
	if s, ok := ib.(string); ok {
		return []byte(s)
	}
	msg := fmt.Sprintf("db.convertToByte can't conver (%s) to []byte", ib)
	panic(msg)
}

func convertToString(ib interface{}) string {
	if bs, ok := ib.([]byte); ok {
		return string(bs)
	}
	if s, ok := ib.(string); ok {
		return s
	}
	msg := fmt.Sprintf("db.convertToString can't conver (%s) to string", ib)
	panic(msg)
}
