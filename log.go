package main

import (
	"k8s.io/klog"
)

type klogAdapter struct {
}

func (*klogAdapter) Infoln(args ...interface{}) {
	klog.Infoln(args...)
}

func (*klogAdapter) Infof(format string, args ...interface{}) {
	klog.Infof(format, args...)
}
func (*klogAdapter) Errorln(args ...interface{}) {
	klog.Errorln(args...)
}

func (*klogAdapter) Errorf(format string, args ...interface{}) {
	klog.Errorf(format, args...)
}
