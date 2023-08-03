package worker

import "github.com/spgnk/rtc/utils"

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
	w.logger.ERROR(v, tags)
}

func (w *workerLog) INFO(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["node_id"] = w.id
	tags["status"] = "info"
	w.logger.INFO(v, tags)
}

func (w *workerLog) STACK(v ...string) {
	w.logger.STACK(v...)
}

func (w *workerLog) WARN(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["node_id"] = w.id
	tags["status"] = "warn"
	w.logger.WARN(v, tags)
}

func (w *workerLog) DEBUG(v string, tags map[string]any) {
	if tags == nil {
		tags = make(map[string]any)
	}
	tags["node_id"] = w.id
	tags["status"] = "debug"
	w.logger.DEBUG(v, tags)
}
