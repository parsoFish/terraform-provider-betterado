package testutils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// loadSecretsEnv makes the live acceptance tests self-sufficient for credentials:
// it walks up from the working directory to find a `secrets.env` file at the repo
// root (the first-class forge live-validation pattern) and loads its KEY=VALUE
// pairs into the process environment. This means the tests do NOT depend on the
// exact env the runner (forge, CI, a shell) happens to inject — they provision
// their own creds from the project's secrets.env.
//
// Precedence: an already-exported env var WINS (a shell/CI that set the var is
// never overridden). Missing secrets.env is not an error here — PreCheck then
// reports exactly which required var is absent, pointing the operator at
// secrets.env. Returns the path loaded, or "" if none was found.
//
// secrets.env is gitignored (never committed) and must use the CANONICAL names
// the provider expects: AZDO_ORG_SERVICE_URL + AZDO_PERSONAL_ACCESS_TOKEN.
func loadSecretsEnv() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		candidate := filepath.Join(dir, "secrets.env")
		if data, err := os.ReadFile(candidate); err == nil {
			applyDotenv(string(data))
			return candidate
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "" // reached the filesystem root without finding one
		}
		dir = parent
	}
}

// applyDotenv parses a minimal KEY=VALUE dotenv body and sets each var that is
// not already present in the environment. Blank lines and `#` comments are
// skipped; surrounding single/double quotes on the value are stripped.
func applyDotenv(content string) {
	sc := bufio.NewScanner(strings.NewReader(content))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		key := strings.TrimSpace(line[:eq])
		val := strings.Trim(strings.TrimSpace(line[eq+1:]), `"'`)
		if key == "" {
			continue
		}
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, val)
		}
	}
}
