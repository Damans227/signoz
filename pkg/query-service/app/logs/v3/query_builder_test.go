package v3

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"go.signoz.io/signoz/pkg/query-service/constants"
	v3 "go.signoz.io/signoz/pkg/query-service/model/v3"
)

var testGetClickhouseColumnNameData = []struct {
	Name               string
	AttributeKey       v3.AttributeKey
	ExpectedColumnName string
}{
	{
		Name:               "attribute",
		AttributeKey:       v3.AttributeKey{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
		ExpectedColumnName: "attributes_string_value[indexOf(attributes_string_key, 'user_name')]",
	},
	{
		Name:               "resource",
		AttributeKey:       v3.AttributeKey{Key: "servicename", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource},
		ExpectedColumnName: "resources_string_value[indexOf(resources_string_key, 'servicename')]",
	},
	{
		Name:               "selected field",
		AttributeKey:       v3.AttributeKey{Key: "servicename", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag, IsColumn: true},
		ExpectedColumnName: "servicename",
	},
	{
		Name:               "same name as top level column",
		AttributeKey:       v3.AttributeKey{Key: "trace_id", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
		ExpectedColumnName: "attributes_string_value[indexOf(attributes_string_key, 'trace_id')]",
	},
	{
		Name:               "top level column",
		AttributeKey:       v3.AttributeKey{Key: "trace_id", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag, IsColumn: true},
		ExpectedColumnName: "trace_id",
	},
}

func TestGetClickhouseColumnName(t *testing.T) {
	for _, tt := range testGetClickhouseColumnNameData {
		Convey("testGetClickhouseColumnNameData", t, func() {
			columnName := getClickhouseColumnName(tt.AttributeKey)
			So(columnName, ShouldEqual, tt.ExpectedColumnName)
		})
	}
}

var testGetSelectLabelsData = []struct {
	Name              string
	AggregateOperator v3.AggregateOperator
	GroupByTags       []v3.AttributeKey
	SelectLabels      string
}{
	{
		Name:              "select fields for groupBy attribute",
		AggregateOperator: v3.AggregateOperatorCount,
		GroupByTags:       []v3.AttributeKey{{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
		SelectLabels:      " attributes_string_value[indexOf(attributes_string_key, 'user_name')] as user_name,",
	},
	{
		Name:              "select fields for groupBy resource",
		AggregateOperator: v3.AggregateOperatorCount,
		GroupByTags:       []v3.AttributeKey{{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource}},
		SelectLabels:      " resources_string_value[indexOf(resources_string_key, 'user_name')] as user_name,",
	},
	{
		Name:              "select fields for groupBy attribute and resource",
		AggregateOperator: v3.AggregateOperatorCount,
		GroupByTags: []v3.AttributeKey{
			{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource},
			{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
		},
		SelectLabels: " resources_string_value[indexOf(resources_string_key, 'user_name')] as user_name, attributes_string_value[indexOf(attributes_string_key, 'host')] as host,",
	},
	{
		Name:              "select fields for groupBy materialized columns",
		AggregateOperator: v3.AggregateOperatorCount,
		GroupByTags:       []v3.AttributeKey{{Key: "host", IsColumn: true}},
		SelectLabels:      " host as host,",
	},
	{
		Name:              "trace_id field as an attribute",
		AggregateOperator: v3.AggregateOperatorCount,
		GroupByTags:       []v3.AttributeKey{{Key: "trace_id", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
		SelectLabels:      " attributes_string_value[indexOf(attributes_string_key, 'trace_id')] as trace_id,",
	},
}

func TestGetSelectLabels(t *testing.T) {
	for _, tt := range testGetSelectLabelsData {
		Convey("testGetSelectLabelsData", t, func() {
			selectLabels := getSelectLabels(tt.AggregateOperator, tt.GroupByTags)
			So(selectLabels, ShouldEqual, tt.SelectLabels)
		})
	}
}

var timeSeriesFilterQueryData = []struct {
	Name           string
	FilterSet      *v3.FilterSet
	GroupBy        []v3.AttributeKey
	ExpectedFilter string
	Fields         map[string]v3.AttributeKey
	Error          string
}{
	{
		Name: "Test attribute and resource attribute",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "john", Operator: "="},
			{Key: v3.AttributeKey{Key: "k8s_namespace", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource}, Value: "my_service", Operator: "!="},
		}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'user_name')] = 'john' AND resources_string_value[indexOf(resources_string_key, 'k8s_namespace')] != 'my_service'",
	},
	{
		Name: "Test materialized column",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag, IsColumn: true}, Value: "john", Operator: "="},
			{Key: v3.AttributeKey{Key: "k8s_namespace", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource}, Value: "my_service", Operator: "!="},
		}},
		ExpectedFilter: " AND user_name = 'john' AND resources_string_value[indexOf(resources_string_key, 'k8s_namespace')] != 'my_service'",
	},
	{
		Name: "Test like",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "102.%", Operator: "like"},
		}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'host')] ILIKE '102.%'",
	},
	{
		Name: "Test IN",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "bytes", DataType: v3.AttributeKeyDataTypeFloat64, Type: v3.AttributeKeyTypeTag}, Value: []interface{}{1, 2, 3, 4}, Operator: "in"},
		}},
		ExpectedFilter: " AND attributes_float64_value[indexOf(attributes_float64_key, 'bytes')] IN [1,2,3,4]",
	},
	{
		Name: "Test DataType int64",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "bytes", DataType: v3.AttributeKeyDataTypeInt64, Type: v3.AttributeKeyTypeTag}, Value: 10, Operator: ">"},
		}},
		ExpectedFilter: " AND attributes_int64_value[indexOf(attributes_int64_key, 'bytes')] > 10",
	},
	{
		Name: "Test NOT IN",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: []interface{}{"john", "bunny"}, Operator: "nin"},
		}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'name')] NOT IN ['john','bunny']",
	},
	{
		Name: "Test exists",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "bytes", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "", Operator: "exists"},
		}},
		ExpectedFilter: " AND has(attributes_string_key, 'bytes')",
	},
	{
		Name: "Test not exists",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "bytes", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "", Operator: "nexists"},
		}},
		ExpectedFilter: " AND not has(attributes_string_key, 'bytes')",
	},
	{
		Name: "Test contains",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "102.", Operator: "contains"},
		}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'host')] ILIKE '%102.%'",
	},
	{
		Name: "Test not contains",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "102.", Operator: "ncontains"},
		}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'host')] NOT ILIKE '%102.%'",
	},
	{
		Name: "Test groupBy",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "102.", Operator: "ncontains"},
		}},
		GroupBy:        []v3.AttributeKey{{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'host')] NOT ILIKE '%102.%' AND indexOf(attributes_string_key, 'host') > 0",
	},
	{
		Name: "Test groupBy isColumn",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "102.", Operator: "ncontains"},
		}},
		GroupBy:        []v3.AttributeKey{{Key: "host", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag, IsColumn: true}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'host')] NOT ILIKE '%102.%'",
	},
	{
		Name: "Wrong data",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "bytes", Type: v3.AttributeKeyTypeTag, DataType: v3.AttributeKeyDataTypeFloat64}, Value: true, Operator: "="},
		}},
		Error: "failed to validate and cast value for bytes: invalid data type, expected float, got bool",
	},
	{
		Name: "Test top level field with metadata",
		FilterSet: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
			{Key: v3.AttributeKey{Key: "body", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "%test%", Operator: "like"},
		}},
		ExpectedFilter: " AND attributes_string_value[indexOf(attributes_string_key, 'body')] ILIKE '%test%'",
	},
}

func TestBuildLogsTimeSeriesFilterQuery(t *testing.T) {
	for _, tt := range timeSeriesFilterQueryData {
		Convey("TestBuildLogsTimeSeriesFilterQuery", t, func() {
			query, err := buildLogsTimeSeriesFilterQuery(tt.FilterSet, tt.GroupBy)
			if tt.Error != "" {
				So(err.Error(), ShouldEqual, tt.Error)
			} else {
				So(err, ShouldBeNil)
				So(query, ShouldEqual, tt.ExpectedFilter)
			}

		})
	}
}

var testBuildLogsQueryData = []struct {
	Name              string
	PanelType         v3.PanelType
	Start             int64
	End               int64
	Step              int64
	BuilderQuery      *v3.BuilderQuery
	GroupByTags       []v3.AttributeKey
	TableName         string
	AggregateOperator v3.AggregateOperator
	ExpectedQuery     string
	Type              int
	PreferRPM         bool
}{
	{
		Name:      "Test aggregate count on select field",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorCount,
			Expression:        "A",
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) group by ts order by value DESC",
	},
	{
		Name:      "Test aggregate count on a attribute",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCount,
			Expression:         "A",
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND has(attributes_string_key, 'user_name') group by ts order by value DESC",
	},
	{
		Name:      "Test aggregate count on a with filter",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "user_name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCount,
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "bytes", DataType: v3.AttributeKeyDataTypeFloat64, Type: v3.AttributeKeyTypeTag}, Value: 100, Operator: ">"},
			}},
			Expression: "A",
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_float64_value[indexOf(attributes_float64_key, 'bytes')] > 100.000000 AND has(attributes_string_key, 'user_name') group by ts order by value DESC",
	},
	{
		Name:      "Test aggregate count distinct and order by value",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			OrderBy:            []v3.OrderBy{{ColumnName: "#SIGNOZ_VALUE", Order: "ASC"}},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(distinct(name))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) group by ts order by value ASC",
	},
	{
		Name:      "Test aggregate count distinct on non selected field",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) group by ts order by value DESC",
	},
	{
		Name:      "Test aggregate count distinct with filter and groupBy",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
				{Key: v3.AttributeKey{Key: "x", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource}, Value: "abc", Operator: "!="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}, {ColumnName: "ts", Order: "ASC", Key: "ts", IsColumn: true}},
		},
		TableName: "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts," +
			" attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"toFloat64(count(distinct(name))) as value from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' AND resources_string_value[indexOf(resources_string_key, 'x')] != 'abc' " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test aggregate count with multiple filter,groupBy and orderBy",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
				{Key: v3.AttributeKey{Key: "x", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource}, Value: "abc", Operator: "!="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, {Key: "x", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeResource}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}, {ColumnName: "x", Order: "ASC"}},
		},
		TableName: "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts," +
			" attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"resources_string_value[indexOf(resources_string_key, 'x')] as x, " +
			"toFloat64(count(distinct(name))) as value from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' AND resources_string_value[indexOf(resources_string_key, 'x')] != 'abc' " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"AND indexOf(resources_string_key, 'x') > 0 " +
			"group by method,x,ts " +
			"order by method ASC,x ASC",
	},
	{
		Name:      "Test aggregate avg",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", DataType: v3.AttributeKeyDataTypeFloat64, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorAvg,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}, {ColumnName: "x", Order: "ASC", Key: "x", IsColumn: true}},
		},
		TableName: "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts," +
			" attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"avg(attributes_float64_value[indexOf(attributes_float64_key, 'bytes')]) as value " +
			"from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test aggregate sum",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorSum,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName: "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts," +
			" attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"sum(bytes) as value " +
			"from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test aggregate min",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorMin,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName: "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts," +
			" attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"min(bytes) as value " +
			"from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test aggregate max",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorMax,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName: "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts," +
			" attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"max(bytes) as value " +
			"from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test aggregate PXX",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorP05,
			Expression:         "A",
			Filters:            &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			GroupBy:            []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy:            []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName: "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts," +
			" attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"quantile(0.05)(bytes) as value " +
			"from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test aggregate RateSum",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", IsColumn: true},
			AggregateOperator:  v3.AggregateOperatorRateSum,
			Expression:         "A",
			Filters:            &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			GroupBy:            []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy:            []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName: "logs",
		PreferRPM: true,
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, attributes_string_value[indexOf(attributes_string_key, 'method')] as method" +
			", sum(bytes)/1.000000 as value from signoz_logs.distributed_logs " +
			"where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts order by method ASC",
	},
	{
		Name:      "Test aggregate rate",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", Type: v3.AttributeKeyTypeTag, DataType: v3.AttributeKeyDataTypeFloat64},
			AggregateOperator:  v3.AggregateOperatorRate,
			Expression:         "A",
			Filters:            &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			GroupBy:            []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy:            []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName: "logs",
		PreferRPM: false,
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, attributes_string_value[indexOf(attributes_string_key, 'method')] as method" +
			", count(attributes_float64_value[indexOf(attributes_float64_key, 'bytes')])/60.000000 as value " +
			"from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test aggregate RateSum without materialized column",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "bytes", Type: v3.AttributeKeyTypeTag, DataType: v3.AttributeKeyDataTypeFloat64},
			AggregateOperator:  v3.AggregateOperatorRateSum,
			Expression:         "A",
			Filters:            &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			GroupBy:            []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy:            []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName: "logs",
		PreferRPM: true,
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, " +
			"attributes_string_value[indexOf(attributes_string_key, 'method')] as method, " +
			"sum(attributes_float64_value[indexOf(attributes_float64_key, 'bytes')])/1.000000 as value " +
			"from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) " +
			"AND indexOf(attributes_string_key, 'method') > 0 " +
			"group by method,ts " +
			"order by method ASC",
	},
	{
		Name:      "Test Noop",
		PanelType: v3.PanelTypeList,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			SelectColumns:     []v3.AttributeKey{},
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters:           &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
		},
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string," +
			"CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64," +
			"CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string " +
			"from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) order by timestamp DESC",
	},
	{
		Name:      "Test Noop order by custom",
		PanelType: v3.PanelTypeList,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			SelectColumns:     []v3.AttributeKey{},
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters:           &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			OrderBy:           []v3.OrderBy{{ColumnName: "method", DataType: v3.AttributeKeyDataTypeString, Order: "ASC", IsColumn: true}},
		},
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string," +
			"CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64," +
			"CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string " +
			"from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) order by method ASC",
	},
	{
		Name:      "Test Noop with filter",
		PanelType: v3.PanelTypeList,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			SelectColumns:     []v3.AttributeKey{},
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "severity_number", DataType: v3.AttributeKeyDataTypeInt64, IsColumn: true}, Operator: "!=", Value: 0},
			}},
		},
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string," +
			"CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64," +
			"CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string " +
			"from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND severity_number != 0 order by timestamp DESC",
	},
	{
		Name:      "Test aggregate with having clause",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Having: []v3.Having{
				{
					ColumnName: "name",
					Operator:   ">",
					Value:      10,
				},
			},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) group by ts having value > 10 order by value DESC",
	},
	{
		Name:      "Test aggregate with having clause and filters",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			Having: []v3.Having{
				{
					ColumnName: "name",
					Operator:   ">",
					Value:      10,
				},
			},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' group by ts having value > 10 order by value DESC",
	},
	{
		Name:      "Test top level key",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "body", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeUnspecified, IsColumn: true}, Value: "%test%", Operator: "like"},
			},
			},
			Having: []v3.Having{
				{
					ColumnName: "name",
					Operator:   ">",
					Value:      10,
				},
			},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND body ILIKE '%test%' group by ts having value > 10 order by value DESC",
	},
	{
		Name:      "Test attribute with same name as top level key",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "body", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "%test%", Operator: "like"},
			},
			},
			Having: []v3.Having{
				{
					ColumnName: "name",
					Operator:   ">",
					Value:      10,
				},
			},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 60 SECOND) AS ts, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_string_value[indexOf(attributes_string_key, 'body')] ILIKE '%test%' group by ts having value > 10 order by value DESC",
	},

	// Tests for table panel type
	{
		Name:      "TABLE: Test count",
		PanelType: v3.PanelTypeTable,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorCount,
			Expression:        "A",
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT now() as ts, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) order by value DESC",
	},
	{
		Name:      "TABLE: Test count with groupBy",
		PanelType: v3.PanelTypeTable,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorCount,
			Expression:        "A",
			GroupBy: []v3.AttributeKey{
				{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT now() as ts, attributes_string_value[indexOf(attributes_string_key, 'name')] as name, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND indexOf(attributes_string_key, 'name') > 0 group by name order by value DESC",
	},
	{
		Name:      "TABLE: Test count with groupBy, orderBy",
		PanelType: v3.PanelTypeTable,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorCount,
			Expression:        "A",
			GroupBy: []v3.AttributeKey{
				{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			},
			OrderBy: []v3.OrderBy{
				{ColumnName: "name", Order: "DESC"},
			},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT now() as ts, attributes_string_value[indexOf(attributes_string_key, 'name')] as name, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND indexOf(attributes_string_key, 'name') > 0 group by name order by name DESC",
	},
}

func TestBuildLogsQuery(t *testing.T) {
	for _, tt := range testBuildLogsQueryData {
		Convey("TestBuildLogsQuery", t, func() {
			query, err := buildLogsQuery(tt.PanelType, tt.Start, tt.End, tt.Step, tt.BuilderQuery, "", tt.PreferRPM)
			So(err, ShouldBeNil)
			So(query, ShouldEqual, tt.ExpectedQuery)

		})
	}
}

var testGetZerosForEpochNanoData = []struct {
	Name       string
	Epoch      int64
	Multiplier int64
	Result     int64
}{
	{
		Name:       "Test 1",
		Epoch:      1680712080000,
		Multiplier: 1000000,
		Result:     1680712080000000000,
	},
	{
		Name:       "Test 1",
		Epoch:      1680712080000000000,
		Multiplier: 1,
		Result:     1680712080000000000,
	},
}

func TestGetZerosForEpochNano(t *testing.T) {
	for _, tt := range testGetZerosForEpochNanoData {
		Convey("testGetZerosForEpochNanoData", t, func() {
			multiplier := getZerosForEpochNano(tt.Epoch)
			So(multiplier, ShouldEqual, tt.Multiplier)
			So(tt.Epoch*multiplier, ShouldEqual, tt.Result)
		})
	}
}

var testOrderBy = []struct {
	Name      string
	PanelType v3.PanelType
	Items     []v3.OrderBy
	Tags      []v3.AttributeKey
	Result    string
}{
	{
		Name:      "Test 1",
		PanelType: v3.PanelTypeGraph,
		Items: []v3.OrderBy{
			{
				ColumnName: "name",
				Order:      "asc",
			},
			{
				ColumnName: constants.SigNozOrderByValue,
				Order:      "desc",
			},
		},
		Tags: []v3.AttributeKey{
			{Key: "name"},
		},
		Result: "name asc,value desc",
	},
	{
		Name:      "Test 2",
		PanelType: v3.PanelTypeGraph,
		Items: []v3.OrderBy{
			{
				ColumnName: "name",
				Order:      "asc",
			},
			{
				ColumnName: "bytes",
				Order:      "asc",
			},
		},
		Tags: []v3.AttributeKey{
			{Key: "name"},
			{Key: "bytes"},
		},
		Result: "name asc,bytes asc",
	},
	{
		Name:      "Test Graph item not present in tag",
		PanelType: v3.PanelTypeGraph,
		Items: []v3.OrderBy{
			{
				ColumnName: "name",
				Order:      "asc",
			},
			{
				ColumnName: "bytes",
				Order:      "asc",
			},
			{
				ColumnName: "method",
				Order:      "asc",
			},
		},
		Tags: []v3.AttributeKey{
			{Key: "name"},
			{Key: "bytes"},
		},
		Result: "name asc,bytes asc",
	},
	{
		Name:      "Test 3",
		PanelType: v3.PanelTypeList,
		Items: []v3.OrderBy{
			{
				ColumnName: "name",
				Order:      "asc",
			},
			{
				ColumnName: constants.SigNozOrderByValue,
				Order:      "asc",
			},
			{
				ColumnName: "bytes",
				Order:      "asc",
			},
		},
		Tags: []v3.AttributeKey{
			{Key: "name"},
			{Key: "bytes"},
		},
		Result: "name asc,value asc,bytes asc",
	},
	{
		Name:      "Test 4",
		PanelType: v3.PanelTypeList,
		Items: []v3.OrderBy{
			{
				ColumnName: "name",
				Order:      "asc",
			},
			{
				ColumnName: constants.SigNozOrderByValue,
				Order:      "asc",
			},
			{
				ColumnName: "bytes",
				Order:      "asc",
			},
			{
				ColumnName: "response_time",
				Order:      "desc",
				Key:        "response_time",
				Type:       v3.AttributeKeyTypeTag,
				DataType:   v3.AttributeKeyDataTypeString,
			},
		},
		Tags: []v3.AttributeKey{
			{Key: "name"},
			{Key: "bytes"},
		},
		Result: "name asc,value asc,bytes asc,attributes_string_value[indexOf(attributes_string_key, 'response_time')] desc",
	},
}

func TestOrderBy(t *testing.T) {
	for _, tt := range testOrderBy {
		Convey("testOrderBy", t, func() {
			res := orderByAttributeKeyTags(tt.PanelType, tt.Items, tt.Tags)
			So(res, ShouldResemble, tt.Result)
		})
	}
}

// if there is no group by then there is no point of limit in ts and table queries
// since the above will result in a single ts

// handle only when there is a group by something.

var testPrepLogsQueryData = []struct {
	Name              string
	PanelType         v3.PanelType
	Start             int64
	End               int64
	Step              int64
	BuilderQuery      *v3.BuilderQuery
	GroupByTags       []v3.AttributeKey
	TableName         string
	AggregateOperator v3.AggregateOperator
	ExpectedQuery     string
	Options           Options
}{
	{
		Name:      "Test TS with limit- first",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			Limit:   10,
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT method from (SELECT attributes_string_value[indexOf(attributes_string_key, 'method')] as method, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' AND indexOf(attributes_string_key, 'method') > 0 group by method order by value DESC) LIMIT 10",
		Options:       Options{GraphLimitQtype: constants.FirstQueryGraphLimit, PreferRPM: true},
	},
	{
		Name:      "Test TS with limit- first - with order by value",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			Limit:   10,
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: constants.SigNozOrderByValue, Order: "ASC"}},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT method from (SELECT attributes_string_value[indexOf(attributes_string_key, 'method')] as method, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' AND indexOf(attributes_string_key, 'method') > 0 group by method order by value ASC) LIMIT 10",
		Options:       Options{GraphLimitQtype: constants.FirstQueryGraphLimit, PreferRPM: true},
	},
	{
		Name:      "Test TS with limit- first - with order by attribute",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			Limit:   10,
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT method from (SELECT attributes_string_value[indexOf(attributes_string_key, 'method')] as method, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' AND indexOf(attributes_string_key, 'method') > 0 group by method order by method ASC) LIMIT 10",
		Options:       Options{GraphLimitQtype: constants.FirstQueryGraphLimit, PreferRPM: true},
	},
	{
		Name:      "Test TS with limit- second",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			Limit:   2,
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 0 SECOND) AS ts, attributes_string_value[indexOf(attributes_string_key, 'method')] as method, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' AND indexOf(attributes_string_key, 'method') > 0 AND (method) GLOBAL IN (%s) group by method,ts order by value DESC",
		Options:       Options{GraphLimitQtype: constants.SecondQueryGraphLimit},
	},
	{
		Name:      "Test TS with limit- second - with order by",
		PanelType: v3.PanelTypeGraph,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:          "A",
			AggregateAttribute: v3.AttributeKey{Key: "name", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag},
			AggregateOperator:  v3.AggregateOperatorCountDistinct,
			Expression:         "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
			GroupBy: []v3.AttributeKey{{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			OrderBy: []v3.OrderBy{{ColumnName: "method", Order: "ASC"}},
			Limit:   2,
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT toStartOfInterval(fromUnixTimestamp64Nano(timestamp), INTERVAL 0 SECOND) AS ts, attributes_string_value[indexOf(attributes_string_key, 'method')] as method, toFloat64(count(distinct(attributes_string_value[indexOf(attributes_string_key, 'name')]))) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET' AND indexOf(attributes_string_key, 'method') > 0 AND (method) GLOBAL IN (%s) group by method,ts order by method ASC",
		Options:       Options{GraphLimitQtype: constants.SecondQueryGraphLimit},
	},
	// Live tail
	{
		Name:      "Live Tail Query",
		PanelType: v3.PanelTypeList,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}, Value: "GET", Operator: "="},
			},
			},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string,CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64,CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string from signoz_logs.distributed_logs where %s  AND attributes_string_value[indexOf(attributes_string_key, 'method')] = 'GET'",
		Options:       Options{IsLivetailQuery: true},
	},
	{
		Name:      "Live Tail Query W/O filter",
		PanelType: v3.PanelTypeList,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters:           &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string,CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64,CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string from signoz_logs.distributed_logs where %s",
		Options:       Options{IsLivetailQuery: true},
	},
	{
		Name:      "Table query w/o limit",
		PanelType: v3.PanelTypeTable,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorCount,
			Expression:        "A",
			Filters:           &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT now() as ts, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) order by value DESC",
		Options:       Options{},
	},
	{
		Name:      "Table query with limit",
		PanelType: v3.PanelTypeTable,
		Start:     1680066360726210000,
		End:       1680066458000000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorCount,
			Expression:        "A",
			Filters:           &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			Limit:             10,
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT now() as ts, toFloat64(count(*)) as value from signoz_logs.distributed_logs where (timestamp >= 1680066360726210000 AND timestamp <= 1680066458000000000) order by value DESC LIMIT 10",
		Options:       Options{},
	},
}

func TestPrepareLogsQuery(t *testing.T) {
	for _, tt := range testPrepLogsQueryData {
		Convey("TestBuildLogsQuery", t, func() {
			query, err := PrepareLogsQuery(tt.Start, tt.End, "", tt.PanelType, tt.BuilderQuery, tt.Options)
			So(err, ShouldBeNil)
			So(query, ShouldEqual, tt.ExpectedQuery)

		})
	}
}

var testPrepLogsQueryLimitOffsetData = []struct {
	Name              string
	PanelType         v3.PanelType
	Start             int64
	End               int64
	Step              int64
	BuilderQuery      *v3.BuilderQuery
	GroupByTags       []v3.AttributeKey
	TableName         string
	AggregateOperator v3.AggregateOperator
	ExpectedQuery     string
	Options           Options
}{
	{
		Name:      "Test limit less than pageSize - order by ts",
		PanelType: v3.PanelTypeList,
		Start:     1680518666000000000,
		End:       1691618704365000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters:           &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			OrderBy:           []v3.OrderBy{{ColumnName: constants.TIMESTAMP, Order: "desc", Key: constants.TIMESTAMP, DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeUnspecified, IsColumn: true}},
			Limit:             1,
			Offset:            0,
			PageSize:          5,
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string,CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64,CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string from signoz_logs.distributed_logs where (timestamp >= 1680518666000000000 AND timestamp <= 1691618704365000000) order by timestamp desc LIMIT 1",
	},
	{
		Name:      "Test limit greater than pageSize - order by ts",
		PanelType: v3.PanelTypeList,
		Start:     1680518666000000000,
		End:       1691618704365000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "id", Type: v3.AttributeKeyTypeUnspecified, DataType: v3.AttributeKeyDataTypeString, IsColumn: true}, Operator: v3.FilterOperatorLessThan, Value: "2TNh4vp2TpiWyLt3SzuadLJF2s4"},
			}},
			OrderBy:  []v3.OrderBy{{ColumnName: constants.TIMESTAMP, Order: "desc", Key: constants.TIMESTAMP, DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeUnspecified, IsColumn: true}},
			Limit:    100,
			Offset:   10,
			PageSize: 10,
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string,CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64,CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string from signoz_logs.distributed_logs where (timestamp >= 1680518666000000000 AND timestamp <= 1691618704365000000) AND id < '2TNh4vp2TpiWyLt3SzuadLJF2s4' order by timestamp desc LIMIT 10",
	},
	{
		Name:      "Test limit less than pageSize  - order by custom",
		PanelType: v3.PanelTypeList,
		Start:     1680518666000000000,
		End:       1691618704365000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters:           &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{}},
			OrderBy:           []v3.OrderBy{{ColumnName: "method", Order: "desc", Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			Limit:             1,
			Offset:            0,
			PageSize:          5,
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string,CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64,CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string from signoz_logs.distributed_logs where (timestamp >= 1680518666000000000 AND timestamp <= 1691618704365000000) order by attributes_string_value[indexOf(attributes_string_key, 'method')] desc LIMIT 1 OFFSET 0",
	},
	{
		Name:      "Test limit greater than pageSize - order by custom",
		PanelType: v3.PanelTypeList,
		Start:     1680518666000000000,
		End:       1691618704365000000,
		Step:      60,
		BuilderQuery: &v3.BuilderQuery{
			QueryName:         "A",
			AggregateOperator: v3.AggregateOperatorNoOp,
			Expression:        "A",
			Filters: &v3.FilterSet{Operator: "AND", Items: []v3.FilterItem{
				{Key: v3.AttributeKey{Key: "id", Type: v3.AttributeKeyTypeUnspecified, DataType: v3.AttributeKeyDataTypeString, IsColumn: true}, Operator: v3.FilterOperatorLessThan, Value: "2TNh4vp2TpiWyLt3SzuadLJF2s4"},
			}},
			OrderBy:  []v3.OrderBy{{ColumnName: "method", Order: "desc", Key: "method", DataType: v3.AttributeKeyDataTypeString, Type: v3.AttributeKeyTypeTag}},
			Limit:    100,
			Offset:   50,
			PageSize: 50,
		},
		TableName:     "logs",
		ExpectedQuery: "SELECT timestamp, id, trace_id, span_id, trace_flags, severity_text, severity_number, body,CAST((attributes_string_key, attributes_string_value), 'Map(String, String)') as  attributes_string,CAST((attributes_int64_key, attributes_int64_value), 'Map(String, Int64)') as  attributes_int64,CAST((attributes_float64_key, attributes_float64_value), 'Map(String, Float64)') as  attributes_float64,CAST((resources_string_key, resources_string_value), 'Map(String, String)') as resources_string from signoz_logs.distributed_logs where (timestamp >= 1680518666000000000 AND timestamp <= 1691618704365000000) AND id < '2TNh4vp2TpiWyLt3SzuadLJF2s4' order by attributes_string_value[indexOf(attributes_string_key, 'method')] desc LIMIT 50 OFFSET 50",
	},
}

func TestPrepareLogsQueryLimitOffset(t *testing.T) {
	for _, tt := range testPrepLogsQueryLimitOffsetData {
		Convey("TestBuildLogsQuery", t, func() {
			query, err := PrepareLogsQuery(tt.Start, tt.End, "", tt.PanelType, tt.BuilderQuery, tt.Options)
			So(err, ShouldBeNil)
			So(query, ShouldEqual, tt.ExpectedQuery)

		})
	}
}
