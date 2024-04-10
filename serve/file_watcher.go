package serve

import (
	"context"
	"github.com/fsnotify/fsnotify"
	"log"
)

type FileWatcher struct {
}

func (fw *FileWatcher) Serve(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			log.Println("event:", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				log.Println("modified file:", event.Name)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			log.Println("error:", err)
		}
	}

	return nil
}
