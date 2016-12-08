package rebolt

import "testing"

//TestUpdateRedis TestUpdateRedis
func TestUpdateRedis(t *testing.T) {
	InitDB(Config{
		RedisConf: &RedisConfig{
			Network:  "tcp",
			Addr:     "localhost:6379",
			PoolSize: 2,
		},
	})
	mydb, err := GetDB("redis", 0)
	if err != nil {
		t.FailNow()
	}
	TUpdateComm(mydb, t)
}
