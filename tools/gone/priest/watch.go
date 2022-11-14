package priest

import (
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"path"
)

func doWatch(fn func(string, string, fsnotify.Op), scanDirs []string) {
	watch(func(event fsnotify.Event) {
		if event.Op&fsnotify.Write == fsnotify.Write || event.Op == fsnotify.Remove {
			if path.Ext(event.Name) == ".go" {
				fn(path.Dir(event.Name), event.Name, event.Op)
			}
		}
	}, scanDirs)
}

func watch(fn func(event fsnotify.Event), dirs []string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error(err)
	}
	defer watcher.Close()

	done := make(chan bool)
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
			}
		}
	}()

	for _, dir := range dirs {
		watchRecursively(watcher, dir)
	}
	<-done
}

func watchRecursively(watcher *fsnotify.Watcher, dir string) {
	log.Info("watch ", dir)
	err := watcher.Add(dir)
	if err != nil {
		log.Error(err)
		return
	}

	dirs, _ := ioutil.ReadDir(dir)
	for i := range dirs {
		if dirs[i].IsDir() {
			watchRecursively(watcher, path.Join(dir, dirs[i].Name()))
		}
	}
}
