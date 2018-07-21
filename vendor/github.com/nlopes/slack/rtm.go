package slack

import (
	"context"
	"encoding/json"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	websocketDefaultTimeout = 10 * time.Second
)

// StartRTM calls the "rtm.start" endpoint and returns the provided URL and the full Info block.
//
// To have a fully managed Websocket connection, use `NewRTM`, and call `ManageConnection()` on it.
func (api *Client) StartRTM() (info *Info, websocketURL string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), websocketDefaultTimeout)
	defer cancel()

	return api.StartRTMContext(ctx)
}

// StartRTMContext calls the "rtm.start" endpoint and returns the provided URL and the full Info block with a custom context.
//
// To have a fully managed Websocket connection, use `NewRTM`, and call `ManageConnection()` on it.
func (api *Client) StartRTMContext(ctx context.Context) (info *Info, websocketURL string, err error) {
	response := &infoResponseFull{}
	err = post(ctx, api.httpclient, "rtm.start", url.Values{"token": {api.token}}, response, api.debug)
	if err != nil {
		return nil, "", err
	}

	api.Debugln("Using URL:", response.Info.URL)
	return &response.Info, response.Info.URL, response.Err()
}

// ConnectRTM calls the "rtm.connect" endpoint and returns the provided URL and the compact Info block.
//
// To have a fully managed Websocket connection, use `NewRTM`, and call `ManageConnection()` on it.
func (api *Client) ConnectRTM() (info *Info, websocketURL string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), websocketDefaultTimeout)
	defer cancel()

	return api.ConnectRTMContext(ctx)
}

// ConnectRTMContext calls the "rtm.connect" endpoint and returns the
// provided URL and the compact Info block with a custom context.
//
// To have a fully managed Websocket connection, use `NewRTM`, and call `ManageConnection()` on it.
func (api *Client) ConnectRTMContext(ctx context.Context) (info *Info, websocketURL string, err error) {
	response := &infoResponseFull{}
	err = post(ctx, api.httpclient, "rtm.connect", url.Values{"token": {api.token}}, response, api.debug)
	if err != nil {
		api.Debugf("Failed to connect to RTM: %s", err)
		return nil, "", err
	}

	api.Debugln("Using URL:", response.Info.URL)
	return &response.Info, response.Info.URL, response.Err()
}

// RTMOption options for the managed RTM.
type RTMOption func(*RTM)

// RTMOptionUseStart as of 11th July 2017 you should prefer setting this to false, see:
// https://api.slack.com/changelog/2017-04-start-using-rtm-connect-and-stop-using-rtm-start
func RTMOptionUseStart(b bool) RTMOption {
	return func(rtm *RTM) {
		rtm.useRTMStart = b
	}
}

// RTMOptionDialer takes a gorilla websocket Dialer and uses it as the
// Dialer when opening the websocket for the RTM connection.
func RTMOptionDialer(d *websocket.Dialer) RTMOption {
	return func(rtm *RTM) {
		rtm.dialer = d
	}
}

// NewRTM returns a RTM, which provides a fully managed connection to
// Slack's websocket-based Real-Time Messaging protocol.
func (api *Client) NewRTM(options ...RTMOption) *RTM {
	result := &RTM{
		Client:           *api,
		IncomingEvents:   make(chan RTMEvent, 50),
		outgoingMessages: make(chan OutgoingMessage, 20),
		pings:            make(map[int]time.Time),
		isConnected:      false,
		wasIntentional:   true,
		killChannel:      make(chan bool),
		disconnected:     make(chan struct{}),
		forcePing:        make(chan bool),
		rawEvents:        make(chan json.RawMessage),
		idGen:            NewSafeID(1),
		mu:               &sync.Mutex{},
	}

	for _, opt := range options {
		opt(result)
	}

	return result
}

// NewRTMWithOptions Deprecated just use NewRTM(RTMOptionsUseStart(true))
// returns a RTM, which provides a fully managed connection to
// Slack's websocket-based Real-Time Messaging protocol.
// This also allows to configure various options available for RTM API.
func (api *Client) NewRTMWithOptions(options *RTMOptions) *RTM {
	if options != nil {
		return api.NewRTM(RTMOptionUseStart(options.UseRTMStart))
	}
	return api.NewRTM()
}
