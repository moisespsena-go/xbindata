# embed example

XBindata example with embed assets.

## Install

    go get -u github.com/moisespsena-go/xbindata/...
    cd $GOPATH/src/github.com/moisespsena-go/xbindata/samples/embed
    
## Generate bindata and Data Store program

    xbindata -embed -pkg assets -prefix assets_root -o assets/assets.go assets_root/...
    
## Compile

    go build -o bin/embed
    
## Append asset contents into executable

    go run assets/data_store/main.go bin/embed


## Run

    cd bin
    ./embed -h