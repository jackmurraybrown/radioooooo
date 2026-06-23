package analytics

// ⋆˙⟡ icecast listener source — polls /admin/listclients for per-listener data
// needs admin credentials for the icecast admin API

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"
)

type IcecastSource struct {
	baseURL  string
	user     string
	password string
	mounts   []string
	client   *http.Client
}

type IcecastConfig struct {
	BaseURL  string
	User     string
	Password string
	Mounts   []string
}

func NewIcecastSource(cfg IcecastConfig) *IcecastSource {
	return &IcecastSource{
		baseURL:  cfg.BaseURL,
		user:     cfg.User,
		password: cfg.Password,
		mounts:   cfg.Mounts,
		client:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *IcecastSource) Poll(ctx context.Context) ([]ListenerSnapshot, error) {
	var all []ListenerSnapshot

	for _, mount := range s.mounts {
		listeners, err := s.pollMount(ctx, mount)
		if err != nil {
			return nil, fmt.Errorf("icecast mount %s: %w", mount, err)
		}
		all = append(all, listeners...)
	}

	return all, nil
}

func (s *IcecastSource) pollMount(ctx context.Context, mount string) ([]ListenerSnapshot, error) {
	url := fmt.Sprintf("%s/admin/listclients?mount=%s", s.baseURL, mount)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(s.user, s.password)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("poll failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var stats icecastListClients
	if err := xml.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}

	var snapshots []ListenerSnapshot
	for _, src := range stats.Sources {
		for _, l := range src.Listeners {
			snapshots = append(snapshots, ListenerSnapshot{
				Mount:     mount,
				IP:        l.IP,
				UserAgent: l.UserAgent,
			})
		}
	}

	return snapshots, nil
}

// . ݁₊ ✶ icecast /admin/listclients XML shape
type icecastListClients struct {
	XMLName xml.Name             `xml:"icestats"`
	Sources []icecastMountSource `xml:"source"`
}

type icecastMountSource struct {
	Mount     string            `xml:"mount,attr"`
	Listeners []icecastListener `xml:"listener"`
}

type icecastListener struct {
	IP        string `xml:"IP"`
	UserAgent string `xml:"UserAgent"`
	Connected int    `xml:"Connected"`
	ID        int    `xml:"ID"`
}
