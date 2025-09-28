package ss13_se

import (
	"crypto/sha256"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	byondURL  string = "http://www.byond.com/games/Exadv1/SpaceStation13?format=text"
	userAgent string = "ss13hub/2.0pre"
)

func scrapeByond(webClient *http.Client, now time.Time) ([]ServerEntry, error) {
	body, err := openPage(webClient, byondURL)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	servers, err := parseByondText(now, body)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func openPage(webClient *http.Client, url string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", userAgent)

	resp, err := webClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("bad http.Response.Status: %s", resp.Status)
	}
	return resp.Body, nil
}

func parseByondText(now time.Time, body io.Reader) ([]ServerEntry, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	var servers []ServerEntry
	var cur ServerEntry

	reTitle := regexp.MustCompile(`<b>(.*?)</b>`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "world/") {
			if cur.ID != "" {
				servers = append(servers, cur)
			}
			cur = ServerEntry{}
			cur.Time = now
			continue
		}

		if strings.HasPrefix(line, "url =") {
			cur.GameURL = strings.Trim(strings.TrimPrefix(line, "url = "), `"`)
		}

		if strings.HasPrefix(line, "status =") {
			raw := strings.Trim(strings.TrimPrefix(line, "status = "), `"`)
			m := reTitle.FindStringSubmatch(raw)
			if len(m) > 1 {
				cur.Title = html.UnescapeString(m[1])
				cur.ID = makeID(cur.Title)
			}
		}

		if strings.HasPrefix(line, "players =") {
			inside := strings.TrimPrefix(line, "players = list(")
			inside = strings.TrimSuffix(inside, ")")
			inside = strings.TrimSpace(inside)
			if inside == "" {
				cur.Players = 0
			} else {
				cur.Players = len(strings.Split(inside, ","))
			}
		}
	}

	if cur.ID != "" {
		servers = append(servers, cur)
	}

	return servers, nil
}

func makeID(title string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(title)))
}
