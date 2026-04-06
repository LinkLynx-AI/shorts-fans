package dbschema

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

const relationColumnsQuery = `
SELECT
	n.nspname AS schema_name,
	c.relname AS relation_name,
	c.relkind::TEXT AS relation_kind,
	a.attnum AS ordinal_position,
	a.attname AS column_name,
	pg_catalog.format_type(a.atttypid, a.atttypmod) AS data_type,
	NOT a.attnotnull AS is_nullable,
	pg_get_expr(ad.adbin, ad.adrelid) AS column_default
FROM pg_attribute AS a
JOIN pg_class AS c
	ON c.oid = a.attrelid
JOIN pg_namespace AS n
	ON n.oid = c.relnamespace
LEFT JOIN pg_attrdef AS ad
	ON ad.adrelid = a.attrelid
	AND ad.adnum = a.attnum
WHERE c.relkind IN ('r', 'v')
	AND a.attnum > 0
	AND NOT a.attisdropped
	AND n.nspname NOT IN ('pg_catalog', 'information_schema')
	AND NOT (n.nspname = 'public' AND c.relname = 'schema_migrations')
ORDER BY n.nspname, c.relname, c.relkind, a.attnum;
`

const tableConstraintsQuery = `
SELECT
	n.nspname AS schema_name,
	c.relname AS table_name,
	con.conname AS constraint_name,
	con.contype::TEXT AS constraint_type,
	pg_get_constraintdef(con.oid, true) AS definition,
	COALESCE(
		ARRAY(
			SELECT att.attname
			FROM unnest(con.conkey) WITH ORDINALITY AS cols(attnum, ord)
			JOIN pg_attribute AS att
				ON att.attrelid = con.conrelid
				AND att.attnum = cols.attnum
			ORDER BY cols.ord
		),
		ARRAY[]::TEXT[]
	) AS columns,
	refn.nspname AS ref_schema_name,
	refc.relname AS ref_table_name,
	COALESCE(
		ARRAY(
			SELECT att.attname
			FROM unnest(con.confkey) WITH ORDINALITY AS cols(attnum, ord)
			JOIN pg_attribute AS att
				ON att.attrelid = con.confrelid
				AND att.attnum = cols.attnum
			ORDER BY cols.ord
		),
		ARRAY[]::TEXT[]
	) AS ref_columns,
	con.confdeltype::TEXT AS on_delete,
	con.confupdtype::TEXT AS on_update
FROM pg_constraint AS con
JOIN pg_class AS c
	ON c.oid = con.conrelid
JOIN pg_namespace AS n
	ON n.oid = c.relnamespace
LEFT JOIN pg_class AS refc
	ON refc.oid = con.confrelid
LEFT JOIN pg_namespace AS refn
	ON refn.oid = refc.relnamespace
WHERE c.relkind = 'r'
	AND con.contype IN ('p', 'u', 'f', 'c')
	AND n.nspname NOT IN ('pg_catalog', 'information_schema')
	AND NOT (n.nspname = 'public' AND c.relname = 'schema_migrations')
ORDER BY n.nspname, c.relname, con.contype, con.conname;
`

const tableIndexesQuery = `
SELECT
	n.nspname AS schema_name,
	t.relname AS table_name,
	idx.relname AS index_name,
	pg_get_indexdef(i.indexrelid) AS definition,
	i.indisunique AS is_unique,
	COALESCE(
		ARRAY(
			SELECT att.attname
			FROM unnest(i.indkey) WITH ORDINALITY AS cols(attnum, ord)
			JOIN pg_attribute AS att
				ON att.attrelid = i.indrelid
				AND att.attnum = cols.attnum
			WHERE cols.attnum > 0
			ORDER BY cols.ord
		),
		ARRAY[]::TEXT[]
	) AS columns,
	pg_get_expr(i.indpred, i.indrelid) AS predicate
FROM pg_index AS i
JOIN pg_class AS idx
	ON idx.oid = i.indexrelid
JOIN pg_class AS t
	ON t.oid = i.indrelid
JOIN pg_namespace AS n
	ON n.oid = t.relnamespace
LEFT JOIN pg_constraint AS con
	ON con.conindid = i.indexrelid
WHERE t.relkind = 'r'
	AND con.oid IS NULL
	AND n.nspname NOT IN ('pg_catalog', 'information_schema')
	AND NOT (n.nspname = 'public' AND t.relname = 'schema_migrations')
ORDER BY n.nspname, t.relname, idx.relname;
`

const viewsQuery = `
SELECT
	n.nspname AS schema_name,
	c.relname AS view_name,
	pg_get_viewdef(c.oid, true) AS definition
FROM pg_class AS c
JOIN pg_namespace AS n
	ON n.oid = c.relnamespace
WHERE c.relkind = 'v'
	AND n.nspname NOT IN ('pg_catalog', 'information_schema')
ORDER BY n.nspname, c.relname;
`

type catalog struct {
	Columns     []relationColumn
	Constraints []tableConstraint
	Indexes     []tableIndex
	Views       []viewDefinition
}

type relationColumn struct {
	SchemaName      string
	RelationName    string
	RelationKind    string
	OrdinalPosition int
	ColumnName      string
	DataType        string
	IsNullable      bool
	ColumnDefault   string
}

type tableConstraint struct {
	SchemaName     string
	TableName      string
	ConstraintName string
	ConstraintType string
	Definition     string
	Columns        []string
	RefSchemaName  string
	RefTableName   string
	RefColumns     []string
	OnDelete       string
	OnUpdate       string
}

type tableIndex struct {
	SchemaName string
	TableName  string
	IndexName  string
	Definition string
	IsUnique   bool
	Columns    []string
	Predicate  string
}

type viewDefinition struct {
	SchemaName string
	ViewName   string
	Definition string
}

func inspectCatalog(ctx context.Context, config *pgx.ConnConfig) (catalog, error) {
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return catalog{}, fmt.Errorf("connect temp database for inspection: %w", err)
	}
	defer conn.Close(ctx)

	columns, err := loadRelationColumns(ctx, conn)
	if err != nil {
		return catalog{}, err
	}

	constraints, err := loadTableConstraints(ctx, conn)
	if err != nil {
		return catalog{}, err
	}

	indexes, err := loadTableIndexes(ctx, conn)
	if err != nil {
		return catalog{}, err
	}

	views, err := loadViews(ctx, conn)
	if err != nil {
		return catalog{}, err
	}

	return catalog{
		Columns:     columns,
		Constraints: constraints,
		Indexes:     indexes,
		Views:       views,
	}, nil
}

func loadRelationColumns(ctx context.Context, conn *pgx.Conn) ([]relationColumn, error) {
	rows, err := conn.Query(ctx, relationColumnsQuery)
	if err != nil {
		return nil, fmt.Errorf("query relation columns: %w", err)
	}
	defer rows.Close()

	var columns []relationColumn
	for rows.Next() {
		var row relationColumn
		var columnDefault *string

		if err := rows.Scan(
			&row.SchemaName,
			&row.RelationName,
			&row.RelationKind,
			&row.OrdinalPosition,
			&row.ColumnName,
			&row.DataType,
			&row.IsNullable,
			&columnDefault,
		); err != nil {
			return nil, fmt.Errorf("scan relation column: %w", err)
		}

		if columnDefault != nil {
			row.ColumnDefault = strings.TrimSpace(*columnDefault)
		}

		columns = append(columns, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate relation columns: %w", err)
	}

	return columns, nil
}

func loadTableConstraints(ctx context.Context, conn *pgx.Conn) ([]tableConstraint, error) {
	rows, err := conn.Query(ctx, tableConstraintsQuery)
	if err != nil {
		return nil, fmt.Errorf("query table constraints: %w", err)
	}
	defer rows.Close()

	var constraints []tableConstraint
	for rows.Next() {
		var row tableConstraint
		var refSchemaName *string
		var refTableName *string

		if err := rows.Scan(
			&row.SchemaName,
			&row.TableName,
			&row.ConstraintName,
			&row.ConstraintType,
			&row.Definition,
			&row.Columns,
			&refSchemaName,
			&refTableName,
			&row.RefColumns,
			&row.OnDelete,
			&row.OnUpdate,
		); err != nil {
			return nil, fmt.Errorf("scan table constraint: %w", err)
		}

		row.Definition = strings.TrimSpace(row.Definition)
		if refSchemaName != nil {
			row.RefSchemaName = strings.TrimSpace(*refSchemaName)
		}
		if refTableName != nil {
			row.RefTableName = strings.TrimSpace(*refTableName)
		}
		constraints = append(constraints, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate table constraints: %w", err)
	}

	return constraints, nil
}

func loadTableIndexes(ctx context.Context, conn *pgx.Conn) ([]tableIndex, error) {
	rows, err := conn.Query(ctx, tableIndexesQuery)
	if err != nil {
		return nil, fmt.Errorf("query table indexes: %w", err)
	}
	defer rows.Close()

	var indexes []tableIndex
	for rows.Next() {
		var row tableIndex
		var predicate *string

		if err := rows.Scan(
			&row.SchemaName,
			&row.TableName,
			&row.IndexName,
			&row.Definition,
			&row.IsUnique,
			&row.Columns,
			&predicate,
		); err != nil {
			return nil, fmt.Errorf("scan table index: %w", err)
		}

		row.Definition = strings.TrimSpace(row.Definition)
		if predicate != nil {
			row.Predicate = strings.TrimSpace(*predicate)
		}

		indexes = append(indexes, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate table indexes: %w", err)
	}

	return indexes, nil
}

func loadViews(ctx context.Context, conn *pgx.Conn) ([]viewDefinition, error) {
	rows, err := conn.Query(ctx, viewsQuery)
	if err != nil {
		return nil, fmt.Errorf("query views: %w", err)
	}
	defer rows.Close()

	var views []viewDefinition
	for rows.Next() {
		var row viewDefinition

		if err := rows.Scan(
			&row.SchemaName,
			&row.ViewName,
			&row.Definition,
		); err != nil {
			return nil, fmt.Errorf("scan view definition: %w", err)
		}

		row.Definition = strings.TrimSpace(row.Definition)
		views = append(views, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate views: %w", err)
	}

	sort.Slice(views, func(i, j int) bool {
		if views[i].SchemaName != views[j].SchemaName {
			return views[i].SchemaName < views[j].SchemaName
		}
		return views[i].ViewName < views[j].ViewName
	})

	return views, nil
}

type schemaDocument struct {
	Source  documentSource `yaml:"source"`
	Schemas []schemaNode   `yaml:"schemas"`
}

type documentSource struct {
	Generator    string   `yaml:"generator"`
	MigrationDir string   `yaml:"migration_dir"`
	Migrations   []string `yaml:"migrations"`
}

type schemaNode struct {
	Name   string      `yaml:"name"`
	Tables []tableNode `yaml:"tables,omitempty"`
	Views  []viewNode  `yaml:"views,omitempty"`
}

type tableNode struct {
	Name              string                `yaml:"name"`
	Columns           []columnNode          `yaml:"columns"`
	PrimaryKey        []string              `yaml:"primary_key,omitempty"`
	ForeignKeys       []foreignKeyNode      `yaml:"foreign_keys,omitempty"`
	UniqueConstraints []namedColumnsNode    `yaml:"unique_constraints,omitempty"`
	CheckConstraints  []namedExpressionNode `yaml:"check_constraints,omitempty"`
	Indexes           []indexNode           `yaml:"indexes,omitempty"`
}

type columnNode struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	Nullable bool   `yaml:"nullable"`
	Default  string `yaml:"default,omitempty"`
}

type foreignKeyNode struct {
	Name       string               `yaml:"name"`
	Columns    []string             `yaml:"columns"`
	References foreignKeyTargetNode `yaml:"references"`
	OnDelete   string               `yaml:"on_delete,omitempty"`
	OnUpdate   string               `yaml:"on_update,omitempty"`
}

type foreignKeyTargetNode struct {
	Schema  string   `yaml:"schema"`
	Table   string   `yaml:"table"`
	Columns []string `yaml:"columns"`
}

type namedColumnsNode struct {
	Name    string   `yaml:"name"`
	Columns []string `yaml:"columns"`
}

type namedExpressionNode struct {
	Name       string `yaml:"name"`
	Expression string `yaml:"expression"`
}

type indexNode struct {
	Name       string   `yaml:"name"`
	Columns    []string `yaml:"columns,omitempty"`
	Unique     bool     `yaml:"unique,omitempty"`
	Predicate  string   `yaml:"predicate,omitempty"`
	Definition string   `yaml:"definition"`
}

type viewNode struct {
	Name       string       `yaml:"name"`
	Columns    []columnNode `yaml:"columns"`
	Definition string       `yaml:"definition"`
}

type schemaBuilder struct {
	node   schemaNode
	tables map[string]*tableNode
	views  map[string]*viewNode
}

func buildDocument(source documentSource, catalog catalog) schemaDocument {
	schemas := make(map[string]*schemaBuilder)

	getSchema := func(name string) *schemaBuilder {
		if existing, ok := schemas[name]; ok {
			return existing
		}

		builder := &schemaBuilder{
			node:   schemaNode{Name: name},
			tables: make(map[string]*tableNode),
			views:  make(map[string]*viewNode),
		}
		schemas[name] = builder
		return builder
	}

	for _, relationColumn := range catalog.Columns {
		schema := getSchema(relationColumn.SchemaName)
		column := columnNode{
			Name:     relationColumn.ColumnName,
			Type:     relationColumn.DataType,
			Nullable: relationColumn.IsNullable,
			Default:  relationColumn.ColumnDefault,
		}

		switch relationColumn.RelationKind {
		case "r":
			table := getTable(schema, relationColumn.RelationName)
			table.Columns = append(table.Columns, column)
		case "v":
			view := getView(schema, relationColumn.RelationName)
			view.Columns = append(view.Columns, column)
		}
	}

	for _, constraint := range catalog.Constraints {
		table := getTable(getSchema(constraint.SchemaName), constraint.TableName)

		switch constraint.ConstraintType {
		case "p":
			table.PrimaryKey = append([]string(nil), constraint.Columns...)
		case "u":
			table.UniqueConstraints = append(table.UniqueConstraints, namedColumnsNode{
				Name:    constraint.ConstraintName,
				Columns: append([]string(nil), constraint.Columns...),
			})
		case "c":
			table.CheckConstraints = append(table.CheckConstraints, namedExpressionNode{
				Name:       constraint.ConstraintName,
				Expression: constraint.Definition,
			})
		case "f":
			foreignKey := foreignKeyNode{
				Name:    constraint.ConstraintName,
				Columns: append([]string(nil), constraint.Columns...),
				References: foreignKeyTargetNode{
					Schema:  constraint.RefSchemaName,
					Table:   constraint.RefTableName,
					Columns: append([]string(nil), constraint.RefColumns...),
				},
			}

			if action := foreignKeyAction(constraint.OnDelete); action != "NO ACTION" && action != "" {
				foreignKey.OnDelete = action
			}
			if action := foreignKeyAction(constraint.OnUpdate); action != "NO ACTION" && action != "" {
				foreignKey.OnUpdate = action
			}

			table.ForeignKeys = append(table.ForeignKeys, foreignKey)
		}
	}

	for _, index := range catalog.Indexes {
		table := getTable(getSchema(index.SchemaName), index.TableName)
		table.Indexes = append(table.Indexes, indexNode{
			Name:       index.IndexName,
			Columns:    append([]string(nil), index.Columns...),
			Unique:     index.IsUnique,
			Predicate:  index.Predicate,
			Definition: index.Definition,
		})
	}

	for _, view := range catalog.Views {
		viewNode := getView(getSchema(view.SchemaName), view.ViewName)
		viewNode.Definition = view.Definition
	}

	document := schemaDocument{
		Source:  source,
		Schemas: make([]schemaNode, 0, len(schemas)),
	}

	schemaNames := make([]string, 0, len(schemas))
	for schemaName := range schemas {
		schemaNames = append(schemaNames, schemaName)
	}
	sort.Strings(schemaNames)

	for _, schemaName := range schemaNames {
		builder := schemas[schemaName]

		for tableName := range builder.tables {
			builder.node.Tables = append(builder.node.Tables, *builder.tables[tableName])
		}
		for viewName := range builder.views {
			builder.node.Views = append(builder.node.Views, *builder.views[viewName])
		}

		sort.Slice(builder.node.Tables, func(i, j int) bool {
			return builder.node.Tables[i].Name < builder.node.Tables[j].Name
		})
		sort.Slice(builder.node.Views, func(i, j int) bool {
			return builder.node.Views[i].Name < builder.node.Views[j].Name
		})

		for index := range builder.node.Tables {
			sort.Slice(builder.node.Tables[index].ForeignKeys, func(i, j int) bool {
				return builder.node.Tables[index].ForeignKeys[i].Name < builder.node.Tables[index].ForeignKeys[j].Name
			})
			sort.Slice(builder.node.Tables[index].UniqueConstraints, func(i, j int) bool {
				return builder.node.Tables[index].UniqueConstraints[i].Name < builder.node.Tables[index].UniqueConstraints[j].Name
			})
			sort.Slice(builder.node.Tables[index].CheckConstraints, func(i, j int) bool {
				return builder.node.Tables[index].CheckConstraints[i].Name < builder.node.Tables[index].CheckConstraints[j].Name
			})
			sort.Slice(builder.node.Tables[index].Indexes, func(i, j int) bool {
				return builder.node.Tables[index].Indexes[i].Name < builder.node.Tables[index].Indexes[j].Name
			})
		}

		document.Schemas = append(document.Schemas, builder.node)
	}

	return document
}

func getTable(schema *schemaBuilder, tableName string) *tableNode {
	if table, ok := schema.tables[tableName]; ok {
		return table
	}

	table := &tableNode{Name: tableName}
	schema.tables[tableName] = table
	return table
}

func getView(schema *schemaBuilder, viewName string) *viewNode {
	if view, ok := schema.views[viewName]; ok {
		return view
	}

	view := &viewNode{Name: viewName}
	schema.views[viewName] = view
	return view
}

func foreignKeyAction(code string) string {
	switch strings.TrimSpace(code) {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return ""
	}
}
