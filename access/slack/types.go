package main

import (
	"fmt"
	"strings"

	"github.com/gravitational/teleport-plugins/access"
)

type RequestData struct {
	User          string
	Roles         []string
	RequestReason string
}

type SlackDataMessage struct {
	ChannelID string
	Timestamp string
}

type SlackData = []SlackDataMessage

type PluginData struct {
	RequestData
	SlackData
}

func DecodePluginData(dataMap access.PluginDataMap) (data PluginData) {
	data.User = dataMap["user"]
	data.Roles = strings.Split(dataMap["roles"], ",")
	data.RequestReason = dataMap["request_reason"]
	if channelID, timestamp := dataMap["channel_id"], dataMap["timestamp"]; channelID != "" && timestamp != "" {
		data.SlackData = append(data.SlackData, SlackDataMessage{ChannelID: channelID, Timestamp: timestamp})
	}
	for _, encodedMsg := range strings.Split(dataMap["messages"], ",") {
		if parts := strings.Split(encodedMsg, "/"); len(parts) == 2 {
			data.SlackData = append(data.SlackData, SlackDataMessage{ChannelID: parts[0], Timestamp: parts[1]})
		}
	}
	return
}

func EncodePluginData(data PluginData) access.PluginDataMap {
	var encodedMessages []string
	for _, msg := range data.SlackData {
		encodedMessages = append(encodedMessages, fmt.Sprintf("%s/%s", msg.ChannelID, msg.Timestamp))
	}
	return access.PluginDataMap{
		"user":           data.User,
		"roles":          strings.Join(data.Roles, ","),
		"request_reason": data.RequestReason,
		"messages":       strings.Join(encodedMessages, ","),
	}
}
