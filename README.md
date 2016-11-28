# rebolt

A refined common interface for [bolt](https://github.com/boltdb/bolt) and [redis](https://redis.io/).

## Motivation:
  One of my project was a server end service, it used redis as cache. But, for some reason,
  I have to port it on the client end, covering Windows OS ranging from windows 7 to windows10. Redis doesn't have windows release officially, so I decide to switch the
  cache db from redis to bolt.   

  But, redis interface and bolt interface are very different. Additionally, to my opinion,
  redis has a good design, but most go library's interface for redis isn't intuitive. Bolt has a good interface, but doesn't provide redis-like data structures such as hashes, sets, etc, so it isn't good to use, either.

  To improve this situation, I designed an NEW interface, took the good part of both bolt
  and redis.

## Limitation:

  * This is a tiny set of interfaces that bridge the gap between redis and bolt. But it is still not complete, some data structures, such as lists, bitmaps, sorted sets are currently not supported.

  * Welcome your suggestions and contributions.

## Dependencies:

  * [radix.v2](https://github.com/mediocregopher/radix.v2) for interacting with redis server.

  * [boltdb](https://github.com/boltdb/bolt)

## Show API By Example:

  0. Acquire the lib:

    ```
    go get github.com/jibuji/rebolt
    ```

  1. Configure DB:

  ``` go

  rebolt.InitDB(rebolt.Config{
		BoltConf: &rebolt.BoltConfig{
			DBPath: "mybolt.db",
		},
		RedisConf: &rebolt.RedisConfig{
			Network:  "tcp",
			Addr:     "localhost:6379",
			PoolSize: 2,
		},
	})

  ```

  2. read and write from/to db

  ```go

  mydb, err := GetDB("redis", 0)
	if err != nil {
		panic(err)
	}
	mydb.Update(func(tx rebolt.ITX) error {
		key, value := []byte("hello"), []byte("world")
		tx.Set(key, value)
		if bytes.Compare(value, tx.Get(key)) != 0 {
			return errors.New("gotten wrong value!");
		}
    return nil
  })

  ```

3. More usage example, please refer

 [`bolt_test.go`](https://github.com/jibuji/rebolt/blob/master/bolt_test.go)
