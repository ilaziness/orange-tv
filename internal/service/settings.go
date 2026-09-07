package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ilaziness/orange-tv/internal/constant"
	"github.com/ilaziness/orange-tv/internal/dto"
	"github.com/ilaziness/orange-tv/internal/model"
)

// StrVal reads a string setting from the map, returning "" if missing.
func StrVal(m map[string]model.SystemSettings, key string) string {
	it, ok := m[key]
	if !ok {
		return ""
	}
	return it.SettingValue
}

// IntVal reads an integer setting from the map, returning def if missing or invalid.
func IntVal(m map[string]model.SystemSettings, key string, def int) int {
	v := strings.TrimSpace(StrVal(m, key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// BoolVal reads a boolean setting from the map, returning def if missing or invalid.
func BoolVal(m map[string]model.SystemSettings, key string, def bool) bool {
	v := strings.TrimSpace(StrVal(m, key))
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		if n, err := strconv.Atoi(v); err == nil {
			return n != 0
		}
		return def
	}
}

// ParsePlatformFlags parses a per-platform JSON setting value.
// Missing or invalid JSON falls back to def for all platforms; missing keys use def.
func ParsePlatformFlags(raw string, def bool) dto.PlatformFlags {
	out := dto.PlatformFlags{Web: def, Desktop: def, App: def, TV: def}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	var partial map[string]bool
	if err := json.Unmarshal([]byte(raw), &partial); err != nil {
		return out
	}
	if v, ok := partial[constant.ClientTypeWeb]; ok {
		out.Web = v
	}
	if v, ok := partial[constant.ClientTypeDesktop]; ok {
		out.Desktop = v
	}
	if v, ok := partial[constant.ClientTypeApp]; ok {
		out.App = v
	}
	if v, ok := partial[constant.ClientTypeTV]; ok {
		out.TV = v
	}
	return out
}

// PickPlatformFlag returns the flag for the given client type.
func PickPlatformFlag(flags dto.PlatformFlags, clientType string) bool {
	switch clientType {
	case constant.ClientTypeDesktop:
		return flags.Desktop
	case constant.ClientTypeApp:
		return flags.App
	case constant.ClientTypeTV:
		return flags.TV
	default:
		return flags.Web
	}
}

// PlatformBoolVal reads a per-platform JSON feature flag for clientType.
func PlatformBoolVal(m map[string]model.SystemSettings, key, clientType string, def bool) bool {
	return PickPlatformFlag(ParsePlatformFlags(StrVal(m, key), def), clientType)
}

// MarshalPlatformFlags encodes platform flags as a compact JSON string for storage.
func MarshalPlatformFlags(flags dto.PlatformFlags) string {
	b, err := json.Marshal(flags)
	if err != nil {
		return `{"web":false,"desktop":false,"app":false,"tv":false}`
	}
	return string(b)
}

// DecodePlatformFlags unmarshals platform flags and requires all four platform keys.
func DecodePlatformFlags(raw json.RawMessage) (dto.PlatformFlags, error) {
	raw = json.RawMessage(strings.TrimSpace(string(raw)))
	if len(raw) == 0 || string(raw) == "null" {
		return dto.PlatformFlags{}, fmt.Errorf("端开关不能为空")
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		return dto.PlatformFlags{}, fmt.Errorf("无效的端开关数据")
	}
	for _, k := range []string{
		constant.ClientTypeWeb,
		constant.ClientTypeDesktop,
		constant.ClientTypeApp,
		constant.ClientTypeTV,
	} {
		v, ok := keys[k]
		if !ok {
			return dto.PlatformFlags{}, fmt.Errorf("端开关缺少 %s", k)
		}
		switch string(v) {
		case "true", "false":
			// ok
		case "null", "":
			return dto.PlatformFlags{}, fmt.Errorf("端开关 %s 不能为空", k)
		default:
			return dto.PlatformFlags{}, fmt.Errorf("端开关 %s 必须是布尔值", k)
		}
	}
	var flags dto.PlatformFlags
	if err := json.Unmarshal(raw, &flags); err != nil {
		return dto.PlatformFlags{}, fmt.Errorf("无效的端开关数据")
	}
	return flags, nil
}

// AndPlatformFlags returns the per-platform AND of a and b.
func AndPlatformFlags(a, b dto.PlatformFlags) dto.PlatformFlags {
	return dto.PlatformFlags{
		Web:     a.Web && b.Web,
		Desktop: a.Desktop && b.Desktop,
		App:     a.App && b.App,
		TV:      a.TV && b.TV,
	}
}
