package tools

func (r *Registry) loadCore() {
	defs := []struct {
		name   string
		desc   string
		schema map[string]any
	}{
		{"get_scene_info", "Scene name, frame, objects, render engine.", noArgs()},
		{"list_objects", "List objects (name+type). opt: limit int.",
			withProps("limit", "integer", "Max to return (default 50)")},
		{"get_object", "Location, rotation, scale of one object.",
			required("name", "string", "Object name")},
		{"create_object", "Add mesh/light/camera. type:CUBE|SPHERE|CYLINDER|PLANE|CONE|TORUS|EMPTY|CAMERA|LIGHT.",
			withProps(
				"type", "string", "CUBE|SPHERE|CYLINDER|PLANE|CONE|TORUS|EMPTY|CAMERA|LIGHT",
				"name", "string", "Object name",
				"location", "array", "[x,y,z]",
			)},
		{"move_object", "Move/rotate/scale. all optional per-axis.",
			withProps(
				"name", "string", "Object name",
				"location", "array", "[x,y,z]",
				"rotation", "array", "[rx,ry,rz] radians",
				"scale", "array", "[sx,sy,sz]",
			)},
		{"delete_object", "Delete by name.",
			required("name", "string", "Object name")},
		{"set_material", "Principled BSDF: color[RGBA 0-1], metallic, roughness.",
			withProps(
				"name", "string", "Object name",
				"color", "array", "[R,G,B,A] 0-1",
				"metallic", "number", "0-1",
				"roughness", "number", "0-1",
				"material_name", "string", "Material name",
			)},
		{"render_scene", "Render to file. Returns path.",
			withProps(
				"output_path", "string", "e.g. /tmp/out.png",
				"width", "integer", "px",
				"height", "integer", "px",
				"samples", "integer", "Cycles samples",
			)},
		{"set_render_engine", "CYCLES|BLENDER_EEVEE|BLENDER_WORKBENCH.",
			required("engine", "string", "CYCLES|BLENDER_EEVEE|BLENDER_WORKBENCH")},
		{"exec_python", "Run Python in Blender. Set result= for output.",
			required("code", "string", "Python code")},
	}

	for _, d := range defs {
		name := d.name
		r.add(Tool{Name: name, Description: d.desc, InputSchema: d.schema}, func(args map[string]any) (string, error) {
			res, err := send(name, args)
			if err != nil {
				return "", err
			}
			return fmtResult(name, res), nil
		})
	}
}
