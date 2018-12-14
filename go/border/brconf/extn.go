package brconf

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

const ExtnCfgName = "extn.json"

type ExtnConf struct {
	IFCfg map[common.IFIDType][]addr.IA
}

func ExtnLoadFromFile(path string) (*ExtnConf, error) {
	ec := &ExtnConf{IFCfg: make(map[common.IFIDType][]addr.IA)}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Not having the extra config file is ok, we just don't load it
			return ec, nil
		}
		return nil, common.NewBasicError("Unable to read extn config file", err)
	}
	if err := json.Unmarshal(b, ec); err != nil {
		return nil, common.NewBasicError("Unable to parse extn config file", err)
	}
	return ec, nil
}
