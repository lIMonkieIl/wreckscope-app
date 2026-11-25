package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NotesFor returns the GitHub release body for a version tag (with or without v).
func NotesFor(ctx context.Context, ver string) (string, error) {
	tag := strings.TrimSpace(ver)
	if tag == "" {
		return "", fmt.Errorf("empty version")
	}
	if !strings.HasPrefix(strings.ToLower(tag), "v") {
		tag = "v" + tag
	}
	url := "https://api.github.com/repos/" + githubRepo + "/releases/tags/" + tag
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "WreckScope")

	client := &http.Client{Timeout: 20 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github release %s: %s", tag, res.Status)
	}
	var payload struct {
		Body string `json:"body"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	return strings.TrimSpace(payload.Body), nil
}
