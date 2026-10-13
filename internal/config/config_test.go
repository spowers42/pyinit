package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		in      Input
		want    ProjectConfig
		wantErr string
	}{
		{
			name: "hyphens become underscores",
			in:   Input{ProjectName: "my-cool-app", PythonVersion: "3.12", OutputDir: "/tmp/x"},
			want: ProjectConfig{ProjectName: "my-cool-app", PackageName: "my_cool_app", PythonVersion: "3.12", PythonVersionTag: "312", OutputDir: "/tmp/x"},
		},
		{
			name: "patch version tag",
			in:   Input{ProjectName: "app", PythonVersion: "3.12.4", OutputDir: "/tmp/x"},
			want: ProjectConfig{ProjectName: "app", PackageName: "app", PythonVersion: "3.12.4", PythonVersionTag: "312", OutputDir: "/tmp/x"},
		},
		{
			name: "empty output dir defaults to home",
			in:   Input{ProjectName: "app", PythonVersion: "3.11"},
			want: ProjectConfig{ProjectName: "app", PackageName: "app", PythonVersion: "3.11", PythonVersionTag: "311", OutputDir: home},
		},
		{
			name: "tilde expanded",
			in:   Input{ProjectName: "app", PythonVersion: "3.11", OutputDir: "~/code"},
			want: ProjectConfig{ProjectName: "app", PackageName: "app", PythonVersion: "3.11", PythonVersionTag: "311", OutputDir: filepath.Join(home, "code")},
		},
		{
			name: "bare tilde",
			in:   Input{ProjectName: "app", PythonVersion: "3.11", OutputDir: "~"},
			want: ProjectConfig{ProjectName: "app", PackageName: "app", PythonVersion: "3.11", PythonVersionTag: "311", OutputDir: home},
		},
		{
			name: "relative made absolute",
			in:   Input{ProjectName: "app", PythonVersion: "3.11", OutputDir: "sub"},
			want: ProjectConfig{ProjectName: "app", PackageName: "app", PythonVersion: "3.11", PythonVersionTag: "311", OutputDir: filepath.Join(cwd, "sub")},
		},
		{
			name: "passthrough fields",
			in:   Input{ProjectName: "app", Description: "d", AuthorName: "A", AuthorEmail: "a@b.c", PythonVersion: "3.13", OutputDir: "/tmp/x"},
			want: ProjectConfig{ProjectName: "app", PackageName: "app", Description: "d", AuthorName: "A", AuthorEmail: "a@b.c", PythonVersion: "3.13", PythonVersionTag: "313", OutputDir: "/tmp/x"},
		},
		{name: "empty name", in: Input{PythonVersion: "3.12"}, wantErr: "project name is required"},
		{name: "uppercase name", in: Input{ProjectName: "App", PythonVersion: "3.12"}, wantErr: "must be lowercase and start with a letter (a-z, 0-9, - and _ allowed)"},
		{name: "digit-leading name", in: Input{ProjectName: "1app", PythonVersion: "3.12"}, wantErr: "must be lowercase and start with a letter (a-z, 0-9, - and _ allowed)"},
		{name: "empty version", in: Input{ProjectName: "app"}, wantErr: "python version must be MAJOR.MINOR or MAJOR.MINOR.PATCH (e.g. 3.12)"},
		{name: "bad version word", in: Input{ProjectName: "app", PythonVersion: "latest"}, wantErr: "python version must be MAJOR.MINOR or MAJOR.MINOR.PATCH (e.g. 3.12)"},
		{name: "major only", in: Input{ProjectName: "app", PythonVersion: "3"}, wantErr: "python version must be MAJOR.MINOR or MAJOR.MINOR.PATCH (e.g. 3.12)"},
		{name: "four parts", in: Input{ProjectName: "app", PythonVersion: "3.12.1.2"}, wantErr: "python version must be MAJOR.MINOR or MAJOR.MINOR.PATCH (e.g. 3.12)"},
		{name: "trailing dot", in: Input{ProjectName: "app", PythonVersion: "3.12."}, wantErr: "python version must be MAJOR.MINOR or MAJOR.MINOR.PATCH (e.g. 3.12)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.in)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
