package gin

import (
	"testing"

	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/misc"
	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
	type args struct {
		def   misc.QueryDefinition
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    misc.Query
		wantErr bool
	}{
		{
			name: "simple: u integer parse",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorEqual}, misc.DataTypeUInteger), value: "12"},
			want: misc.NewQuery("id", misc.QueryOperatorEqual, misc.NewOperand(uint(12))),
		},
		{
			name: "simple: u integer parse with parentheses",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorEqual}, misc.DataTypeUInteger), value: "(12)"},
			want: misc.NewQuery("id", misc.QueryOperatorEqual, misc.NewOperand(uint(12))),
		},
		{
			name: "multiple: integer in operator",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorContain}, misc.DataTypeUInteger), value: "12,13,14"},
			want: misc.NewQuery("id", misc.QueryOperatorContain, misc.NewOperand([]uint{12, 13, 14})),
		},
		{
			name:    "multiple: integer in operator expect error",
			args:    args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorContain}, misc.DataTypeUInteger), value: "12,13,14,"},
			want:    misc.NewQuery("id", misc.QueryOperatorContain, misc.NewOperand([]uint{12, 13, 14})),
			wantErr: true,
		},
		{
			name: "operand: eq multiple handle",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorEqual}, misc.DataTypeUInteger), value: "eq(12)"},
			want: misc.NewQuery("id", misc.QueryOperatorEqual, misc.NewOperand(uint(12))),
		},
		{
			name: "multiple: integer in operator with parentheses",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorContain}, misc.DataTypeUInteger), value: "(12,13,14)"},
			want: misc.NewQuery("id", misc.QueryOperatorContain, misc.NewOperand([]uint{12, 13, 14})),
		},
		{
			name: "none: should handle empty op",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorEmpty}, misc.DataTypeUInteger), value: "empt()"},
			want: misc.NewEmptyQuery("id", misc.QueryOperatorEmpty),
		},
		{
			name:    "none: expect error where value given",
			args:    args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorEmpty}, misc.DataTypeUInteger), value: "empt(12)"},
			want:    misc.NewEmptyQuery("id", misc.QueryOperatorEmpty),
			wantErr: true,
		},
		{
			name:    "none: expect error where bad value given",
			args:    args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorEmpty}, misc.DataTypeUInteger), value: "empt(12,)"},
			want:    misc.NewEmptyQuery("id", misc.QueryOperatorEmpty),
			wantErr: true,
		},
		{
			name: "more than: simple more than",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorMoreThan}, misc.DataTypeUInteger), value: "mr(12)"},
			want: misc.NewQuery("id", misc.QueryOperatorMoreThan, misc.NewOperand(uint(12))),
		},
		{
			name:    "more than: expect error on multiple param",
			args:    args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorMoreThan}, misc.DataTypeUInteger), value: "mr(12,13,14)"},
			want:    misc.NewQuery("id", misc.QueryOperatorMoreThan, misc.NewOperand([]uint{12, 13, 14})),
			wantErr: true,
		},
		{
			name: "not contain: simple fars",
			args: args{def: misc.NewQueryDefinition("id", []misc.QueryOperator{misc.QueryOperatorContain}, misc.DataTypeString), value: "cn(test,a,b,c)"},
			want: misc.NewQuery("id", misc.QueryOperatorContain, misc.NewOperand([]string{"test", "a", "b", "c"})),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name != "multiple: integer in operator with parentheses" {
				return
			}
			got, err := ParseQuery(tt.args.def, tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			assert.Equal(t, tt.want.GetName(), got.GetName())
			assert.Equal(t, tt.want.GetOperator(), got.GetOperator())
			assert.Equal(t, tt.want.GetOperand(), got.GetOperand())
		})
	}
}
