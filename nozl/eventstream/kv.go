package eventstream

import (
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

func GetMultValIntKVstore(KVstoreName string, keyAll []string) map[string][]byte {
	valMap := make(map[string][]byte)
	kv, err := Eventstream.RetreiveKeyValStore(KVstoreName)
	if err != nil {
		shared.Logger.Error(err.Error())
	}

	for _, val := range keyAll {
		byteVal, err := kv.Get(val)
		if err != nil {
			shared.Logger.Error(err.Error())
		}

		valMap[val] = byteVal.Value()
	}
	return valMap
}
