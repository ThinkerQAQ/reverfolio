package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coder/websocket"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSessionAPIOnlyNoJavaScriptCookiesOrExistingTabs(t *testing.T) {
	var closed atomic.Int64
	var address string
	allowed := map[string]bool{"Target.createTarget": true, "Target.attachToTarget": true, "Network.enable": true, "Page.navigate": true, "Network.getResponseBody": true}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/json/version" {
			fmt.Fprintf(w, `{"webSocketDebuggerUrl":"%s/devtools/browser/fixture"}`, strings.Replace(address, "http:", "ws:", 1))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/json/close/") {
			if r.URL.Path != "/json/close/new-target" {
				t.Error("closed an unrelated target")
			}
			closed.Add(1)
			fmt.Fprint(w, "ok")
			return
		}
		ws, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer ws.CloseNow()
		for {
			_, data, err := ws.Read(r.Context())
			if err != nil {
				return
			}
			var call struct {
				ID        int            `json:"id"`
				Method    string         `json:"method"`
				SessionID string         `json:"sessionId"`
				Params    map[string]any `json:"params"`
			}
			if json.Unmarshal(data, &call) != nil {
				return
			}
			if !allowed[call.Method] {
				t.Error("forbidden browser command", call.Method)
				return
			}
			result := map[string]any{}
			switch call.Method {
			case "Target.createTarget":
				if call.Params["url"] != "about:blank" || call.Params["background"] != true {
					t.Error("must create only a background target")
				}
				result["targetId"] = "new-target"
			case "Target.attachToTarget":
				if call.Params["targetId"] != "new-target" {
					t.Error("must not attach to user targets")
				}
				result["sessionId"] = "new-session"
			case "Page.navigate":
				endpoint := fmt.Sprint(call.Params["url"])
				if !strings.HasPrefix(endpoint, "https://cloud.umami.is/api/websites/website-id/stats?") {
					t.Error("wrong API destination", endpoint)
				}
				for _, event := range []any{map[string]any{"method": "Network.responseReceived", "sessionId": "new-session", "params": map[string]any{"requestId": "r1", "response": map[string]any{"url": endpoint, "status": 200}}}, map[string]any{"method": "Network.loadingFinished", "sessionId": "new-session", "params": map[string]any{"requestId": "r1"}}} {
					b, _ := json.Marshal(event)
					ws.Write(r.Context(), websocket.MessageText, b)
				}
			case "Network.getResponseBody":
				result["body"] = `{"pageviews":12,"visitors":3,"visits":7}`
				result["base64Encoded"] = false
			}
			b, _ := json.Marshal(map[string]any{"id": call.ID, "result": result})
			if ws.Write(r.Context(), websocket.MessageText, b) != nil {
				return
			}
		}
	}))
	defer server.Close()
	address = server.URL
	source := SiteSource{ID: "umami", Provider: "umami", WorkID: 1, AuthMode: "browser", BrowserEndpoint: server.URL, Endpoint: "https://cloud.umami.is/api", Property: "website-id", Timezone: "UTC", LookbackDays: 7, LagDays: 1}
	metrics, problem := (ProviderClient{}).collect(context.Background(), source, "", time.Now())
	if problem != "" || len(metrics) != 3 {
		t.Fatalf("metrics %+v problem %s", metrics, problem)
	}
	if closed.Load() != 1 {
		t.Fatal("temporary browser target leaked")
	}
	if validBrowserEndpoint("http://192.168.1.1:9222") || validBrowserEndpoint("https://127.0.0.1:9222") || validBrowserEndpoint("http://user:pass@127.0.0.1:9222") {
		t.Fatal("unsafe browser endpoint accepted")
	}
}
