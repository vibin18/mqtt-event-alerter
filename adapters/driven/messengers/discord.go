package messengers

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"log/slog"
)

type DiscordOps struct {
	session        *discordgo.Session
	DiscordChannel string
	// DiscordData    *discordgo.MessageSend
}

func NewDiscordMessenger(token, channelId string) *DiscordOps {
	bot, err := discordgo.New("Bot " + token)
	if err != nil {
		slog.Error("discord server initialization error", slog.String("error", err.Error()))
	}

	return &DiscordOps{
		session:        bot,
		DiscordChannel: channelId,
	}
}

func (d *DiscordOps) SendPictureAlert(label, camera, eventId, contentTime string, picture io.ReadCloser) {

	var files []*discordgo.File

	df := &discordgo.File{
		Name:        fmt.Sprintf("%v.jpeg", eventId),
		ContentType: "image/jpeg",
		Reader:      picture,
	}

	files = append(files, df)

	ms := discordgo.MessageSend{
		Content: fmt.Sprintf("A %v detected on %v at %v", label, camera, contentTime),
		Files:   files,
	}

	_, err := d.session.ChannelMessageSendComplex(d.DiscordChannel, &ms)
	if err != nil {
		slog.Error("discord message send failure", slog.Any("error", err.Error()))
	}
	slog.Info("message send to discord messenger with snapshot")
}

func (d *DiscordOps) SendTextAlert(message string) {
	slog.Debug("sending text alert", slog.String("payload", message))
	_, err := d.session.ChannelMessageSend(d.DiscordChannel, message)
	if err != nil {
		slog.Error("discord message send failure", slog.Any("error", err.Error()))
	}
	slog.Info("text message send to discord messenger")
}
