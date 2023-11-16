package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/h2non/filetype"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

const ApplicationName = "Telegram_qBittorrent_Notifier"

var ApplicationVersion = "0.0.0"

var Verbose bool

var magicWord string
var category string
var tags string

func main() {
	configDir, _ := os.UserConfigDir()
	defaultConfigPath := filepath.Join(configDir, ApplicationName, "config.yaml")

	rootFlags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "telegram-bot-token",
			Usage:       "Telegram `BOT_TOKEN` for sending notifications",
			DefaultText: "N/A",
			Category:    "Telegram",
		}),
		altsrc.NewInt64Flag(&cli.Int64Flag{
			Name:        "telegram-chat-id",
			Usage:       "Telegram `CHAT_ID` to receive notifications",
			DefaultText: "N/A",
			Category:    "Telegram",
		}),
		&cli.StringFlag{
			Name:  "config",
			Usage: "Load configuration from `FILE`",
			Value: defaultConfigPath,
		},
		&cli.BoolFlag{
			Name:        "verbose",
			Hidden:      true,
			Destination: &Verbose,
		},
	}

	sendFlags := []cli.Flag{
		altsrc.NewStringFlag(&cli.StringFlag{
			Name:        "magic-word",
			Usage:       "Custom `PREFIX` for --category/-l and --tags/-g options",
			Value:       "6д9",
			Destination: &magicWord,
		}),
		&cli.StringFlag{
			Name:     "torrent-name",
			Aliases:  []string{"n"},
			Usage:    "`%N`: Torrent name",
			Category: "Torrent Information",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "category",
			Aliases:  []string{"l"},
			Usage:    "`%L`: Category",
			Category: "Torrent Information",
			Action: func(c *cli.Context, v string) error {
				if !strings.HasPrefix(v, magicWord) {
					return fmt.Errorf("option --category/-l must start with %s", magicWord)
				}
				category = "#" + strings.TrimPrefix(v, magicWord)
				return nil
			},
		},
		&cli.StringFlag{
			Name:     "tags",
			Aliases:  []string{"g"},
			Usage:    "`%G`: Tags (separated by comma)",
			Category: "Torrent Information",
			Action: func(c *cli.Context, v string) error {
				if !strings.HasPrefix(v, magicWord) {
					return fmt.Errorf("option --tags/-g must start with %s", magicWord)
				}
				tags = "#" + strings.Join(strings.Split(strings.TrimPrefix(v, magicWord), ","), " #")
				return nil
			},
		},
		&cli.StringFlag{
			Name:     "content-path",
			Aliases:  []string{"f"},
			Usage:    "`%F`: Content path (same as root path for multifile torrent)",
			Category: "Torrent Information",
		},
		&cli.StringFlag{
			Name:     "root-path",
			Aliases:  []string{"r"},
			Usage:    "`%R`: Root path (first torrent subdirectory path)",
			Category: "Torrent Information",
		},
		&cli.StringFlag{
			Name:     "save-path",
			Aliases:  []string{"d"},
			Usage:    "`%D`: Save path",
			Category: "Torrent Information",
		},
		&cli.StringFlag{
			Name:     "number-of-files",
			Aliases:  []string{"c"},
			Usage:    "`%C`: Number of files",
			Category: "Torrent Information",
			Action: func(c *cli.Context, v string) error {
				if _, err := strconv.Atoi(v); err != nil {
					return fmt.Errorf("option --number-of-files/-c must be a number")
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:     "torrent-size",
			Aliases:  []string{"z"},
			Usage:    "`%Z`: Torrent size (bytes)",
			Category: "Torrent Information",
			Action: func(c *cli.Context, v string) error {
				if _, err := strconv.Atoi(v); err != nil {
					return fmt.Errorf("option --torrent-size/-z must be a number")
				}
				return nil
			},
		},
		&cli.StringFlag{
			Name:     "current-tracker",
			Aliases:  []string{"t"},
			Usage:    "`%T`: Current tracker",
			Category: "Torrent Information",
		},
		&cli.StringFlag{
			Name:     "info-hash-v1",
			Aliases:  []string{"i", "info-hash"},
			Usage:    "`%I`: Info hash v1",
			Category: "Torrent Information",
		},
		&cli.StringFlag{
			Name:     "info-hash-v2",
			Aliases:  []string{"j"},
			Usage:    "`%J`: Info hash v2",
			Category: "Torrent Information",
		},
		&cli.StringFlag{
			Name:     "torrent-id",
			Aliases:  []string{"k"},
			Usage:    "`%K`: Torrent ID",
			Category: "Torrent Information",
		},
		&cli.StringFlag{
			Name:  "thumbnail-source",
			Usage: "Generate a thumbnail from `FILE` (recommended to be the same as the `--content-path/-f` option)",
		},
	}

	app := &cli.App{
		Name:    ApplicationName,
		Usage:   "A simple CLI tool for qBittorrent that sends a notification to Telegram chat via bot on torrent finished",
		Version: ApplicationVersion,
		Commands: []*cli.Command{
			{
				Name:   "send",
				Usage:  "Send a notification with provided torrent information",
				Before: altsrc.InitInputSourceWithContext(sendFlags, altsrc.NewYamlSourceFromFlagFunc("config")),
				Action: sendNotification,
				Flags:  sendFlags,
			},
		},
		Flags:  rootFlags,
		Before: altsrc.InitInputSourceWithContext(rootFlags, altsrc.NewYamlSourceFromFlagFunc("config")),
	}

	app.Suggest = true

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func sendNotification(c *cli.Context) error {
	if !c.IsSet("telegram-bot-token") || !c.IsSet("telegram-chat-id") {
		return fmt.Errorf("global option --telegram-bot-token and --telegram-chat-id are required")
	}

	notification := fmt.Sprintf("[%s]\n\n🔔 Download Completed!\n\n", ApplicationName)

	groups := [][]struct {
		name  string
		value string
	}{
		{
			{"💿 Torrent Name", c.String("torrent-name")},
		},
		{
			{"📄 Number of Files", c.String("number-of-files")},
			{"📏 Torrent Size", humanizeBytes(c.String("torrent-size"))},
		},
		{
			{"📂 Content Path", c.String("content-path")},
			{"🏠 Root Path", c.String("root-path")},
			{"💾 Save Path", c.String("save-path")},
		},
		{
			{"🔍 Current Tracker", c.String("current-tracker")},
			{"🌐 Info Hash V1", c.String("info-hash-v1")},
			{"🌐 Info Hash V2", c.String("info-hash-v2")},
			{"🔑 Torrent ID", c.String("torrent-id")},
		},
		{
			{"📚 Category", category},
			{"🏷️ Tags", tags},
		},
	}

	for _, group := range groups {
		hasGroup := false
		groupInfo := ""
		for _, field := range group {
			if field.value != "" {
				groupInfo += fmt.Sprintf("%s: %s\n", field.name, field.value)
				hasGroup = true
			}
		}
		if hasGroup {
			notification += groupInfo + "\n"
		}
	}
	notification = strings.TrimSpace(notification)

	bot, err := tgbotapi.NewBotAPI(c.String("telegram-bot-token"))
	if err != nil {
		return err
	}

	if Verbose {
		bot.Debug = true
		log.Printf("Authorized on account @%s", bot.Self.UserName)
	}

	var message tgbotapi.Chattable
	message = tgbotapi.NewMessage(c.Int64("telegram-chat-id"), notification)

	thumbnailSource := c.String("thumbnail-source")
	if fileInfo, err := os.Stat(thumbnailSource); err == nil && !fileInfo.IsDir() {
		if thumbnail, err := generateThumbnail(thumbnailSource); err == nil {
			photo := tgbotapi.NewPhoto(c.Int64("telegram-chat-id"), tgbotapi.FileReader{Reader: thumbnail})
			photo.Caption = notification
			message = photo
		}
	}

	if _, err := bot.Send(message); err != nil {
		return err
	}

	return nil
}

func humanizeBytes(s string) string {
	rawBytes, err := humanize.ParseBytes(s)
	if err != nil {
		return ""
	}
	return humanize.Bytes(rawBytes)
}

func isVideoFile(fileName string) (bool, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return false, fmt.Errorf("error opening file: %w", err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			log.Printf("error closing file: %v", err)
		}
	}()

	header := make([]byte, 261)
	if _, err := file.Read(header); err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}

	return filetype.IsVideo(header), nil
}

func generateThumbnail(fileName string) (io.Reader, error) {
	isVideo, err := isVideoFile(fileName)
	if err != nil {
		return nil, err
	}
	if !isVideo {
		return nil, fmt.Errorf("file is not a video: %s", fileName)
	}

	buf := bytes.NewBuffer(nil)
	err = ffmpeg.Input(fileName).
		Filter("thumbnail", ffmpeg.Args{}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf).
		Silent(true).
		Run()
	if err != nil {
		return nil, fmt.Errorf("ffmpeg error generating thumbnail: %w", err)
	}

	return buf, nil
}
