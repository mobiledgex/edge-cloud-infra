package orm

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/mobiledgex/edge-cloud/util"
)

func ValidName(name string) error {
	err := util.ValidObjName(name)
	if err != nil {
		return err
	}
	if strings.HasPrefix(name, ".") {
		return fmt.Errorf("Name cannot start with '.'")
	}
	if strings.HasPrefix(name, "-") {
		return fmt.Errorf("Name cannot start with '-'")
	}
	if strings.HasSuffix(name, ".") {
		return fmt.Errorf("Name cannot end with '.'")
	}
	if strings.Contains(name, "::") {
		return fmt.Errorf("Name cannot contain ::")
	}
	if strings.Contains(name, "&") {
		return fmt.Errorf("Name cannot contain &")
	}
	if strings.HasSuffix(name, ".git") {
		return fmt.Errorf("Name cannot end with '.git'")
	}
	if strings.HasSuffix(name, ".atom") {
		return fmt.Errorf("Name cannot end with '.atom'")
	}
	if strings.HasSuffix(name, "-cache") {
		return fmt.Errorf("Name cannot end with '-cache'")
	}
	return nil
}

// Gitlab groups can only contain letters, digits, _ . -
// cannot start with '-' or end in '.', '.git' or '.atom'
// This combines the rules for both name and path.
func GitlabGroupSanitize(name string) string {
	name = strings.TrimPrefix(name, "-")
	name = strings.TrimSuffix(name, ".")
	if strings.HasSuffix(name, ".git") {
		name = name[:len(name)-4] + "-git"
	}
	if strings.HasSuffix(name, ".atom") {
		name = name[:len(name)-5] + "-atom"
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) ||
			r == '_' || r == '.' || r == '-' {
			return r
		}
		return '-'
	}, name)
}
