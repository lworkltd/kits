package rabbit

import "strings"

func filterBooleanConfigs(defaultConfig *map[string]bool, prefix string, cs map[string]bool, caseSensitive bool) map[string]bool {
	for key, value := range cs {
		index := strings.Index(key, prefix)
		if index < 0 {
			continue
		}

		k := string(key[len(prefix):])
		if !caseSensitive {
			k = strings.ToLower(k)
		}

		(*defaultConfig)[k] = value
	}

	return *defaultConfig
}
