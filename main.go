package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mqu/go-notify"
)

type notification struct {
	Title, DescriptionHeader, LineFormat, ActionName string
}

type source struct {
	Name, CheckUpdatesCommand, UpdateCommand string

	// state members
	HasUpdates bool
}

type config struct {
	Notification notification
	Source       []source
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "error: too few arguments")
		os.Exit(1)
	}

	var config config
	_, err := toml.DecodeFile(os.Args[1], &config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: could not parse config file")
		os.Exit(1)
	}

	if !notify.Init(os.Args[0]) {
		fmt.Fprintln(os.Stderr, "error: could not init libnotify context")
		os.Exit(1)
	}

	hasUpdates := false
	var notificationDescription strings.Builder
	notificationDescription.WriteString(config.Notification.DescriptionHeader + "\n")
	for _, source := range config.Source {
		err := exec.Command("sh", "-c", source.CheckUpdatesCommand).Run()
		if err == nil {
			hasUpdates = true
			source.HasUpdates = true
			notificationDescription.WriteString(fmt.Sprintf(config.Notification.LineFormat+"\n", source.Name))
		}
	}

	if hasUpdates {
		update := func(*notify.NotifyNotification, string, interface{}) {
			for _, source := range config.Source {
				exec.Command("sh", "-c", source.UpdateCommand).Run()
			}
		}

		notification := notify.NotificationNew(config.Notification.Title, notificationDescription.String(), "")
		notification.AddAction("action_click", config.Notification.ActionName, update, nil)
		notification.SetUrgency(notify.NOTIFY_URGENCY_CRITICAL)
		notification.Show()
		notify.UnInit()
	}
}
