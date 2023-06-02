package core

import (
	"github.com/nats-io/nats-server/v2/nozl/eventstream"
	"github.com/nats-io/nats-server/v2/nozl/shared"
)

// Function to that take in the KV store name and the name of keys that want to be
// retrieved. It then iterates over those keys and returns their values as a map
func GetMultValKVstore(KVstoreName string, keyAll []string) map[string][]byte {
	valMap := make(map[string][]byte)
	kv, err := eventstream.Eventstream.RetreiveKeyValStore(KVstoreName)
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
