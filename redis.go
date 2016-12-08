package rebolt

import (
	"errors"
	"log"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

var gRedisConfig RedisConfig

func configureRedis(conf *RedisConfig) error {
	gRedisConfig.Network = conf.Network
	gRedisConfig.Addr = conf.Addr
	gRedisConfig.PoolSize = conf.PoolSize

	p, err := pool.New(gRedisConfig.Network, gRedisConfig.Addr, gRedisConfig.PoolSize)
	if err != nil {
		return err
	}
	p.Empty()
	return nil
}

//redisDB IDB implementation
type redisDB struct {
	pool *pool.Pool
}

//Close close the pool
func (db redisDB) Close() {
	db.pool.Empty()
}

//Update read-write ops
func (db redisDB) Update(f func(tx ITX) error) {
	op(db.pool, f, true)
}

//View read ops
func (db redisDB) View(f func(tx ITX) error) {
	op(db.pool, f, false)
}

func op(p *pool.Pool, f func(tx ITX) error, rw bool) {
	conn, err := p.Get()
	if err != nil {
		panic(err)
	}
	defer p.Put(conn)

	rtx := redisTx{
		conn: conn,
		rw:   rw,
	}
	for {
		err = f(rtx)
		//if err is connection error, do recconnet
		if err == nil {
			break
		}
		log.Printf("redis DB op error=%s, try again", err)
	}
}

type redisTx struct {
	conn       *redis.Client
	rw         bool
	flatBucket []byte
}

func (tx redisTx) Set(key, val interface{}) {
	tx.conn.Cmd("SET", key, val)
}

func (tx redisTx) Get(key interface{}) []byte {
	resp := tx.conn.Cmd("GET", key)
	val, _ := resp.Bytes()
	return val
}

func (tx redisTx) Del(key interface{}) {
	tx.conn.Cmd("DEL", key)
}

func (tx redisTx) SIsMember(key, field interface{}) bool {
	resp := tx.conn.Cmd("SISMEMBER", key, field)
	is, _ := resp.Int()
	return is == 1
}

func (tx redisTx) SMembers(key interface{}) [][]byte {
	resp := tx.conn.Cmd("SMEMBERS", key)
	ms, _ := resp.ListBytes()
	return ms
}

func (tx redisTx) SAdd(key, member interface{}) {
	tx.conn.Cmd("SADD", key, member)
}

func (tx redisTx) SRem(key, member interface{}) {
	tx.conn.Cmd("SREM", key, member)
}

func (tx redisTx) HGet(key, field interface{}) []byte {
	resp := tx.conn.Cmd("HGET", key, field)
	val, _ := resp.Bytes()
	return val
}

func (tx redisTx) HSet(key, field, value interface{}) {
	tx.conn.Cmd("HSET", key, field, value)
}

func (tx redisTx) HMSet(frs ...interface{}) {
	tx.conn.Cmd("HMSET", frs...)
}

func (tx redisTx) HMGet(kfields ...interface{}) [][]byte {
	resp := tx.conn.Cmd("HMGET", kfields...)
	ba, _ := resp.ListBytes()
	return ba
}

func (tx redisTx) HGetAll(key interface{}) [][]byte {
	resp := tx.conn.Cmd("HGETALL", key)
	ba, _ := resp.ListBytes()
	return ba
}

func (tx redisTx) Keys(pattern interface{}) [][]byte {
	resp := tx.conn.Cmd("Keys", pattern)
	ba, _ := resp.ListBytes()
	return ba
}

func (tx redisTx) Watch(keys ...interface{}) {
	tx.conn.Cmd("Watch", keys...)
}

func (tx redisTx) Multi() {
	tx.conn.Cmd("Multi")
}

func (tx redisTx) Exec() error {
	resp := tx.conn.Cmd("Exec")
	if resp.Err != nil {
		tx.conn.Cmd("DISCARD")
		return errors.New("exec error")
	}
	if resp.IsType(redis.Nil) {
		return errors.New("exec aborted")
	}
	return nil
}

var dbMap = map[int]IDB{}

//getRedisDB getRedisDB
func getRedisDB(dbIndex int) (IDB, error) {
	mydb, _ := dbMap[dbIndex]

	if mydb != nil {
		return mydb, nil
	}
	pool, err := createPool(dbIndex)
	db := redisDB{
		pool: pool,
	}
	if err != nil {
		dbMap[dbIndex] = db
	}
	return db, err
}

//CreatePool create a redis pool for global usage.
func createPool(dbIndex int) (*pool.Pool, error) {
	p, err := pool.NewCustom(gRedisConfig.Network, gRedisConfig.Addr, gRedisConfig.PoolSize,
		func(network, addr string) (*redis.Client, error) {
			conn, err := redis.Dial(network, addr)
			if err == nil {
				conn.Cmd("SELECT", dbIndex)
			}
			return conn, err
		},
	)

	if err != nil {
		return nil, err
	}
	//simple tests
	conn, err := p.Get()
	if err != nil {
		return nil, err
	}
	defer p.Put(conn)
	return p, nil
}
