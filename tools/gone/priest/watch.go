package priest

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func doWatch(fn func(string, string, fsnotify.Op), scanDirs []string) {
	watch(func(event fsnotify.Event) {
		if event.Op&fsnotify.Write == fsnotify.Write || event.Op == fsnotify.Remove {
			if filepath.Ext(event.Name) == ".go" {
				fn(filepath.Dir(event.Name), event.Name, event.Op)
			}
		}
	}, scanDirs)
}

var done chan any

func watch(fn func(event fsnotify.Event), dirs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer func(watcher *fsnotify.Watcher) {
		err := watcher.Close()
		if err != nil {
			log.Error(err)
		}
	}(watcher)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fn(event)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Error("error:", err)
			case <-done:
				return
			}
		}
	}()

	for _, dir := range dirs {
		watchRecursively(watcher, dir)
	}
}

func watchRecursively(watcher *fsnotify.Watcher, dir string) {
	log.Info("watch ", dir)
	err := watcher.Add(dir)
	if err != nil {
		log.Error(err)
		return
	}

	dirs, _ := os.ReadDir(dir)
	for i := range dirs {
		if dirs[i].IsDir() {
			watchRecursively(watcher, filepath.Join(dir, dirs[i].Name()))
		}
	}
}
