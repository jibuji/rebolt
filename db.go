// Copyright 2016 jibuji. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rebolt encapsulate bolt and redis operation into a tiny, common,
// and easy to use operations, so you can change your db between bolt and redis
// without pains.
//
// Remmeber the key and value types in both bolt and redis:
// `string, bool`, any integer types, any float types, and arrays of this types.
// The underhood representation is `[]byte`, it's your responsibility to interpret it.
package rebolt

import (
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

//IDB database interface
type IDB interface {
	// Update with write or read operation to the db, you should
	// put your db-ops into the callback function implementations.
	// If you just want some read operation on the db, please using
	// View function, this will give you more concurrency-efficient.
	Update(callback func(tx ITX) error)

	//View with only read operation, you should
	// put your db-ops into the callback function implementations.
	// Write to db here is not allowed.
	View(callback func(tx ITX) error)
}

//ITX transaction interface, you can grab ITX interface handle in `Update` or
// `View` functions.
type ITX interface {
	//Set stores the key-value pair in db.
	Set(key interface{}, value interface{})
	//Get retrieves the value by the given key.
	Get(key interface{}) []byte
	//Del rm the key-value pair by the given key
	Del(key interface{})
	//Keys get all keys that match the given `glob` pattern
	Keys(pattern interface{}) [][]byte
	//SIsMember return if the field is in the set indicated by the given key
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
