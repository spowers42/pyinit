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
	validVersion = regexp.MustCompile(`^([1-9]\d*)\.(0|[1-9]\d*)(?:\.(?:0|[1-9]\d*))?$`)
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
	_, err := versionTag(s)
	return err
}

// ValidateOutputDir reports whether s can be resolved to an output directory.
func ValidateOutputDir(s string) error {
	_, err := resolveOutputDir(s)
	return err
}

// versionTag converts "3.12" or "3.12.4" to "312" for ruff's target-version.
func versionTag(version string) (string, error) {
	parts := validVersion.FindStringSubmatch(version)
	if parts == nil {
		return "", errors.New("python version must be MAJOR.MINOR or MAJOR.MINOR.PATCH (e.g. 3.12)")
	}
	return parts[1] + parts[2], nil
}

// New validates in and returns a complete ProjectConfig.
func New(in Input) (ProjectConfig, error) {
	if err := ValidateName(in.ProjectName); err != nil {
		return ProjectConfig{}, err
	}
	tag, err := versionTag(in.PythonVersion)
	if err != nil {
		return ProjectConfig{}, err
	}
	outDir, err := resolveOutputDir(in.OutputDir)
	if err != nil {
		return ProjectConfig{}, err
	}
	return ProjectConfig{
		ProjectName:      in.ProjectName,
		PackageName:      strings.ReplaceAll(in.ProjectName, "-", "_"),
		Description:      in.Description,
		AuthorName:       in.AuthorName,
		AuthorEmail:      in.AuthorEmail,
		PythonVersion:    in.PythonVersion,
		PythonVersionTag: tag,
		OutputDir:        outDir,
	}, nil
}

// resolveOutputDir defaults "" to the home directory, expands a leading "~" or
// "~/", and makes the result absolute. "~user" forms are rejected.
func resolveOutputDir(dir string) (string, error) {
	switch {
	case dir == "":
		return os.UserHomeDir()
	case dir == "~" || strings.HasPrefix(dir, "~/"):
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, dir[1:]), nil
	case dir[0] == '~':
		return "", errors.New("only ~ and ~/path are supported, not ~user")
	}
	return filepath.Abs(dir)
}
