package xbindata

import (
	"context"
	"fmt"
	"os"
	"strings"
)

type configContextKeyType uint8

const (
	ContextEnvKey configContextKeyType = iota
	ContextInputKey
)

func ContextWithEnv(ctx context.Context, env map[string]string, noInherits ...bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	} else {
		if len(noInherits) == 0 || !noInherits[0] {
			oldEnv := ctx.Value(ContextEnvKey)
			if oldEnv != nil {
				for key, value := range oldEnv.(map[string]string) {
					env[key] = value
				}
			}
		}
	}
	return context.WithValue(ctx, ContextEnvKey, env)
}

func ContextEnv(ctx context.Context) map[string]string {
	if ctx != nil {
		env := ctx.Value(ContextEnvKey)
		if env != nil {
			return env.(map[string]string)
		}
	}
	return map[string]string{}
}

func ContextWithInputKey(ctx context.Context, key string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ContextInputKey, key)
}

func InputKey(ctx context.Context, format ...string) (key string) {
	if ctx != nil {
		old := ctx.Value(ContextInputKey)
		if old != nil {
			if len(format) == 0 || format[0] == "" {
				return old.(string)
			}
			return fmt.Sprintf(format[0], old)
		}
	}
	return
}

func Env(update ...map[string]string) map[string]string {
	items := make(map[string]string)
	for _, item := range os.Environ() {
		kv := strings.Split(item, "=")
		items[kv[0]] = kv[1]
	}
	for _, env := range update {
		for k, v := range env {
			items[k] = v
		}
	}
	return items
}

func EnvS(update ...map[string]string) (env []string) {
	for k, v := range Env(update...) {
		env = append(env, k+"="+v)
	}
	return
}
