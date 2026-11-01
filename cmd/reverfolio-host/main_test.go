package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"strings"
	"testing"
)

func fixture(t *testing.T, msg any) []byte {
	t.Helper()
	raw, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var framed bytes.Buffer
	if err = binary.Write(&framed, binary.LittleEndian, uint32(len(raw))); err != nil {
		t.Fatal(err)
	}
	framed.Write(raw)
	return framed.Bytes()
}
func decode(t *testing.T, b []byte) nativeReply {
	t.Helper()
	payload, err := readFrame(bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	var v nativeReply
	if err = json.Unmarshal(payload, &v); err != nil {
		t.Fatal(err)
	}
	return v
}
func TestNativeProtocolAndExplicitCommands(t *testing.T) {
	for _, tc := range []struct {
		cmd, state string
		ok         bool
	}{
		{"status", "stopped", true},
		{"ensure_bridge", "started", true},
		{"bad", "", false},
	} {
		var out bytes.Buffer
		called := 0
		err := serve(bytes.NewReader(fixture(t, nativeRequest{Command: tc.cmd})), &out, func(s string) nativeReply {
			called++
			if s != tc.cmd {
				t.Fatal("wrong command")
			}
			if s == "bad" {
				return nativeReply{Error: "unsupported command"}
			}
			return nativeReply{OK: true, State: tc.state}
		})
		if err != nil {
			t.Fatal(err)
		}
		if called != 1 {
			t.Fatal("request executed twice")
		}
		got := decode(t, out.Bytes())
		if got.OK != tc.ok || got.State != tc.state {
			t.Fatalf("%+v", got)
		}
	}
}
func TestNativeProtocolRejectsOversizedPayload(t *testing.T) {
	var payload bytes.Buffer
	_ = binary.Write(&payload, binary.LittleEndian, uint32(maxNativeMessage+1))
	called := false
	err := serve(bytes.NewReader(payload.Bytes()), &bytes.Buffer{}, func(string) nativeReply { called = true; return nativeReply{OK: true} })
	if err == nil || called {
		t.Fatalf("oversized request accepted: %v", err)
	}
}
func TestNativeInvalidJSONDoesNotInvokeStart(t *testing.T) {
	var out bytes.Buffer
	framed := fixture(t, "not an object")
	if err := serve(bytes.NewReader(framed), &out, func(string) nativeReply { t.Fatal("invoked"); return nativeReply{} }); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(decode(t, out.Bytes()).Error, "invalid") {
		t.Fatal("missing error response")
	}
}
