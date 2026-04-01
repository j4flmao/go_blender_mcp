package main

import "os"

type Config struct {
	EnablePolyHaven bool
	EnableHyper3D   bool
	EnableSketchfab bool
	EnableHunyuan   bool
	Hyper3DKey      string
	SketchfabKey    string
	HunyuanKey      string
}

func LoadConfig() Config {
	return Config{
		EnablePolyHaven: os.Getenv("ENABLE_PH") == "1",
		EnableHyper3D:   os.Getenv("ENABLE_H3D") == "1",
		EnableSketchfab: os.Getenv("ENABLE_SF") == "1",
		EnableHunyuan:   os.Getenv("ENABLE_HY") == "1",
		Hyper3DKey:      os.Getenv("HYPER3D_KEY"),
		SketchfabKey:    os.Getenv("SKETCHFAB_KEY"),
		HunyuanKey:      os.Getenv("HUNYUAN_KEY"),
	}
}
