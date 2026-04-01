package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"

	"blender-mcp-go/tools"
)

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func errResp(id any, code int, msg string) Response {
	return Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}

func serve(r io.Reader, w io.Writer, cfg Config) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	enc := json.NewEncoder(w)

	reg := tools.NewRegistry(
		cfg.EnablePolyHaven,
		cfg.EnableHyper3D,
		cfg.EnableSketchfab,
		cfg.EnableHunyuan,
		cfg.Hyper3DKey,
		cfg.SketchfabKey,
		cfg.HunyuanKey,
	)

	for sc.Scan() {
		var req Request
		if err := json.Unmarshal(sc.Bytes(), &req); err != nil {
			_ = enc.Encode(errResp(nil, -32700, "parse error"))
			continue
		}

		resp := Response{JSONRPC: "2.0", ID: req.ID}

		switch req.Method {
		case "initialize":
			resp.Result = map[string]any{
				"protocolVersion": "2024-11-05",
				"serverInfo":      map[string]any{"name": "blender-mcp-go", "version": "2.0.0"},
				"capabilities":    map[string]any{"tools": map[string]any{}},
			}
		case "tools/list":
			resp.Result = map[string]any{"tools": reg.List()}
		case "tools/call":
			var p struct {
				Name      string         `json:"name"`
				Arguments map[string]any `json:"arguments"`
			}
			if err := json.Unmarshal(req.Params, &p); err != nil {
				resp = errResp(req.ID, -32602, "invalid params")
				break
			}
			text, err := reg.Call(p.Name, p.Arguments)
			if err != nil {
				resp = errResp(req.ID, -32000, err.Error())
				break
			}
			resp.Result = map[string]any{
				"content": []map[string]any{{"type": "text", "text": text}},
			}
		case "notifications/initialized":
			continue
		default:
			resp = errResp(req.ID, -32601, fmt.Sprintf("unknown: %s", req.Method))
		}

		_ = enc.Encode(resp)
	}
}
