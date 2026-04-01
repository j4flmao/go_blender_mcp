package tools

import "fmt"

func (r *Registry) loadPolyHaven() {
	defs := []struct {
		name   string
		desc   string
		schema map[string]any
	}{
		{"import_polyhaven_hdri", "Download & apply Poly Haven HDRI.",
			withProps(
				"name", "string", "Asset slug e.g. autumn_field",
				"resolution", "string", "1k|2k|4k (default 1k)",
			)},
		{"import_polyhaven_texture", "Download Poly Haven texture (diff/rough/nor_gl/ao).",
			withProps(
				"name", "string", "Asset slug",
				"resolution", "string", "1k|2k|4k",
				"type", "string", "diff|rough|nor_gl|ao (default diff)",
			)},
		{"import_polyhaven_model", "Download & import Poly Haven 3D model (.blend).",
			withProps(
				"name", "string", "Asset slug",
				"resolution", "string", "1k|2k|4k (default 1k)",
			)},
	}

	for _, d := range defs {
		name := d.name
		r.add(Tool{Name: name, Description: d.desc, InputSchema: d.schema}, func(args map[string]any) (string, error) {
			s := r.runtime()
			if !s.EnablePolyHaven {
				return "", fmt.Errorf("Poly Haven disabled — enable it in Blender MCP panel (Poly Haven)")
			}
			res, err := send(name, args)
			if err != nil {
				return "", err
			}
			return fmtResult(name, res), nil
		})
	}
}
