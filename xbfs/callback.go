package xbfs

import "github.com/moisespsena-go/assetfs/assetfsapi"

type FsLoadCallback interface {
	Callback(fs assetfsapi.Interface)
}

type FsLoadCallbackFunc func(fs assetfsapi.Interface)

func (f FsLoadCallbackFunc) Callback(fs assetfsapi.Interface) {
	f(fs)
}
