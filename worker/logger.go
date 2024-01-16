package worker

import (
	"github.com/lamhai1401/gologs/logs"
	"github.com/spgnk/rtc/utils"
)

type workerLog struct {
	id     string
	logger utils.Log
}

func (w *workerLog) ERROR(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["node_id"] = w.id
	tags["status"] = "error"
	if w.logger == nil {
		logs.Error(v, tags)
	} else {
		w.logger.ERROR(v, tags)
	}
}

func (w *workerLog) INFO(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["node_id"] = w.id
	tags["status"] = "info"
	if w.logger == nil {
		logs.Error(v, tags)
	} else {
		w.logger.INFO(v, tags)
	}
}

func (w *workerLog) STACK(v ...string) {
	if w.logger == nil {
		logs.Stack(v...)
	} else {
		w.logger.STACK(v...)
	}
}

func (w *workerLog) WARN(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["node_id"] = w.id
	tags["status"] = "warn"
	if w.logger == nil {
		logs.Error(v, tags)
	} else {
		w.logger.WARN(v, tags)
	}
}

func (w *workerLog) DEBUG(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["node_id"] = w.id
	tags["status"] = "debug"

	if w.logger == nil {
		logs.Error(v, tags)
	} else {
		w.logger.DEBUG(v, tags)
	}
}
