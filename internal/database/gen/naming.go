package gen

import (
	"fmt"
	"strings"
)

// ToCamelCase converts snake_case to CamelCase.
func ToCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = titleCaser.String(parts[i])
	}
	return strings.Join(parts, "")
}

// ToSnakeCase converts CamelCase to snake_case.
func ToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			prev := rune(s[i-1])
			if prev >= 'a' && prev <= 'z' {
				result = append(result, '_')
			} else if i+1 < len(s) && rune(s[i+1]) >= 'a' && rune(s[i+1]) <= 'z' && prev >= 'A' && prev <= 'Z' {
				result = append(result, '_')
			}
		}
		result = append(result, r)
	}
	return strings.ToLower(string(result))
}

// MapSQLTypeToGo maps a SQL type name to a Go type.
func MapSQLTypeToGo(sqlType string, nullable bool) string {
	typeMap := map[string]string{
		"tinyint":   "int8",
		"smallint":  "int16",
		"int":       "int32",
		"integer":   "int32",
		"bigint":    "int64",
		"float":     "float32",
		"double":    "float64",
		"decimal":   "float64",
		"numeric":   "float64",
		"char":      "string",
		"varchar":   "string",
		"text":      "string",
		"date":      "time.Time",
		"datetime":  "time.Time",
		"timestamp": "time.Time",
		"boolean":   "bool",
		"json":      "string",
		"real":      "float64",
		"blob":      "[]byte",
	}

	goType, ok := typeMap[strings.ToLower(sqlType)]
	if !ok {
		goType = "string"
	}

	if nullable {
		goType = "*" + goType
	}

	return goType
}

// GenerateModelCode renders Go struct source for a database table.
func GenerateModelCode(opts GenerateOptions, modelName, tableName string, columns []Column, foreignKeys []ForeignKey) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("package %s\n\n", opts.PackageName))

	imports := []string{`"github.com/uptrace/bun"`}
	hasTime := false
	for _, col := range columns {
		goType := MapSQLTypeToGo(col.Type, col.Nullable)
		if goType == "time.Time" || goType == "*time.Time" {
			hasTime = true
			break
		}
	}
	if hasTime {
		imports = append(imports, `"time"`)
	}

	sb.WriteString("import (\n")
	for _, imp := range imports {
		sb.WriteString(fmt.Sprintf("\t%s\n", imp))
	}
	sb.WriteString(")\n\n")

	sb.WriteString(fmt.Sprintf("// %s represents the %s table.\n", modelName, tableName))
	sb.WriteString(fmt.Sprintf("type %s struct {\n", modelName))

	alias := "t"
	if len(tableName) > 0 {
		alias = string(tableName[0])
	}
	sb.WriteString(fmt.Sprintf("\tbun.BaseModel `bun:\"table:%s,alias:%s\"`\n\n", tableName, alias))

	for _, col := range columns {
		if col.Comment != "" {
			sb.WriteString(fmt.Sprintf("\t// %s\n", col.Comment))
		}

		goType := MapSQLTypeToGo(col.Type, col.Nullable)
		fieldName := ToCamelCase(col.Name)

		sb.WriteString(fmt.Sprintf("\t%s %s", fieldName, goType))

		for _, fk := range foreignKeys {
			if fk.ColumnName == col.Name {
				refModel := ToCamelCase(fk.ReferencedTableName)
				refField := ToCamelCase(fk.ReferencedColumnName)
				sb.WriteString(fmt.Sprintf("\n\t// FK: %s -> %s(%s)", col.Name, refModel, refField))
				break
			}
		}

		bunTag := fmt.Sprintf(` bun:"%s"`, col.Name)
		sb.WriteString(bunTag)

		if opts.JSONTags {
			sb.WriteString(fmt.Sprintf(` json:"%s"`, ToSnakeCase(fieldName)))
		}

		if opts.ValidatorTags && !col.Nullable {
			sb.WriteString(` validate:"required"`)
		}

		sb.WriteString("\n")
	}

	sb.WriteString("}\n")

	return sb.String()
}
