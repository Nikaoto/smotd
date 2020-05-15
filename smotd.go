package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

const TimeFormat = "2006-01-02T15:04:05-07:00"
const ResetHour = 6
const SecondsInDay = 3600 * 24
const ShowInterval = 3600 * 2 // every 2 hours
const HelpText = `smotd - simple message of the day
Usage: smotd message_file .smotd_hist [-i interval]

Ideally, smotd should be run from a script which is executed regularly
(from .bashrc for example). It will print the contents of message_file to
stdout, write the current time and date to the history file (.smotd_hist in
this case) and exit.

Options:
	-i interval
        Makes smotd print the message every 'interval' seconds. For example,
        'smotd message.txt .smotd_history -i 7200' will make smotd check if 7200
        seconds (2 hours) passed since the last time the message was shown and
        if so will update .smotd_history and print message.txt to stdout.

	-h, --help
		Show this help text and exit.
`

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Message file not given\n")
		return
	}

	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Printf("%s", HelpText)
		return
	}

	if len(os.Args) < 3 {
		fmt.Printf("History file not given\n")
		return
	}

	// TODO use "flags" pkg to parse -i and reset interval in seconds
	//  right now, it's just ShowInterval (2 hours)
	intervalTrigger := false
	if len(os.Args) > 3 && os.Args[3] == "-i" {
		intervalTrigger = true
	}

	MsgFileName := os.Args[1]
	HistFileName := os.Args[2]

	histFileExists := true
	histFileInfo, err := os.Stat(HistFileName)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Printf("Error stating history file \"%s\"\n", HistFileName)
			return
		} else {
			histFileExists = false
		}
	}

	// Get last show time
	var lastShowDate time.Time
	if histFileExists {
		histFile, err := os.Open(HistFileName)
		if err != nil {
			fmt.Printf("Error opening history file \"%s\"\n", HistFileName)
			return
		}
		histBuf := make([]byte, histFileInfo.Size())
		_, _ = histFile.Read(histBuf)
		lastShowDate, _ = time.Parse(TimeFormat, strings.TrimSpace(string(histBuf)))
		histFile.Close()
	} else {
		lastShowDate = time.Unix(1, 0)
	}

	// Check if need to show
	now := time.Now()
	shouldShowMessage := false
	if intervalTrigger {
		// Show if interval passed
		if now.Unix()-lastShowDate.Unix() >= ShowInterval {
			shouldShowMessage = true
		}
	} else {
		// Show daily
		if now.Unix()-lastShowDate.Unix() >= SecondsInDay ||
			(lastShowDate.Hour() < ResetHour && now.Hour() >= ResetHour) {
			shouldShowMessage = true
		}
	}

	// Quit if don't need to show
	if !shouldShowMessage {
		return
	}

	// Show message
	msgFileInfo, err := os.Stat(MsgFileName)
	if err != nil {
		fmt.Printf("Error stating message file \"%s\"\n", MsgFileName)
		return
	}
	msgBuf := make([]byte, msgFileInfo.Size())
	msgFile, err := os.Open(MsgFileName)
	if err != nil {
		fmt.Printf("Error opening message file \"%s\"\n", MsgFileName)
		return
	}
	_, _ = msgFile.Read(msgBuf)
	msgFile.Close()
	fmt.Println(strings.TrimSpace(string(msgBuf)))

	// Update history
	histFile, err := os.OpenFile(HistFileName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		fmt.Printf("Couldn't update history file \"%s\", error while opening\n", MsgFileName)
		return
	}

	newDate := []byte(time.Now().Format(TimeFormat))
	newDate = append(newDate, '\n')
	_, err = histFile.Write(newDate)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Printf("Permission denied when updating history file \"%s\"\n", MsgFileName)
		} else {
			fmt.Printf("Coulnd't write to history file \"%s\"\n", MsgFileName)
		}
	}
	histFile.Close()
}
