package main

import (
	"log"

	"github.com/zanz1n/duvua/config"
	"github.com/zanz1n/duvua/internal/welcome"
	staticembed "github.com/zanz1n/duvua/static"
)

func welcomeImageGenerator() *welcome.ImageGenerator {
	cfg := config.GetConfig()

	template, err := welcome.LoadTemplate(staticembed.Assets, "welcomer.png")
	if err != nil {
		log.Fatalln("Failed to load welcomer image template:", err)
	}

	font, err := welcome.LoadFont(staticembed.Assets, "jetbrains-mono.ttf")
	if err != nil {
		log.Fatalln("Failed to load welcomer image font:", err)
	}

	return welcome.NewImageGenerator(
		template,
		font,
		cfg.Welcomer.ImageQuality,
	)
}
