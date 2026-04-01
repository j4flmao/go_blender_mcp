package tools

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func (r *Registry) loadSketchfab(apiKey string) {
	r.add(
		Tool{
			Name:        "sketchfab_search",
			Description: "Search Sketchfab for free downloadable 3D models.",
			InputSchema: withProps(
				"query", "string", "Search term e.g. 'viking rigged animated axe'",
				"limit", "integer", "Max results (default 5)",
				"downloadable", "boolean", "Only downloadable (default true)",
				"animated", "boolean", "Only animated (default true)",
				"rigged", "boolean", "Only rigged (default true)",
			),
		},
		func(args map[string]any) (string, error) {
			s := r.runtime()
			if !s.EnableSketchfab {
				return "", fmt.Errorf("Sketchfab disabled — enable it in Blender MCP panel (Sketchfab)")
			}
			key := apiKey
			if key == "" {
				key = s.SketchfabKey
			}
			if key == "" {
				return "", fmt.Errorf("SKETCHFAB_KEY not set — set it in Blender MCP panel or env")
			}
			q, _ := args["query"].(string)
			if q == "" {
				return "", fmt.Errorf("query required")
			}

			limit := 5
			if l, ok := args["limit"].(float64); ok {
				limit = int(l)
			}
			if limit <= 0 {
				limit = 5
			}

			downloadable := true
			if v, ok := args["downloadable"].(bool); ok {
				downloadable = v
			}
			animated := true
			if v, ok := args["animated"].(bool); ok {
				animated = v
			}
			rigged := true
			if v, ok := args["rigged"].(bool); ok {
				rigged = v
			}

			u, _ := url.Parse("https://api.sketchfab.com/v3/search")
			params := u.Query()
			params.Set("type", "models")
			params.Set("q", q)
			params.Set("count", fmt.Sprintf("%d", limit))
			if downloadable {
				params.Set("downloadable", "true")
			} else {
				params.Set("downloadable", "false")
			}
			if animated {
				params.Set("animated", "true")
			}
			if rigged {
				params.Set("rigged", "true")
			}
			u.RawQuery = params.Encode()

			req, _ := http.NewRequest("GET", u.String(), nil)
			req.Header.Set("Authorization", "Token "+key)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", fmt.Errorf("sketchfab: %w", err)
			}
			defer resp.Body.Close()

			raw, _ := io.ReadAll(resp.Body)
			var result map[string]any
			_ = json.Unmarshal(raw, &result)

			results, _ := result["results"].([]any)
			if len(results) == 0 {
				return "no results found", nil
			}

			out := fmt.Sprintf("%d results:\n", len(results))
			for _, item := range results {
				m, _ := item.(map[string]any)
				uid := m["uid"]
				name := m["name"]
				anim := m["animationCount"]
				rig := m["isRigged"]
				dl := m["isDownloadable"]
				out += fmt.Sprintf("  uid:%v name:%v anim:%v rigged:%v downloadable:%v\n", uid, name, anim, rig, dl)
			}
			return out, nil
		},
	)
}
