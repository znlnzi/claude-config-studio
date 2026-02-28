package main

import (
	"context"
	"embed"

	"github.com/znlnzi/claude-config-studio/services"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()
	configService := services.NewConfigService()
	projectService := services.NewProjectService()
	mcpService := services.NewMCPService()
	hooksService := services.NewHooksService()
	templateService := services.NewTemplateService()
	extensionService := services.NewExtensionService()
	pluginService := services.NewPluginService()
	exportService := services.NewExportService()
	skillService := services.NewSkillService()

	err := wails.Run(&options.App{
		Title:     "ClaudeCode Config Studio",
		Width:     1100,
		Height:    720,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 245, G: 245, B: 247, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			configService.SetContext(ctx)
			projectService.SetContext(ctx)
			mcpService.SetContext(ctx)
			hooksService.SetContext(ctx)
			templateService.SetContext(ctx)
			extensionService.SetContext(ctx)
			pluginService.SetContext(ctx)
			exportService.SetContext(ctx)
			skillService.SetContext(ctx)
		},
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                 true,
				HideTitleBar:              false,
				FullSizeContent:           true,
				UseToolbar:                true,
				HideToolbarSeparator:      true,
			},
			Appearance: mac.NSAppearanceNameAqua,
			About: &mac.AboutInfo{
				Title:   "ClaudeCode Config Studio",
				Message: "Claude Code Configuration Management Tool v0.1.0",
			},
		},
		Bind: []interface{}{
			app,
			configService,
			projectService,
			mcpService,
			hooksService,
			templateService,
			extensionService,
			pluginService,
			exportService,
			skillService,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
