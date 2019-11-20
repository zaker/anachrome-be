package services

import (
	"log"

	"github.com/fsnotify/fsnotify"
)

//WatchDir gives of event after dirs changed
func WatchDir(dirPath, glob string) (chan bool, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	defer watcher.Close()
	log.Println("Watching:", dirPath)
	filesChanged := make(chan bool, 1)
	go func() {
		for {
			select {
			case event := <-watcher.Events:

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					filesChanged <- true
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.Println("error:", err)
				}
			}
		}
	}()

	err = watcher.Add(dirPath)
	if err != nil {
		return nil, err
	}
	return filesChanged, nil
}
