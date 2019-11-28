package main

import (
	"github.com/zaker/anachrome-be/cmd"
)

func main() {
	cmd.Execute()
	// flag.Parse()
	// log.Println("Reading config from ", *confFile)
	// conf, err := config.Load(*confFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// s := services.NewSPA(conf.AppDir)
	// absPath, err := filepath.Abs(conf.AppDir)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// c := make(chan notify.EventInfo, 1)
	// if err := notify.Watch(absPath[:len(absPath)-5]+"/...", c, notify.Create|notify.Write); err != nil {
	// 	log.Fatal(err)
	// }
	// defer notify.Stop(c)
	// go func() {
	// 	for {
	// 		select {
	// 		case ei := <-c:
	// 			dirPath, fileName := filepath.Split(ei.Path())
	// 			basePath := filepath.Base(dirPath)

	// 			if basePath == "dist" && fileName == "index.html" {
	// 				log.Println("Hit")
	// 				go s.IndexParse()
	// 			}
	// 			//
	// 		}
	// 	}

	// }()

	// s.IndexParse()
	// // home route

}
