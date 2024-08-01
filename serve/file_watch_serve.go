package serve

import (
	"context"
	"github.com/expgo/config"
	"github.com/expgo/log"
	"github.com/expgo/serve"
	"github.com/fsnotify/fsnotify"
	"github.com/thejerf/suture/v4"
)

//go:generate ag

func init() {
	// init FileWatcher singleton
	__FileWatchServe()
}

// @Singleton(LocalGetter)
type FileWatchServe struct {
	watcher *fsnotify.Watcher
	log.InnerLog
}

func (fws *FileWatchServe) Init() {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		fws.L.Fatal(err)
	}

	fws.watcher = watcher

	serve.AddServe(fws, "ConfigFileWatcher", suture.Spec{})

	config.AddFileWatcher(fws)
}

func (fws *FileWatchServe) WatchFile(filename string) {
	_ = fws.watcher.Add(filename)
}

func (fws *FileWatchServe) Close() {
	if fws.watcher != nil {
		err := fws.watcher.Close()
		if err != nil {
			fws.L.Warnf("close watcher err: %v", err)
		}
	}
}

func (fws *FileWatchServe) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-fws.watcher.Events:
			if !ok {
				return nil
			}

			fws.L.Debugf("event: %v", event)

			if event.Has(fsnotify.Write) {
				fws.L.Infof("modified file: %v", event.Name)
				if err := config.FileUpdate(event.Name); err != nil {
					fws.L.Errorf("update file err: %v", err)
				}
			}
		case wErr, ok := <-fws.watcher.Errors:
			if !ok {
				return nil
			}
			fws.L.Warnf("error: %v", wErr)
		}
	}
}
