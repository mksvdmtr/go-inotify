package main

import (
	"flag"
	"fmt"
	"github.com/ashwanthkumar/slack-go-webhook"
	inotify "gopkg.in/fsnotify.v0"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
	"time"
)

type Config struct {
	WatchFile string `yaml:"watch_file"`
	MattermostChannel string `yaml:"mattermost_channel"`
	MattermostToken string `yaml:"mattermost_token"`
	MattermostMentions string `yaml:"mattermost_mentions"`
}

var config Config

func (c *Config) notify(msg string) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(path.Join(userHomeDir, ".go-inotify.lock")); !os.IsNotExist(err) {
		log.Println("~/.go-inotify.lock exist")
		return
	}
	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("Error getting hostname: %s", err)
	}
	webhookUrl := "https://bfchat.ru/hooks/" + c.MattermostToken
	text := fmt.Sprintf("%s **Зафиксированы изменения в %s на %s.**\n", c.MattermostMentions, msg, host)

	payload := slack.Payload{
		Text:      text,
		Username:  "Inotifywait",
		Channel:   c.MattermostChannel,
		IconEmoji: ":warning:",
	}
	slackErr := slack.Send(webhookUrl, "", payload)

	if slackErr != nil {
		log.Printf("Error sending message: %s\n", slackErr)
	}
}


func main() {
	configFile := flag.String("c", "config.yml", "go-inotify -c config.yml")
	flag.Parse()
	config.MattermostMentions = ""
	configsFileData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		flag.Usage()
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(configsFileData, &config); err != nil {
		log.Fatal(err)
	}
	watcher, err := inotify.NewWatcher()
	if err != nil {
		config.notify(err.Error())
		log.Fatal(err)
	}

	go func () {
		for {
			if err := watcher.WatchFlags(config.WatchFile, inotify.FSN_MODIFY | inotify.FSN_DELETE | inotify.FSN_RENAME); err != nil {
				config.notify(err.Error())
				log.Println(err)
				if err := watcher.RemoveWatch(config.WatchFile); err != nil {
					log.Println(err)
				}
				time.Sleep(time.Second * 60)
			}
			time.Sleep(time.Second * 15)
		}
	}()

	for {
		select {
		case ev := <-watcher.Event:
			log.Println("event:", ev)
			config.notify(ev.String())
		case err := <-watcher.Error:
			log.Println("error:", err)
			config.notify(err.Error())
		}
	}

}
