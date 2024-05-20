package xorm

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestMustIn(t *testing.T) {
	type args struct {
		sql  string
		args []any
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []any
	}{
		{
			name: "test1",
			args: args{
				sql:  "name in (?)",
				args: []any{[]any{"a", "b", "c"}},
			},
			want:  "name in (?, ?, ?)",
			want1: []any{"a", "b", "c"},
		}, {
			name: "test2",
			args: args{
				sql:  "name in (?) and id x in (?)",
				args: []any{[]any{"a", "b", "c"}, []any{1, 2}},
			},
			want:  "name in (?, ?, ?) and id x in (?, ?)",
			want1: []any{"a", "b", "c", 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := MustIn(tt.args.sql, tt.args.args...)
			if got != tt.want {
				t.Errorf("MustIn() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("MustIn() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}

	t.Run("panic", func(t *testing.T) {
		var err error
		func() {
			defer func() {
				err = recover().(error)
			}()
			_, _ = MustIn("name in (?) and x in (?)", []any{[]any{"a", "b", "c"}})
		}()
		assert.Error(t, err)
	})
}

func TestMustNamed(t *testing.T) {
	type args struct {
		inputSql string
		arg      any
	}
	tests := []struct {
		name     string
		args     args
		wantSql  string
		wantArgs []any
	}{
		{
			name: "test1",
			args: args{
				inputSql: "name = :name",
				arg:      map[string]any{"name": "dapeng"},
			},
			wantSql:  "name = ?",
			wantArgs: []any{"dapeng"},
		},
		{
			name: "test2",
			args: args{
				inputSql: "name = :name and id = :id and status in (:status)",
				arg: map[string]any{
					"name":   "dapeng",
					"id":     1,
					"status": []any{1, 2},
				},
			},
			wantSql:  "name = ? and id = ? and status in (?, ?)",
			wantArgs: []any{"dapeng", 1, 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSql, gotArgs := MustNamed(tt.args.inputSql, tt.args.arg)
			if gotSql != tt.wantSql {
				t.Errorf("MustNamed() gotSql = %v, want %v", gotSql, tt.wantSql)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("MustNamed() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}

	t.Run("panic", func(t *testing.T) {
		var err error
		func() {
			defer func() {
				info := recover()
				err = info.(error)
			}()
			_, _ = MustNamed("name in (:name) and x in (?)", []any{map[string]any{"a": 1, "b": 2, "c": 3}})
		}()
		assert.Error(t, err)
	})
}

func Test_sqlDeal(t *testing.T) {
	type args struct {
		sql  string
		args []any
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 []any
	}{
		{
			name: "test1",
			args: args{
				sql:  "name in (?)",
				args: []any{[]any{"dapeng"}},
			},
			want:  "name in (?)",
			want1: []any{"dapeng"},
		},
		{
			name: "test2",
			args: args{
				sql: "name in (:name) and id x in (:ids)",
				args: []any{
					NameMap{
						"name": "dapeng",
						"ids":  []int{1, 2},
					},
				},
			},
			want:  "name in (?) and id x in (?, ?)",
			want1: []any{"dapeng", 1, 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := sqlDeal(tt.args.sql, tt.args.args...)
			assert.Equalf(t, tt.want, got, "sqlDeal(%v, %v...)", tt.args.sql, tt.args.args)
			assert.Equalf(t, tt.want1, got1, "sqlDeal(%v, %v...)", tt.args.sql, tt.args.args)
		})
	}
}
