package schema

import (
	"testing"

	"github.com/effective-security/porto/x/fileutil"
	"github.com/effective-security/xdb/mocks/mockschema"
	"github.com/effective-security/xdb/pkg/cli/clisuite"
	dbschema "github.com/effective-security/xdb/schema"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/suite"
)

type testSuite struct {
	clisuite.TestSuite
}

func TestSchema(t *testing.T) {
	suite.Run(t, new(testSuite))
}

func (s *testSuite) TestPrintColumnsCmd() {
	require := s.Require()

	ctrl := gomock.NewController(s.T())
	mock := mockschema.NewMockProvider(ctrl)
	s.Ctl.WithSchemaProvider(mock)

	res := dbschema.Tables{
		{
			Name:   "test",
			Schema: "dbo",
			Columns: dbschema.Columns{
				{
					Name:     "ID",
					Type:     "uint64",
					UdtType:  "int8",
					Nullable: "NO",
				},
			},
		},
	}

	mock.EXPECT().ListTables(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).Times(2)
	mock.EXPECT().ListTables(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.Errorf("query failed")).Times(1)

	cmd := PrintColumnsCmd{
		DB:     "Datahub",
		Schema: "dbo",
		Table:  []string{"Transaction"},
	}

	err := cmd.Run(s.Ctl)
	require.NoError(err)
	s.Equal("Schema: dbo\n"+
		"Table: test\n\n"+
		"  NAME |  TYPE  | UDT  | NULL | MAX | REF  \n"+
		"-------+--------+------+------+-----+------\n"+
		"  ID   | uint64 | int8 | NO   |     |      \n\n", s.Out.String())

	s.Ctl.O = "json"
	s.Out.Reset()

	err = cmd.Run(s.Ctl)
	require.NoError(err)
	s.Equal(
		"[\n  {\n    \"Schema\": \"dbo\",\n    \"Name\": \"test\",\n    \"Columns\": [\n      {\n        \"Name\": \"ID\",\n        \"Type\": \"uint64\",\n        \"UdtType\": \"int8\",\n        \"Nullable\": \"NO\",\n        \"MaxLength\": null\n      }\n    ],\n    \"Indexes\": null,\n    \"PrimaryKey\": null\n  }\n]\n",
		s.Out.String())

	err = cmd.Run(s.Ctl)
	s.EqualError(err, "query failed")
}

func (s *testSuite) TestPrintTablesCmd() {
	require := s.Require()

	ctrl := gomock.NewController(s.T())
	mock := mockschema.NewMockProvider(ctrl)
	s.Ctl.WithSchemaProvider(mock)

	res := dbschema.Tables{
		{
			Name:   "test",
			Schema: "dbo",
			Columns: dbschema.Columns{
				{
					Name:     "ID",
					Type:     "numeric",
					Nullable: "NO",
				},
			},
		},
	}

	mock.EXPECT().ListTables(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).Times(1)
	mock.EXPECT().ListTables(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.Errorf("query failed")).Times(1)

	cmd := PrintTablesCmd{
		DB:     "Datahub",
		Schema: "dbo",
		Table:  []string{"Transaction"},
	}

	err := cmd.Run(s.Ctl)
	require.NoError(err)
	s.Equal("dbo.test\n", s.Out.String())

	err = cmd.Run(s.Ctl)
	s.EqualError(err, "query failed")
}

func (s *testSuite) TestPrintFKCmd() {
	require := s.Require()

	ctrl := gomock.NewController(s.T())
	mock := mockschema.NewMockProvider(ctrl)
	s.Ctl.WithSchemaProvider(mock)

	res := dbschema.ForeignKeys{
		{
			Name:      "FK_1",
			Schema:    "dbo",
			Table:     "from",
			Column:    "col1",
			RefSchema: "dbo",
			RefTable:  "to",
			RefColumn: "col2",
		},
	}

	mock.EXPECT().ListForeignKeys(gomock.Any(), gomock.Any(), gomock.Any()).Return(res, nil).Times(2)
	mock.EXPECT().ListForeignKeys(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.Errorf("query failed")).Times(1)

	cmd := PrintFKCmd{
		DB:     "Datahub",
		Schema: "dbo",
		Table:  []string{"Transaction"},
	}

	err := cmd.Run(s.Ctl)
	require.NoError(err)
	s.Equal(`  NAME | SCHEMA | TABLE | COLUMN | FK SCHEMA | FK TABLE | FK COLUMN  
-------+--------+-------+--------+-----------+----------+------------
  FK_1 | dbo    | from  | col1   | dbo       | to       | col2       

`, s.Out.String())

	s.Ctl.O = "json"
	s.Out.Reset()

	err = cmd.Run(s.Ctl)
	require.NoError(err)
	s.Equal("[\n  {\n    \"Name\": \"FK_1\",\n    \"Schema\": \"dbo\",\n    \"Table\": \"from\",\n    \"Column\": \"col1\",\n    \"RefSchema\": \"dbo\",\n    \"RefTable\": \"to\",\n    \"RefColumn\": \"col2\"\n  }\n]\n", s.Out.String())

	err = cmd.Run(s.Ctl)
	s.EqualError(err, "query failed")
}

func (s *testSuite) TestGenerate() {
	require := s.Require()

	var res dbschema.Tables
	err := fileutil.Unmarshal("testdata/pg_columns.json", &res)
	require.NoError(err)

	cmd := GenerateCmd{
		Package: "model",
		Schema:  "dbo",
		DB:      "testdb",
		Table:   []string{"Transaction"},
	}
	err = cmd.generate(s.Ctl, "org", res)
	require.NoError(err)

	ctrl := gomock.NewController(s.T())
	mock := mockschema.NewMockProvider(ctrl)
	s.Ctl.WithSchemaProvider(mock)

	ret := dbschema.Tables{
		{
			Name:   "test",
			Schema: "dbo",
			Columns: dbschema.Columns{
				{
					Name:     "ID",
					Type:     "numeric",
					Nullable: "NO",
				},
			},
		},
	}

	mock.EXPECT().ListTables(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(ret, nil).Times(1)
	err = cmd.Run(s.Ctl)
	require.NoError(err)
	s.HasText("DO NOT EDIT!", s.Out.String())
}
