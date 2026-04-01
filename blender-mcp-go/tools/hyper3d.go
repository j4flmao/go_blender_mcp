package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (r *Registry) loadHyper3D(apiKey string) {
	r.add(
		Tool{
			Name:        "hyper3d_generate",
			Description: "Generate 3D model from text prompt using Rodin AI. Needs HYPER3D_KEY.",
			InputSchema: withProps(
				"prompt", "string", "Text description of the 3D model",
				"output_path", "string", "Where to save .glb e.g. /tmp/model.glb",
			),
		},
		func(args map[string]any) (string, error) {
			s := r.runtime()
			if !s.EnableHyper3D {
				return "", fmt.Errorf("Hyper3D disabled — enable it in Blender MCP panel (Hyper3D Rodin)")
			}
			key := apiKey
			if key == "" {
				key = s.Hyper3DKey
			}
			if key == "" {
				return "", fmt.Errorf("HYPER3D_KEY not set — set it in Blender MCP panel or env")
			}
			prompt, _ := args["prompt"].(string)
			if prompt == "" {
				return "", fmt.Errorf("prompt required")
			}

			outPath, _ := args["output_path"].(string)
			if outPath == "" {
				outPath = "/tmp/hyper3d_model.glb"
			}

			body, _ := json.Marshal(map[string]any{
				"prompt": prompt,
				"tier":   "Regular",
			})

			req, _ := http.NewRequest("POST", "https://hyperhuman.deemos.com/api/v2/rodin", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+key)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", fmt.Errorf("hyper3d request: %w", err)
			}
			defer resp.Body.Close()

			raw, _ := io.ReadAll(resp.Body)
			var result map[string]any
			_ = json.Unmarshal(raw, &result)

			jobID, _ := result["job_id"].(string)
			if jobID == "" {
				s := string(raw)
				if len(s) > 200 {
					s = s[:200]
				}
				return "", fmt.Errorf("hyper3d: no job_id in response: %s", s)
			}

			return fmt.Sprintf("job submitted: %s — poll status then import .glb to: %s", jobID, outPath), nil
		},
	)
}
