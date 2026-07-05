// Copyright © 2026 Marcel Joachim Kloubert <marcel@kloubert.dev>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
)

// envCwd is the environment variable that overrides the working directory.
const envCwd = "DUPFIND_CWD"

// resolveBaseDir determines the base directory for glob expansion.
//
// Precedence: the --cwd flag value, then the DUPFIND_CWD environment variable,
// then the process working directory. An explicit value (flag or env) is made
// absolute and must point to an existing directory.
func resolveBaseDir(flagCwd string) (string, error) {
	dir := flagCwd
	if dir == "" {
		dir = os.Getenv(envCwd)
	}
	if dir == "" {
		return os.Getwd()
	}

	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("invalid working directory %q: %w", dir, err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("working directory %q: %w", abs, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("working directory %q is not a directory", abs)
	}
	return abs, nil
}
