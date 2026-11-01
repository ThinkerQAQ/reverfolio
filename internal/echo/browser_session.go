package echo

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
)

// This client only navigates a fresh background target to a fixed, readonly
// Umami statistics API. It never evaluates JavaScript, reads cookies/storage,
// exports login credentials or touches an existing user tab.
func validBrowserEndpoint(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Path != "" {
		return false
	}
	ip := net.ParseIP(u.Hostname())
	return (u.Hostname() == "localhost" || ip != nil && ip.IsLoopback()) && u.Port() != ""
}

type cdpMessage struct {
	ID        int             `json:"id"`
	Method    string          `json:"method"`
	SessionID string          `json:"sessionId,omitempty"`
	Params    json.RawMessage `json:"params"`
	Result    json.RawMessage `json:"result"`
	Error     json.RawMessage `json:"error"`
}
type cdpConnection struct {
	conn   *websocket.Conn
	serial int
	events []cdpMessage
}

func (c *cdpConnection) read(ctx context.Context) (cdpMessage, error) {
	_, data, err := c.conn.Read(ctx)
	if err != nil {
		return cdpMessage{}, err
	}
	var m cdpMessage
	err = json.Unmarshal(data, &m)
	return m, err
}
func (c *cdpConnection) call(ctx context.Context, method, session string, params any, out any) error {
	c.serial++
	id := c.serial
	b, _ := json.Marshal(map[string]any{"id": id, "method": method, "sessionId": session, "params": params})
	if err := c.conn.Write(ctx, websocket.MessageText, b); err != nil {
		return err
	}
	for {
		m, err := c.read(ctx)
		if err != nil {
			return err
		}
		if m.ID != id {
			if m.Method != "" {
				c.events = append(c.events, m)
			}
			continue
		}
		if len(m.Error) > 0 {
			return fmt.Errorf("browser command rejected")
		}
		if out != nil {
			return json.Unmarshal(m.Result, out)
		}
		return nil
	}
}
func (p ProviderClient) browserRequest(ctx context.Context, c SiteSource, endpoint string, out any) string {
	if c.Provider != "umami" || c.Endpoint != "https://cloud.umami.is/api" || c.AuthMode != "browser" {
		return "browser_source_not_allowed"
	}
	expected, _ := url.Parse(endpoint)
	if expected == nil || expected.Scheme != "https" || expected.Host != "cloud.umami.is" || expected.Path != "/api/websites/"+c.Property+"/stats" {
		return "browser_api_not_allowed"
	}
	debug := c.BrowserEndpoint
	if debug == "" {
		debug = "http://127.0.0.1:9222"
	}
	if !validBrowserEndpoint(debug) {
		return "browser_endpoint_invalid"
	}
	child, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var version struct {
		WebSocket string `json:"webSocketDebuggerUrl"`
	}
	if problem := p.request(child, "GET", debug+"/json/version", nil, nil, &version); problem != "" {
		return "browser_not_running"
	}
	ws, err := url.Parse(version.WebSocket)
	base, _ := url.Parse(debug)
	if err != nil || ws.Scheme != "ws" || ws.Host != base.Host || !strings.HasPrefix(ws.Path, "/devtools/browser/") {
		return "browser_endpoint_invalid"
	}
	conn, _, err := websocket.Dial(child, version.WebSocket, &websocket.DialOptions{HTTPClient: p.httpClient()})
	if err != nil {
		return "browser_connection_failed"
	}
	defer conn.CloseNow()
	conn.SetReadLimit(8 * 1024 * 1024)
	client := cdpConnection{conn: conn}
	var target struct {
		ID string `json:"targetId"`
	}
	if err = client.call(child, "Target.createTarget", "", map[string]any{"url": "about:blank", "background": true}, &target); err != nil || target.ID == "" {
		return "browser_target_failed"
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(cleanup, "GET", debug+"/json/close/"+url.PathEscape(target.ID), nil)
		if res, err := p.httpClient().Do(req); err == nil {
			res.Body.Close()
		}
	}()
	var attached struct {
		ID string `json:"sessionId"`
	}
	if client.call(child, "Target.attachToTarget", "", map[string]any{"targetId": target.ID, "flatten": true}, &attached) != nil {
		return "browser_target_failed"
	}
	if client.call(child, "Network.enable", attached.ID, map[string]any{}, nil) != nil {
		return "browser_network_unavailable"
	}
	if client.call(child, "Page.navigate", attached.ID, map[string]any{"url": endpoint}, nil) != nil {
		return "browser_navigation_failed"
	}
	var requestID string
	status := 0
	finished := map[string]bool{}
	for {
		var message cdpMessage
		if len(client.events) > 0 {
			message = client.events[0]
			client.events = client.events[1:]
		} else {
			message, err = client.read(child)
			if err != nil {
				return "browser_session_unavailable"
			}
		}
		if message.SessionID != attached.ID {
			continue
		}
		switch message.Method {
		case "Network.responseReceived":
			var event struct {
				RequestID string `json:"requestId"`
				Response  struct {
					URL    string `json:"url"`
					Status int    `json:"status"`
				} `json:"response"`
			}
			if json.Unmarshal(message.Params, &event) == nil && event.Response.URL == endpoint {
				requestID = event.RequestID
				status = event.Response.Status
			}
		case "Network.loadingFinished":
			var event struct {
				RequestID string `json:"requestId"`
			}
			if json.Unmarshal(message.Params, &event) == nil {
				finished[event.RequestID] = true
			}
		}
		if requestID == "" || !finished[requestID] {
			continue
		}
		if status != 200 {
			return fmt.Sprintf("provider_http_%d", status)
		}
		var response struct {
			Body   string `json:"body"`
			Base64 bool   `json:"base64Encoded"`
		}
		if client.call(child, "Network.getResponseBody", attached.ID, map[string]any{"requestId": requestID}, &response) != nil {
			return "browser_response_unavailable"
		}
		data := []byte(response.Body)
		if response.Base64 {
			data, err = base64.StdEncoding.DecodeString(response.Body)
			if err != nil {
				return "response_invalid"
			}
		}
		if len(data) > 4*1024*1024 {
			return "response_too_large"
		}
		if json.Unmarshal(data, out) != nil {
			return "browser_session_expired_or_response_invalid"
		}
		return ""
	}
}
