package main

import (
	"context"
	"fmt"
	"os"
	"runtime"

	flag "github.com/bborbe/flagenv"
	"github.com/bborbe/world/configuration"
	"github.com/bborbe/world/pkg/uploader"
	"github.com/golang/glog"
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	glog.V(1).Infof("building all ...")
	uploader := &uploader.UploadAll{
		Apps: configuration.Apps(),
	}
	if err := uploader.Upload(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v", err)
		os.Exit(1)
	}
	glog.V(1).Infof("building all finished")
}
