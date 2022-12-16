package internal

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	baseURL   = "http://server:8080/"
	pathEmail = "email"
	pathPush  = "push"
)

func CallServer(ctx context.Context, client *http.Client, msg *NotificationMessage) error {
	var url *url.URL
	var err error
	if msg.NotiType == NotiTypeEmail {
		url, err = url.Parse(baseURL + pathEmail)
	} else if msg.NotiType == NotiTypePush {
		url, err = url.Parse(baseURL + pathPush)
	} else {
		err = errors.New(fmt.Sprintf("unsupported notification type: %s", msg.NotiType))
	}
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	_, err = io.ReadAll(res.Body)
	_ = res.Body.Close()

	return err
}
