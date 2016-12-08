package rebolt

import (
	"bytes"
	"testing"
)

//TUpdateComm only used for test
func TUpdateComm(mydb IDB, t *testing.T) {
	mydb.Update(func(tx ITX) error {
		hash := []byte("hash")
		key, value := []byte("hello"), []byte("world")
		tx.Set(key, value)
		k1, v1 := []byte("hi"), []byte("hero")
		tx.Set(k1, v1)
		if bytes.Compare(value, tx.Get(key)) != 0 {
			t.Fail()
		}
		keys := tx.Keys("h*")
		if bytes.Compare(keys[0], k1) != 0 && bytes.Compare(keys[0], key) != 0 &&
			bytes.Compare(keys[0], hash) != 0 {
			t.Fail()
		}
		// tx.Del(key)
		// assert(len(tx.Get(key)) == 0)
		set := []byte("set")
		tx.SAdd(set, key)
		tx.SAdd(set, k1)
		tx.SAdd(set, value)
		ms := tx.SMembers(set)
		if bytes.Compare(ms[0], k1) != 0 && bytes.Compare(ms[0], key) != 0 &&
			bytes.Compare(ms[0], value) == 0 {
			t.Fail()
		}
		tx.SRem(set, k1)
		ms = tx.SMembers(set)
		if bytes.Compare(ms[0], key) != 0 &&
			bytes.Compare(ms[0], value) != 0 {
			t.Fail()
		}
		is1 := tx.SIsMember(set, k1)
		is2 := tx.SIsMember(set, key)
		if is1 || !is2 {
			t.Fail()
		}

		tx.HMSet(hash, key, value, k1, v1)
		ta := tx.HMGet(hash, key, k1)
		if bytes.Compare(ta[0], value) != 0 &&
			bytes.Compare(ta[0], v1) != 0 {
			t.Fail()
		}

		temp := tx.HGet(hash, key)
		if bytes.Compare(temp, value) != 0 {
			t.Fail()
		}
		k2, v2 := []byte("good"), []byte("work")
		tx.HSet(hash, k2, v2)
		temp = tx.HGet(hash, k2)
		if bytes.Compare(temp, v2) != 0 {
			t.Fail()
		}

		ehash := "ehash"
		tx.HMSet(ehash, key, value, "statusHost", "", "factory", "midea",
			"host", "http://jipengfei.com",
			"id", "uidxxx",
			"pwd", "whatpwd",
		)
		ta = tx.HGetAll(ehash)
		// log.Printf("ta=%s", ta)
		ta = tx.HMGet(ehash, key, "statusHost")
		// log.Printf("ta with statusHost=%s", ta)
		if bytes.Compare(ta[0], value) != 0 &&
			bytes.Compare(ta[0], []byte("")) != 0 {
			t.Fail()
		}

		temp = tx.HGet(ehash, "factory")
		// log.Printf("temp=%s, empty=%s", temp, []byte(""))
		if bytes.Compare(temp, []byte("midea")) != 0 {
			t.Fail()
		}

		mykeys := tx.Keys("*")
		for _, mk := range mykeys {
			tx.Del(mk)
		}
		if len(tx.Keys("*")) != 0 {
			t.Errorf("len(tx.Keys('*')) == %d", len(tx.Keys("*")))
		}
		return nil
	})
}
