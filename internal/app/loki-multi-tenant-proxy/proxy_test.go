package proxy

import (
	"testing"

	"github.com/angelbarrera92/loki-multi-tenant-proxy/internal/pkg"
)

func Test_isAuthorized(t *testing.T) {
	users := pkg.Authn{
		[]pkg.User{
			pkg.User{
				"User-a",
				"pass-a",
				"tenant-a",
			},
			pkg.User{
				"User-b",
				"pass-b",
				"tenant-b",
			},
		},
	}
	type args struct {
		user  string
		pass  string
		users *pkg.Authn
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Valid User",
			args{
				"User-a",
				"pass-a",
				&users,
			},
			true,
		}, {
			"Invalid User",
			args{
				"invalid",
				"pass-a",
				&users,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAuthorized(tt.args.user, tt.args.pass, tt.args.users); got != tt.want {
				t.Errorf("isAuthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}
