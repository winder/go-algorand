package node

import (
	"bytes"
	"os"
	"path"

	"github.com/algorand/go-algorand/config"
	"github.com/algorand/go-algorand/util/codecs"
)

func GetConfigs(n *AlgorandFullNode) string {
	configs, err := os.ReadFile(path.Join(n.rootDir, "config.json"))
	if err != nil {
		//return "config.json file not found"
		var defaultConfig config.Local
		var buffer bytes.Buffer
		enc := codecs.NewFormattedJSONEncoder(&buffer)
		err = enc.Encode(defaultConfig)
		configs = buffer.Bytes()
	}
	return string(configs)
}
