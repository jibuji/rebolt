package rebolt

import "testing"

func TestUpdateBolt(t *testing.T) {
	InitDB(Config{
		BoltConf: &BoltConfig{
			DBPath: "mybolt.db",
		},
	})
	mydb, err := GetDB("bolt", 0)
	if err != nil {
		t.FailNow()
	}
	TUpdateComm(mydb, t)
}
