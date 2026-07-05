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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Environment variables that back flags without a dedicated resolver. Precedence
// is always flag > env > default: env values apply only when the flag was not
// explicitly set on the command line.
const (
	envJobs           = "DUPFIND_JOBS"
	envNoTUI          = "DUPFIND_NO_TUI"
	envHash           = "DUPFIND_HASH"
	envFormat         = "DUPFIND_FORMAT"
	envExclude        = "DUPFIND_EXCLUDE"
	envFollowSymlinks = "DUPFIND_FOLLOW_SYMLINKS"
	envColor          = "DUPFIND_COLOR"
	envSI             = "DUPFIND_SI"
)

// applyStringEnv overrides *dst with the env value when the flag was not set.
func applyStringEnv(cmd *cobra.Command, flag, env string, dst *string) {
	if cmd.Flags().Changed(flag) {
		return
	}
	if v, ok := os.LookupEnv(env); ok {
		*dst = v
	}
}

// applyStringSliceEnv overrides *dst with a comma-separated env value when the
// flag was not set.
func applyStringSliceEnv(cmd *cobra.Command, flag, env string, dst *[]string) {
	if cmd.Flags().Changed(flag) {
		return
	}
	v, ok := os.LookupEnv(env)
	if !ok || v == "" {
		return
	}
	var out []string
	for _, part := range strings.Split(v, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	*dst = out
}

// applyBoolEnv overrides *dst with a boolean env value when the flag was not
// set. An unparseable value is an error.
func applyBoolEnv(cmd *cobra.Command, flag, env string, dst *bool) error {
	if cmd.Flags().Changed(flag) {
		return nil
	}
	v, ok := os.LookupEnv(env)
	if !ok || v == "" {
		return nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fmt.Errorf("invalid %s: %q is not a boolean", env, v)
	}
	*dst = b
	return nil
}

// applyIntEnv overrides *dst with an integer env value when the flag was not
// set. An unparseable value is an error.
func applyIntEnv(cmd *cobra.Command, flag, env string, dst *int) error {
	if cmd.Flags().Changed(flag) {
		return nil
	}
	v, ok := os.LookupEnv(env)
	if !ok || v == "" {
		return nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fmt.Errorf("invalid %s: %q is not an integer", env, v)
	}
	*dst = n
	return nil
}

// resolveSI decides the SI-units setting: the --si flag, else DUPFIND_SI.
func resolveSI(cmd *cobra.Command) (bool, error) {
	si, _ := cmd.Flags().GetBool("si")
	if err := applyBoolEnv(cmd, "si", envSI, &si); err != nil {
		return false, err
	}
	return si, nil
}
