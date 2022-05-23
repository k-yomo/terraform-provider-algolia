package algoliautil

import (
	"reflect"
	"testing"
)

func TestIndexExistsInReplicas(t *testing.T) {
	t.Parallel()

	type args struct {
		replicas  []string
		indexName string
		isVirtual bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if target replica exists",
			args: args{
				replicas:  []string{"abc", "def", "virtual(ghi)"},
				indexName: "abc",
				isVirtual: false,
			},
			want: true,
		},
		{
			name: "returns true if target virtual replica exists",
			args: args{
				replicas:  []string{"abc", "def", "virtual(ghi)"},
				indexName: "ghi",
				isVirtual: true,
			},
			want: true,
		},
		{
			name: "returns false if target replica doesn't exist",
			args: args{
				replicas:  []string{"abc", "def", "virtual(ghi)"},
				indexName: "ghi",
				isVirtual: false,
			},
			want: false,
		},
		{
			name: "returns false if target virtual replica doesn't exist",
			args: args{
				replicas:  []string{"abc", "def", "virtual(ghi)"},
				indexName: "abc",
				isVirtual: true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IndexExistsInReplicas(tt.args.replicas, tt.args.indexName, tt.args.isVirtual); got != tt.want {
				t.Errorf("IndexExistsInReplicas() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_RemoveIndexFromReplicas(t *testing.T) {
	t.Parallel()

	type args struct {
		replicas  []string
		indexName string
		isVirtual bool
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "returns replica list excluding the target replica",
			args: args{
				replicas:  []string{"target", "virtual(abc)", "virtual(target)"},
				indexName: "target",
				isVirtual: false,
			},
			want: []string{"virtual(abc)", "virtual(target)"},
		},
		{
			name: "returns replica list excluding the target virtual replica",
			args: args{
				replicas:  []string{"target", "virtual(abc)", "virtual(target)"},
				indexName: "target",
				isVirtual: true,
			},
			want: []string{"target", "virtual(abc)"},
		},
		{
			name: "returns original replica list if there is no target replica",
			args: args{
				replicas:  []string{"abc", "def", "virtual(target)"},
				indexName: "target",
				isVirtual: false,
			},
			want: []string{"abc", "def", "virtual(target)"},
		},
		{
			name: "returns original replica list if there is no target virtual replica",
			args: args{
				replicas:  []string{"virtual(abc)", "virtual(def)", "target", "virtual(target"},
				indexName: "target",
				isVirtual: true,
			},
			want: []string{"virtual(abc)", "virtual(def)", "target", "virtual(target"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveIndexFromReplicas(tt.args.replicas, tt.args.indexName, tt.args.isVirtual); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeVirtualIndexFromReplicaList() = %v, want %v", got, tt.want)
			}
		})
	}
}
