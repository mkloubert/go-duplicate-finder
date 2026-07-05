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
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/mkloubert/go-duplicate-finder/internal/model"
)

// envReportFile names the report file when --report-file is unset.
const envReportFile = "DUPFIND_REPORT_FILE"

// resolveReportInput reads the report bytes by precedence: --report-file, then
// DUPFIND_REPORT_FILE, then STDIN (only when it is piped, i.e. not a TTY).
func resolveReportInput(flagFile string, stdin io.Reader, stdinIsTTY bool) ([]byte, error) {
	if flagFile != "" {
		return os.ReadFile(flagFile)
	}
	if env := os.Getenv(envReportFile); env != "" {
		return os.ReadFile(env)
	}
	if !stdinIsTTY {
		return io.ReadAll(stdin)
	}
	return nil, fmt.Errorf("no input: pass --report-file, set %s, or pipe JSON to stdin", envReportFile)
}

// parseReport unmarshals report bytes into a model.Output. An empty result is
// valid; malformed JSON is an error.
func parseReport(data []byte) (*model.Output, error) {
	var out model.Output
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("invalid report JSON: %w", err)
	}
	if out.Result == nil {
		out.Result = map[string]*model.FileResult{}
	}
	return &out, nil
}
