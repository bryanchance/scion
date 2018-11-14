// Copyright 2018 Anapaya Systems

package pathstorage

import (
	"github.com/scionproto/scion/go/lib/common"
	"github.com/scionproto/scion/go/lib/pathdb"
	"github.com/scionproto/scion/go/lib/revcache"
)

func buildCombinedBackend(pdbConf PathDBConf,
	rcConf RevCacheConf) (pathdb.PathDB, revcache.RevCache, error) {

	if pdbConf.Connection == "" && rcConf.Connection == "" {
		return nil, nil, common.NewBasicError("Missing connection strings for pathdb/revcache", nil)
	}
	// if one config misses the connection string use the other one.
	if pdbConf.Connection == "" {
		pdbConf.Connection = rcConf.Connection
	} else if rcConf.Connection == "" {
		rcConf.Connection = pdbConf.Connection
	}
	pdb, err := newPathDB(pdbConf)
	if err != nil {
		return nil, nil, err
	}
	rc, err := newRevCache(rcConf)
	if err != nil {
		return nil, nil, err
	}
	return pdb, rc, nil
}
