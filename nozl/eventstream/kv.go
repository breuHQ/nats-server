package eventstream

import (
	"strconv"

	"github.com/nats-io/nats-server/v2/nozl/shared"
)

func GetMultValIntKVstore(KVstoreName string, keyAll []string) map[string]int {
	valMap := make(map[string]int)
	kv, err := Eventstream.RetreiveKeyValStore(KVstoreName)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	for _, val := range keyAll {
		byteVal, err := kv.Get(val)
		if err != nil {
			shared.Logger.Error(err.Error())
		}

		valMap[val], err = strconv.Atoi(string(byteVal.Value()))
		if err != nil {
			shared.Logger.Error(err.Error())
		}
	}
	return valMap
}
