package tools

import "fmt"

func (r *Registry) loadHunyuan(apiKey string) {
	r.add(
		Tool{
			Name:        "hunyuan_generate",
			Description: "Generate 3D model from image using Tencent Hunyuan. Needs HUNYUAN_KEY.",
			InputSchema: withProps(
				"image_path", "string", "Local path to input image",
				"output_path", "string", "Where to save output .glb",
			),
		},
		func(args map[string]any) (string, error) {
			s := r.runtime()
			if !s.EnableHunyuan {
				return "", fmt.Errorf("Hunyuan disabled — enable it in Blender MCP panel (Hunyuan)")
			}
			key := apiKey
			if key == "" {
				key = s.HunyuanKey
			}
			if key == "" {
				return "", fmt.Errorf("HUNYUAN_KEY not set — set it in Blender MCP panel or env")
			}
			imgPath, _ := args["image_path"].(string)
			if imgPath == "" {
				return "", fmt.Errorf("image_path required")
			}
			return fmt.Sprintf("hunyuan: submit %s to Hunyuan API (not implemented)", imgPath), nil
		},
	)
}
