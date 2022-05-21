package algoliautil

import "testing"

func TestIsValidRegion(t *testing.T) {
	t.Parallel()

	type args struct {
		r string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "returns true if 'us' region",
			args: args{r: "us"},
			want: true,
		},
		{
			name: "returns true if 'eu' region",
			args: args{r: "eu"},
			want: true,
		},
		{
			name: "returns true if 'de' region",
			args: args{r: "de"},
			want: true,
		},
		{
			name: "returns false if invalid region",
			args: args{r: "invalid"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidRegion(tt.args.r); got != tt.want {
				t.Errorf("IsValidRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}
