//+build debug

package util

import (
	"free5gc/lib/path_util"
)

var UdrLogPath = path_util.Gofree5gcPath("free5gc/udrsslkey.log")
var UdrPemPath = path_util.Gofree5gcPath("free5gc/support/TLS/_debug.pem")
var UdrKeyPath = path_util.Gofree5gcPath("free5gc/support/TLS/_debug.key")
var DefaultUdrConfigPath = path_util.Gofree5gcPath("free5gc/config/udrcfg.conf")
