package callbacks

import (
	"github.com/moisespsena-go/assetfs/assetfsapi"
)

func FSLoaded(fs assetfsapi.Interface) {
	println("fs loaded")
}
