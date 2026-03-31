package config

import (
    "bufio"
    "os"
    "strings"
)

type Config struct {
    Port string
}

func loadEnvFile(path string) map[string]string {
    out := make(map[string]string)
    f, err := os.Open(path)
    if err != nil {
        return out
    }
    defer f.Close()

    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }
        key := strings.TrimSpace(parts[0])
        val := strings.TrimSpace(parts[1])
        out[key] = val
    }
    return out
}

func Load() *Config {
    env := loadEnvFile(".env")

    cfg := &Config{Port: "8080"}

    if v, ok := env["CALENDAR_PORT"]; ok && v != "" {
        cfg.Port = v
    }

    if v := os.Getenv("CALENDAR_PORT"); v != "" {
        cfg.Port = v
    }

    return cfg
}
