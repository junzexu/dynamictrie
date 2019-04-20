package dynamictrie

import (
	"testing"
)

func TestGetBits(t *testing.T) {
	type args struct {
		x    uint64
		p    uint64
		size uint64
	}
	tests := []struct {
		name string
		args args
		want uint64
	}{
		// TODO: Add test cases.
		{"4 bits", args{0x5555555555555555, 0, 4}, 0x05},
		{"4 bits", args{0x5555555555555555, 1, 4}, 0x0A},
		{"1 bits", args{0x5555555555555555, 1, 2}, 0x02},
		{"1 bits", args{0x5555555555555555, 2, 2}, 0x01},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBits(tt.args.x, tt.args.p, tt.args.size); got != tt.want {
				t.Errorf("GetBits() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestV6Net_CheckSub(t *testing.T) {
	type args struct {
		t *V6Net
	}
	tests := []struct {
		name string
		args args
		want V6NetRel
		net  *V6Net
		ones uint64
	}{
		// TODO: Add test cases.
		{"equal", args{NewV6Net("255.255.255.0/24")}, v6NetEqual, NewV6Net("255.255.255.255/24"), 24},
		{"equal", args{NewV6Net("255.255.255.0/16")}, v6NetContainedRight, NewV6Net("255.255.255.255/24"), 16},
		{"equal", args{NewV6Net("255.255.255.0/16")}, v6NetContainedLeft, NewV6Net("255.255.0.255/24"), 16},
		{"equal", args{NewV6Net("255.255.255.255/24")}, v6NetContainRight, NewV6Net("255.255.255.0/16"), 16},
		{"equal", args{NewV6Net("255.255.0.255/24")}, v6NetContainLeft, NewV6Net("255.255.255.0/16"), 16},
		{"equal", args{NewV6Net("255.255.0.255/24")}, v6OvelapRight, NewV6Net("255.255.255.0/24"), 16},
		{"equal", args{NewV6Net("255.255.255.255/24")}, v6OvelapLeft, NewV6Net("255.255.0.0/24"), 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, ones := tt.net.CheckSub(tt.args.t); got != tt.want || ones != tt.ones+128-32 {
				t.Errorf("V6Net.CheckSub() = %v, want %v", got, tt.want)
			}
		})
	}
}
