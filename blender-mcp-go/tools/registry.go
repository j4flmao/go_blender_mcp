package tools

import (
	"fmt"
	"time"

	"blender-mcp-go/blender"
)

var bc = blender.New()

type RuntimeSettings struct {
	EnablePolyHaven bool
	EnableHyper3D   bool
	EnableSketchfab bool
	EnableHunyuan   bool
	Hyper3DKey      string
	SketchfabKey    string
	HunyuanKey      string
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type Registry struct {
	tools     []Tool
	handlers  map[string]func(map[string]any) (string, error)
	fallback  RuntimeSettings
	listAll   bool
	lastProbe time.Time
	cached    RuntimeSettings
}

func NewRegistry(ph, h3d, sf, hy bool, h3dKey, sfKey, hyKey string) *Registry {
	r := &Registry{
		handlers: map[string]func(map[string]any) (string, error){},
		fallback: RuntimeSettings{
			EnablePolyHaven: ph,
			EnableHyper3D:   h3d,
			EnableSketchfab: sf,
			EnableHunyuan:   hy,
			Hyper3DKey:      h3dKey,
			SketchfabKey:    sfKey,
			HunyuanKey:      hyKey,
		},
		listAll: true,
	}
	r.loadCore()
	r.loadPolyHaven()
	r.loadHyper3D(h3dKey)
	r.loadSketchfab(sfKey)
	r.loadHunyuan(hyKey)
	return r
}

func (r *Registry) add(t Tool, fn func(map[string]any) (string, error)) {
	r.tools = append(r.tools, t)
	r.handlers[t.Name] = fn
}

func (r *Registry) List() []Tool {
	return r.tools
}

func (r *Registry) Call(name string, args map[string]any) (string, error) {
	fn, ok := r.handlers[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return fn(args)
}

func (r *Registry) runtime() RuntimeSettings {
	now := time.Now()
	if !r.lastProbe.IsZero() && now.Sub(r.lastProbe) < 250*time.Millisecond {
		return r.cached
	}

	out := r.fallback

	res, err := bc.SendWithTimeout(250*time.Millisecond, "get_settings", map[string]any{})
	if err == nil {
		if v, ok := res["enable_polyhaven"].(bool); ok {
			out.EnablePolyHaven = out.EnablePolyHaven || v
		}
		if v, ok := res["enable_hyper3d"].(bool); ok {
			out.EnableHyper3D = out.EnableHyper3D || v
		}
		if v, ok := res["enable_sketchfab"].(bool); ok {
			out.EnableSketchfab = out.EnableSketchfab || v
		}
		if v, ok := res["enable_hunyuan"].(bool); ok {
			out.EnableHunyuan = out.EnableHunyuan || v
		}

		if out.Hyper3DKey == "" {
			if v, ok := res["hyper3d_key"].(string); ok {
				out.Hyper3DKey = v
			}
		}
		if out.SketchfabKey == "" {
			if v, ok := res["sketchfab_key"].(string); ok {
				out.SketchfabKey = v
			}
		}
		if out.HunyuanKey == "" {
			if v, ok := res["hunyuan_key"].(string); ok {
				out.HunyuanKey = v
			}
		}
	}

	r.lastProbe = now
	r.cached = out
	return out
}

func noArgs() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func withProps(kv ...string) map[string]any {
	p := map[string]any{}
	for i := 0; i+2 < len(kv); i += 3 {
		p[kv[i]] = map[string]any{
			"type":        kv[i+1],
			"description": kv[i+2],
		}
	}
	return map[string]any{"type": "object", "properties": p}
}

func required(key, typ, desc string) map[string]any {
	return map[string]any{
		"type":     "object",
		"required": []string{key},
		"properties": map[string]any{
			key: map[string]any{"type": typ, "description": desc},
		},
	}
}

func send(cmd string, args map[string]any) (map[string]any, error) {
	return bc.Send(cmd, args)
}

func fmtResult(name string, r map[string]any) string {
	switch name {
	case "get_scene_info":
		return fmt.Sprintf("scene:%v frame:%v objects:%v engine:%v", r["name"], r["frame"], r["objects"], r["engine"])
	case "list_objects":
		objs, _ := r["objects"].([]any)
		if len(objs) == 0 {
			return "no objects"
		}
		s := fmt.Sprintf("%d objects:\n", len(objs))
		for _, o := range objs {
			m, _ := o.(map[string]any)
			s += fmt.Sprintf("  %v (%v)\n", m["name"], m["type"])
		}
		return s
	case "get_object":
		return fmt.Sprintf("%v | loc:%v rot:%v scale:%v", r["name"], r["location"], r["rotation"], r["scale"])
	case "create_object":
		return fmt.Sprintf("created:%v at %v", r["created"], r["location"])
	case "move_object":
		return fmt.Sprintf("updated:%v loc:%v", r["updated"], r["location"])
	case "delete_object":
		return fmt.Sprintf("deleted:%v", r["deleted"])
	case "set_material":
		return fmt.Sprintf("material '%v' → %v", r["material"], r["object"])
	case "render_scene":
		return fmt.Sprintf("rendered %v → %v", r["size"], r["rendered"])
	case "set_render_engine":
		return fmt.Sprintf("engine: %v", r["engine"])
	case "exec_python":
		return fmt.Sprintf("result: %v", r["result"])
	case "import_polyhaven_hdri":
		return fmt.Sprintf("hdri:%v (%v) → %v", r["hdri"], r["resolution"], r["path"])
	case "import_polyhaven_texture":
		return fmt.Sprintf("texture:%v type:%v → %v", r["texture"], r["type"], r["path"])
	case "import_polyhaven_model":
		return fmt.Sprintf("model:%v objects:%v", r["model"], r["objects"])
	default:
		return fmt.Sprintf("%v", r)
	}
}
