package main

import (
	"k8s.io/klog"
)

type klogAdapter struct {
}

func (*klogAdapter) Println(args ...interface{}) {
	klog.Infoln(args...)
}

func (*klogAdapter) Printf(format string, args ...interface{}) {
	klog.Infof(format, args...)
}
