package gen

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/jinzhu/inflection"
	"gopkg.in/yaml.v3"
)

var goIdentifierRE = regexp.MustCompile(`^[A-Z][a-zA-Z0-9_]*$`)

// Relation describes a logical relation from a source table column to a target table column.
type Relation struct {
	SourceTable  string
	SourceColumn string
	TargetTable  string
	TargetColumn string
	Field        string
	ReverseField string
}

type relationsFile struct {
	Relations []Relation `yaml:"relations"`
}

// LoadRelations reads logical relations from a YAML file.
func LoadRelations(path string) ([]Relation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read relations file: %w", err)
	}

	var file relationsFile
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("parse relations file: %w", err)
	}
	if err := validateRelations(file.Relations); err != nil {
		return nil, err
	}
	return file.Relations, nil
}

func (r *Relation) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		parts := strings.Split(value.Value, "->")
		if len(parts) != 2 {
			return fmt.Errorf("invalid relation %q: expected source.table_column -> target.table_column", value.Value)
		}
		var err error
		r.SourceTable, r.SourceColumn, err = parseRelationEndpoint(strings.TrimSpace(parts[0]))
		if err != nil {
			return err
		}
		r.TargetTable, r.TargetColumn, err = parseRelationEndpoint(strings.TrimSpace(parts[1]))
		return err
	case yaml.MappingNode:
		allowedFields := map[string]struct{}{
			"source":        {},
			"target":        {},
			"field":         {},
			"reverse_field": {},
		}
		seenFields := make(map[string]struct{}, len(value.Content)/2)
		for i := 0; i < len(value.Content); i += 2 {
			field := value.Content[i].Value
			if _, ok := allowedFields[field]; !ok {
				return fmt.Errorf("unknown relation field %q", field)
			}
			if _, ok := seenFields[field]; ok {
				return fmt.Errorf("duplicate relation field %q", field)
			}
			seenFields[field] = struct{}{}
		}

		var raw struct {
			Source       string `yaml:"source"`
			Target       string `yaml:"target"`
			Field        string `yaml:"field"`
			ReverseField string `yaml:"reverse_field"`
		}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		var err error
		r.SourceTable, r.SourceColumn, err = parseRelationEndpoint(raw.Source)
		if err != nil {
			return err
		}
		r.TargetTable, r.TargetColumn, err = parseRelationEndpoint(raw.Target)
		if err != nil {
			return err
		}
		r.Field = raw.Field
		r.ReverseField = raw.ReverseField
		return nil
	default:
		return fmt.Errorf("invalid relation: expected a string or mapping")
	}
}

func parseRelationEndpoint(value string) (string, string, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid relation endpoint %q: expected table.column", value)
	}
	if err := validateTableName(parts[0]); err != nil {
		return "", "", err
	}
	if err := validateTableName(parts[1]); err != nil {
		return "", "", err
	}
	return parts[0], parts[1], nil
}

func validateRelations(relations []Relation) error {
	seen := make(map[string]struct{}, len(relations))
	for _, relation := range relations {
		if err := validateRelation(relation); err != nil {
			return err
		}
		key := relation.key()
		if _, ok := seen[key]; ok {
			return fmt.Errorf("duplicate relation %s", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateRelation(relation Relation) error {
	for _, name := range []string{relation.SourceTable, relation.SourceColumn, relation.TargetTable, relation.TargetColumn} {
		if err := validateTableName(name); err != nil {
			return fmt.Errorf("invalid relation: %w", err)
		}
	}
	for _, name := range []string{relation.Field, relation.ReverseField} {
		if name != "" && !goIdentifierRE.MatchString(name) {
			return fmt.Errorf("invalid relation field %q", name)
		}
	}
	return nil
}

func (r Relation) key() string {
	return r.SourceTable + "." + r.SourceColumn + "->" + r.TargetTable + "." + r.TargetColumn
}

func relationFieldNames(relation Relation) (string, string) {
	field := relation.Field
	if field == "" {
		field = inflection.Singular(ToCamelCase(relation.TargetTable))
	}
	reverseField := relation.ReverseField
	if reverseField == "" {
		reverseField = inflection.Plural(ToCamelCase(relation.SourceTable))
	}
	return field, reverseField
}

func mergeRelations(physical, logical []Relation) ([]Relation, error) {
	merged := make(map[string]Relation, len(physical)+len(logical))
	for _, relation := range physical {
		merged[relation.key()] = relation
	}
	for _, relation := range logical {
		key := relation.key()
		if existing, ok := merged[key]; ok {
			if relation.Field != "" {
				existing.Field = relation.Field
			}
			if relation.ReverseField != "" {
				existing.ReverseField = relation.ReverseField
			}
			merged[key] = existing
			continue
		}
		merged[key] = relation
	}

	keys := make([]string, 0, len(merged))
	for key := range merged {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	result := make([]Relation, 0, len(merged))
	fields := make(map[string]string, len(merged)*2)
	for _, key := range keys {
		relation := merged[key]
		if err := validateRelation(relation); err != nil {
			return nil, err
		}
		field, reverseField := relationFieldNames(relation)
		if relation.SourceTable == relation.TargetTable && field == reverseField {
			return nil, fmt.Errorf("relation field conflict on %s.%s", relation.SourceTable, field)
		}
		for key, name := range map[string]string{
			relation.SourceTable + "." + field:        relation.key(),
			relation.TargetTable + "." + reverseField: relation.key(),
		} {
			if existing, ok := fields[key]; ok && existing != name {
				return nil, fmt.Errorf("relation field conflict on %s between %s and %s", key, existing, name)
			}
			fields[key] = name
		}
		result = append(result, relation)
	}
	return result, nil
}
