package sqlite3

import (
	"fmt"
	"regexp"
	"strings"
)

// SQLIdentifierValidator provides secure validation and escaping for SQL identifiers
type SQLIdentifierValidator struct {
	// allowedIdentifierPattern defines the allowed pattern for SQL identifiers
	// Allows alphanumeric characters, underscores, and hyphens
	allowedIdentifierPattern *regexp.Regexp

	// maxIdentifierLength defines the maximum allowed length for identifiers
	maxIdentifierLength int
}

// NewSQLIdentifierValidator creates a new validator instance
func NewSQLIdentifierValidator() *SQLIdentifierValidator {
	return &SQLIdentifierValidator{
		// Pattern allows letters, numbers, underscores, and hyphens only
		// Must start with letter or underscore
		allowedIdentifierPattern: regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_-]*$`),
		maxIdentifierLength:      64, // SQLite identifier limit
	}
}

// ValidateIdentifier validates that an identifier is safe to use in SQL
func (v *SQLIdentifierValidator) ValidateIdentifier(identifier string) error {
	if identifier == "" {
		return fmt.Errorf("identifier cannot be empty")
	}

	if len(identifier) > v.maxIdentifierLength {
		return fmt.Errorf("identifier too long: %d characters (max %d)",
			len(identifier), v.maxIdentifierLength)
	}

	if !v.allowedIdentifierPattern.MatchString(identifier) {
		return fmt.Errorf("identifier contains invalid characters: %s", identifier)
	}

	// Check for SQL keywords that should not be used as identifiers
	if v.isSQLKeyword(identifier) {
		return fmt.Errorf("identifier cannot be a SQL keyword: %s", identifier)
	}

	return nil
}

// EscapeIdentifier safely escapes an SQL identifier using double quotes
// This should only be used after validation
func (v *SQLIdentifierValidator) EscapeIdentifier(identifier string) string {
	// Escape any double quotes in the identifier by doubling them
	escaped := strings.ReplaceAll(identifier, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

// SafeIdentifier validates and escapes an identifier in one step
func (v *SQLIdentifierValidator) SafeIdentifier(identifier string) (string, error) {
	if err := v.ValidateIdentifier(identifier); err != nil {
		return "", err
	}
	return v.EscapeIdentifier(identifier), nil
}

// isSQLKeyword checks if the identifier is a common SQL keyword
func (v *SQLIdentifierValidator) isSQLKeyword(identifier string) bool {
	// Convert to uppercase for comparison
	upper := strings.ToUpper(identifier)

	// Common SQL keywords that should be avoided as table names
	keywords := map[string]bool{
		"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true,
		"CREATE": true, "DROP": true, "ALTER": true, "TABLE": true,
		"INDEX": true, "VIEW": true, "TRIGGER": true, "DATABASE": true,
		"SCHEMA": true, "COLUMN": true, "CONSTRAINT": true, "PRIMARY": true,
		"KEY": true, "FOREIGN": true, "REFERENCES": true, "UNIQUE": true,
		"NOT": true, "NULL": true, "DEFAULT": true, "CHECK": true,
		"FROM": true, "WHERE": true, "GROUP": true, "BY": true,
		"HAVING": true, "ORDER": true, "LIMIT": true, "OFFSET": true,
		"UNION": true, "JOIN": true, "INNER": true, "LEFT": true,
		"RIGHT": true, "OUTER": true, "ON": true, "AS": true,
		"AND": true, "OR": true, "IN": true, "EXISTS": true,
		"BETWEEN": true, "LIKE": true, "IS": true, "CASE": true,
		"WHEN": true, "THEN": true, "ELSE": true, "END": true,
		"IF": true, "BEGIN": true, "COMMIT": true, "ROLLBACK": true,
		"TRANSACTION": true, "SAVEPOINT": true, "RELEASE": true,
		"PRAGMA": true, "EXPLAIN": true, "ANALYZE": true,
		"VACUUM": true, "REINDEX": true, "ATTACH": true, "DETACH": true,
	}

	return keywords[upper]
}
