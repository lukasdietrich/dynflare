package systemd

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

func Install(configPath string) error {
	templates, err := template.ParseFS(templateFS, "templates/dynflare.*")
	if err != nil {
		return fmt.Errorf("could not read templates: %w", err)
	}

	ctx, err := newTemplateContext(configPath)
	if err != nil {
		return fmt.Errorf("could not create new template context: %w", err)
	}

	for _, name := range [...]string{"dynflare.service", "dynflare.timer"} {
		if err := writeTemplate(templates, ctx, name); err != nil {
			return err
		}
	}

	return nil
}

func writeTemplate(templates *template.Template, ctx *templateContext, name string) error {
	f, err := openUnitfile(name)
	if err != nil {
		return fmt.Errorf("could not open unit-file: %w", err)
	}

	defer f.Close()

	log.Printf("writing unit-file: %q", f.Name())
	return templates.ExecuteTemplate(f, name, ctx)
}

func openUnitfile(name string) (*os.File, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not lookup homedir: %w", err)
	}

	systemdFolder := filepath.Join(homedir, ".local/share/systemd/user")
	if err := os.MkdirAll(systemdFolder, 0700); err != nil {
		return nil, fmt.Errorf("could not create systemd folder: %w", err)
	}

	systemdFilename := filepath.Join(systemdFolder, name)
	return os.OpenFile(systemdFilename, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
}

type templateContext struct {
	Executable string
	Config     string
}

func newTemplateContext(configPath string) (*templateContext, error) {
	executablePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("could not determine path to executable: %w", err)
	}

	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("could not derive absolute path to config file: %w", err)
	}

	return &templateContext{
		Executable: executablePath,
		Config:     configPath,
	}, nil
}
