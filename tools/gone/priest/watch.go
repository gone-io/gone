package priest

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"sync"
)

func doWatch(fn func(string, string, fsnotify.Op), scanDirs []string, exclude string) {
	watch(func(event fsnotify.Event) {
		if event.Op&fsnotify.Write == fsnotify.Write || event.Op == fsnotify.Remove || event.Op == fsnotify.Create {
			if exclude != event.Name && filepath.Ext(event.Name) == ".go" {
				log.Infof("watch file(%s) changed", event.Name)
				fn(filepath.Dir(event.Name), event.Name, event.Op)
			}
		}
	}, scanDirs)
}

var done chan any
var mutex sync.Mutex

func getWatchDoneChannel() chan any {
	mutex.Lock()
	if done == nil {
		done = make(chan any)
	}
	mutex.Unlock()
	return done
}

func watch(fn func(event fsnotify.Event), dirs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				fn(event)
				//case err, ok := <-watcher.Errors:
				//	if !ok {
				//		return
				//	}
				//	log.Error("error:", err)
			}
		}
	}()

	for _, dir := range dirs {
		watchRecursively(watcher, dir)
	}

	<-getWatchDoneChannel()
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
