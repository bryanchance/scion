package conf

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/scionproto/scion/go/lib/addr"
	"github.com/scionproto/scion/go/lib/common"
)

const ExtnCfgName = "extn.json"

type ExtnConf struct {
	IFCfg map[common.IFIDType][]*addr.ISD_AS
}

func ExtnLoadFromFile(path string) (*ExtnConf, error) {
	ec := &ExtnConf{IFCfg: make(map[common.IFIDType][]*addr.ISD_AS)}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Not having the extra config file is ok, we just don't load it
			return ec, nil
		}
		return nil, common.NewCError("Unable to read extn config file", "err", err)
	}
	if err := json.Unmarshal(b, ec); err != nil {
		return nil, common.NewCError("Unable to parse extn config file", "err", err)
	}
	return ec, nil
}
