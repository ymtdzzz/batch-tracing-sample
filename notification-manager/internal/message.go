package internal

import (
	"bytes"
	"encoding/gob"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
)

const (
	NotiTypeEmail = "email"
	NotiTypePush  = "push"
)

type NotificationMessage struct {
	NotiType string
	ID       string // email: email address, push: device id
}

func (n *NotificationMessage) Encode() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	err := gob.NewEncoder(buf).Encode(n)
	return buf.Bytes(), err
}

func DecodeNotificationMessage(data []byte) (*NotificationMessage, error) {
	var msg NotificationMessage
	buf := bytes.NewBuffer(data)
	err := gob.NewDecoder(buf).Decode(&msg)
	return &msg, err
}

func NewRandomNotification() *NotificationMessage {
	rand.Seed(time.Now().UnixNano())
	random := rand.Intn(100)
	isEmail := random%2 == 0

	if isEmail {
		return &NotificationMessage{
			NotiType: NotiTypeEmail,
			ID:       faker.Email(),
		}
	}
	return &NotificationMessage{
		NotiType: NotiTypePush,
		ID:       uuid.New().String(),
	}
}
