package main

import (
	_ "github.com/DLag/logear/input/filetail"
	_ "github.com/DLag/logear/input/in_logear_forwarder"
	_ "github.com/DLag/logear/output/fluentd_forwarder"
	_ "github.com/DLag/logear/output/out_logear_forwarder"
)
