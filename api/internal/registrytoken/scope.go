package registrytoken

import (
	"fmt"
	"strings"
)

// Scope represents a Docker Registry token scope.
// Format: type:name:actions
// Example: repository:demo/nginx:pull,push
type Scope struct {
	Type    string
	Name    string
	Actions []string
}

// ParseScope parses a scope string according to the Distribution spec.
// The name may contain colons (e.g., port numbers), so we must NOT
// blindly split by ":" into 3 parts. We split from the right.
func ParseScope(s string) (Scope, error) {
	// Find the last colon that separates name from actions.
	lastColon := strings.LastIndex(s, ":")
	if lastColon == -1 {
		return Scope{}, fmt.Errorf("invalid scope: %s", s)
	}

	actionsStr := s[lastColon+1:]
	prefix := s[:lastColon]

	// Now prefix is "type:name"
	firstColon := strings.Index(prefix, ":")
	if firstColon == -1 {
		return Scope{}, fmt.Errorf("invalid scope: %s", s)
	}

	scopeType := prefix[:firstColon]
	name := prefix[firstColon+1:]

	actions := strings.Split(actionsStr, ",")
	for i := range actions {
		actions[i] = strings.TrimSpace(actions[i])
	}

	return Scope{
		Type:    scopeType,
		Name:    name,
		Actions: actions,
	}, nil
}

func IntersectActions(requested, granted []string) []string {
	grantedSet := make(map[string]struct{}, len(granted))
	for _, a := range granted {
		grantedSet[a] = struct{}{}
	}
	var result []string
	for _, a := range requested {
		if _, ok := grantedSet[a]; ok {
			result = append(result, a)
		}
	}
	return result
}
