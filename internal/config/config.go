// Package config validates raw user input and derives a complete ProjectConfig.
package config

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Input is the raw, unvalidated data collected from the user.
type Input struct {
	ProjectName   string
	Description   string
	AuthorName    string
	AuthorEmail   string
	PythonVersion string
	OutputDir     string
}

// ProjectConfig is the validated configuration with derived fields filled in.
type ProjectConfig struct {
	ProjectName      string
	PackageName      string
	Description      string
	AuthorName       string
	AuthorEmail      string
	PythonVersion    string
	PythonVersionTag string // e.g. "312" for 3.12
	OutputDir        string
}

var (
	validName    = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)
	validVersion = regexp.MustCompile(`^(\d+)\.(\d+)(?:\.\d+)?$`)
)

// ValidateName reports whether s is an acceptable project name.
func ValidateName(s string) error {
	if s == "" {
		return errors.New("project name is required")
	}
	if !validName.MatchString(s) {
		return errors.New("must be lowercase and start with a letter (a-z, 0-9, - and _ allowed)")
	}
	return nil
}

// ValidatePythonVersion accepts MAJOR.MINOR or MAJOR.MINOR.PATCH.
func ValidatePythonVersion(s string) error {
	if !validVersion.MatchString(s) {
		return errors.New("python version must be MAJOR.MINOR or MAJOR.MINOR.PATCH (e.g. 3.12)")
	}
	return nil
}

// New validates in and returns a complete ProjectConfig.
func New(in Input) (ProjectConfig, error) {
	if err := ValidateName(in.ProjectName); err != nil {
		return ProjectConfig{}, err
	}
	if err := ValidatePythonVersion(in.PythonVersion); err != nil {
		return ProjectConfig{}, err
	}
	outDir, err := resolveOutputDir(in.OutputDir)
	if err != nil {
		return ProjectConfig{}, err
	}
	m := validVersion.FindStringSubmatch(in.PythonVersion)
	return ProjectConfig{
		ProjectName:      in.ProjectName,
		PackageName:      strings.ReplaceAll(in.ProjectName, "-", "_"),
		Description:      in.Description,
		AuthorName:       in.AuthorName,
		AuthorEmail:      in.AuthorEmail,
		PythonVersion:    in.PythonVersion,
		PythonVersionTag: m[1] + m[2],
		OutputDir:        outDir,
	}, nil
}

func resolveOutputDir(dir string) (string, error) {
	if dir == "" || dir == "~" || strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if dir == "" {
			return home, nil
		}
		return filepath.Join(home, dir[1:]), nil
	}
	return filepath.Abs(dir)
}
