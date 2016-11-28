package rebolt

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

//IDB database interface
type IDB interface {
	Update(func(tx ITX) error)
	View(func(tx ITX) error)
}

//ITX transaction interface
type ITX interface {
	Set(interface{}, interface{})
	Get(interface{}) []byte
	Del(interface{})
	Keys(pattern interface{}) [][]byte
	SIsMember(key, field interface{}) bool
	SMembers(key interface{}) [][]byte
	SAdd(key, member interface{})
	SRem(key, member interface{})
	HGet(key, field interface{}) []byte
	HSet(key, field, value interface{})
	//params:key, field1, val1,  ....
	HMSet(kfrs ...interface{})
	//params:key, field1, field2,  ....
	HMGet(kfields ...interface{}) [][]byte
	HGetAll(key interface{}) [][]byte
	Watch(keys ...interface{})
	Multi()
	Exec() error
}

//GetDB GetDB for
func GetDB(dbType string, dbIndex int) (db IDB, err error) {
	switch dbType {
	case "bolt":
		db, err = getBoltDB(dbIndex)
	case "redis":
		db, err = getRedisDB(dbIndex)
	}
	if err != nil {
		msg := fmt.Sprintf("OpenDB err=%v", err)
		log.Println(msg)
	}
	return
}

//BoltConfig config for bolt
type BoltConfig struct {
	DBPath   string
	OpenMode os.FileMode
	Options  *bolt.Options
}

//RedisConfig config for redis
type RedisConfig struct {
	Network  string
	Addr     string
	PoolSize int
}

//Config  config for both bolt and redis
type Config struct {
	BoltConf  *BoltConfig
	RedisConf *RedisConfig
}

//InitDB for init db stuff
func InitDB(conf Config) error {
	if conf.BoltConf != nil {
		err := configureBolt(conf.BoltConf)
		if err != nil {
			return err
		}
	}
	if conf.RedisConf != nil {
		err := configureRedis(conf.RedisConf)
		if err != nil {
			return err
		}
	}
	return nil
}
