package app

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"strconv"
	"testing"
	"time"
)

//go:embed all:ui/dist
var lanTestAssets embed.FS

func TestLANHTTPInteractive(t *testing.T) {
	app := New(lanTestAssets)
	app.httpPort = 0 // ephemeral
	app.lan = true

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	app.httpPort = ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		app.startHTTP(ctx)
	}()

	base := ""
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		base = "http://127.0.0.1:" + strconv.Itoa(app.httpPort)
		resp, err := http.Get(base + "/api/health")
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == 200 && string(body) == "ok" {
				break
			}
		}
		base = ""
		time.Sleep(50 * time.Millisecond)
	}
	if base == "" {
		t.Fatal("LAN server did not become ready")
	}

	resp, err := http.Get(base + "/")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("SPA status %d", resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodPost, base+"/api/mode", bytes.NewBufferString(`{"mode":"demo"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", base)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if app.demoFeaturesEnabled() {
		if resp.StatusCode != http.StatusNoContent {
			t.Fatalf("mode status %d (want 204)", resp.StatusCode)
		}
		if app.engine.Mode() != "demo" {
			t.Fatalf("mode=%s", app.engine.Mode())
		}
	} else if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("mode status %d (want 400; demo off)", resp.StatusCode)
	}

	resp, err = http.Get(base + "/api/access")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var access map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&access); err != nil {
		t.Fatal(err)
	}
	if access["readOnly"] != false {
		t.Fatalf("readOnly=%v want false", access["readOnly"])
	}
	if access["listening"] != true {
		t.Fatalf("listening=%v want true", access["listening"])
	}

	resp, err = http.Get(base + "/api/settings")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		resp.Body.Close()
		t.Fatal(err)
	}
	resp.Body.Close()
	if settings["theme"] != "wreckscope" || settings["mode"] != "system" {
		t.Fatalf("default settings=%v", settings)
	}

	req, err = http.NewRequest(http.MethodPost, base+"/api/appearance", bytes.NewBufferString(`{"theme":"pine","mode":"light"}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", base)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var saved map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&saved); err != nil {
		resp.Body.Close()
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("appearance status %d", resp.StatusCode)
	}
	if saved["theme"] != "pine" || saved["mode"] != "light" {
		t.Fatalf("saved appearance=%v", saved)
	}
	got := app.GetSettings()
	if got.Theme != "pine" || got.Mode != "light" {
		t.Fatalf("engine appearance %+v", got)
	}

	req, err = http.NewRequest(http.MethodPost, base+"/api/medic", bytes.NewBufferString(`{"slot":4}`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", base)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	var medicResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&medicResp); err != nil {
		resp.Body.Close()
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("medic status %d", resp.StatusCode)
	}
	medics, ok := medicResp["medics"].([]interface{})
	if !ok || len(medics) != 1 || medics[0].(float64) != 4 {
		t.Fatalf("medic response=%v", medicResp)
	}
	if snap := app.Snapshot(); len(snap.Medics) != 1 || snap.Medics[0] != 4 {
		t.Fatalf("snapshot medics=%v", snap.Medics)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("server did not shut down")
	}
}

func TestPhoneAccessWithoutListen(t *testing.T) {
	app := New(lanTestAssets)
	app.lan = true
	app.setHTTPListen(false, "LAN dashboard port 8080 busy")
	access := app.PhoneAccess()
	if access["listening"] != false {
		t.Fatalf("listening=%v", access["listening"])
	}
	urls, _ := access["urls"].([]string)
	if len(urls) != 0 {
		t.Fatalf("urls should be empty when not listening: %v", urls)
	}
	if access["error"] != "LAN dashboard port 8080 busy" {
		t.Fatalf("error=%v", access["error"])
	}
}
