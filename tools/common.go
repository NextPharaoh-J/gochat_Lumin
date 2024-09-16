package tools

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/bwmarrin/snowflake"
	"io"
)

const SessionPrefix = "sess_"

func GetSnowflakeId() string {
	//default node id eq 1,this can modify to different serverId node
	node, _ := snowflake.NewNode(1)
	// Generate a snowflake ID.
	id := node.Generate().String()
	return id
}

func GetRandomToken(length int) string {
	r := make([]byte, length)
	io.ReadFull(rand.Reader, r)
	return base64.URLEncoding.EncodeToString(r)
}

func CreateSeessionId(s string) string {
	return SessionPrefix + s
}

func GetSessionIdByUserId(userId int) string {
	return fmt.Sprintf("sess_map_%d", userId)
}

func GetSessionName(sessionId string) string {
	return SessionPrefix + sessionId
}
