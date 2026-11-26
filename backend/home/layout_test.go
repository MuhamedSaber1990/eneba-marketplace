package home

import (
	"slices"
	"testing"
)

func TestBuildLayout(t *testing.T) {
	tests := []struct {
		name      string
		rows      []LayoutRow
		wantTypes []string
	}{
		{
			name:      "no rows gives an empty list, not nil",
			rows:      nil,
			wantTypes: []string{},
		},
		{
			name: "keeps the order of the rows",
			rows: []LayoutRow{
				{BlockType: "deals", Enabled: true},
				{BlockType: "banner", Enabled: true},
				{BlockType: "platforms", Enabled: true},
			},
			wantTypes: []string{"deals", "banner", "platforms"},
		},
		{
			name: "leaves out disabled blocks",
			rows: []LayoutRow{
				{BlockType: "banner", Enabled: true},
				{BlockType: "deals", Enabled: false},
				{BlockType: "best_sellers", Enabled: true},
			},
			wantTypes: []string{"banner", "best_sellers"},
		},
		{
			name: "leaves out block types the backend cannot serve",
			rows: []LayoutRow{
				{BlockType: "banner", Enabled: true},
				{BlockType: "horoscope", Enabled: true},
				{BlockType: "new_releases", Enabled: true},
			},
			wantTypes: []string{"banner", "new_releases"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blocks := BuildLayout(tt.rows)
			if blocks == nil {
				t.Fatal("got nil, want an empty slice so JSON shows []")
			}
			gotTypes := []string{}
			for _, block := range blocks {
				gotTypes = append(gotTypes, block.Type)
			}
			if !slices.Equal(gotTypes, tt.wantTypes) {
				t.Errorf("types = %v, want %v", gotTypes, tt.wantTypes)
			}
		})
	}
}

func TestBuildLayoutFillsEveryField(t *testing.T) {
	blocks := BuildLayout([]LayoutRow{
		{BlockType: "best_sellers", Title: "Best sellers", SeeAllURL: "/store", Enabled: true},
	})
	want := Block{Type: "best_sellers", Title: "Best sellers", SeeAllURL: "/store", Endpoint: "/api/home/best-sellers"}
	if len(blocks) != 1 || blocks[0] != want {
		t.Errorf("got %+v, want [%+v]", blocks, want)
	}
}

func TestDiscountPercent(t *testing.T) {
	tests := []struct {
		name     string
		original int
		price    int
		want     int
	}{
		{"thirty percent off", 4999, 3499, 30},
		{"rounds to the nearest whole percent", 1000, 666, 33},
		{"rounds half up", 1000, 665, 34},
		{"same price is no discount", 2000, 2000, 0},
		{"a higher price is no discount", 2000, 2500, 0},
		{"a zero original price does not divide by zero", 0, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DiscountPercent(tt.original, tt.price); got != tt.want {
				t.Errorf("DiscountPercent(%d, %d) = %d, want %d", tt.original, tt.price, got, tt.want)
			}
		})
	}
}
