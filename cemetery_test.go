package gone

import (
	"testing"
)

type XPoint interface {
	GetX() int
	GetY() int
}
type Point struct {
	GonerFlag
	x int
	y int

	Index int
}

func (p *Point) GetX() int {
	return p.x
}
func (p *Point) GetY() int {
	return p.y
}

func Test_cemetery_revive(t *testing.T) {
	type fields struct {
		goneList []Tomb
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "inject field is Struct",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						a Point `gone:"point-a"`
						b Point `gone:"point-b"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},
		{
			name: "inject field is Struct pointer",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						a *Point `gone:"point-a"`
						b *Point `gone:"point-b"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is interface",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						a XPoint `gone:"point-a"`
						b XPoint `gone:"point-b"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field(with id) is interface slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c []XPoint `gone:"point-a"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: true,
		},

		{
			name: "inject field is interface slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c []XPoint `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is struct pointer slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c []*Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is struct slice",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c []Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},
		{
			name: "inject field is map[string]interface",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c map[string]XPoint `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is map[string]interface",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c map[GonerId]XPoint `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is map[string]pointer",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c map[GonerId]*Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},

		{
			name: "inject field is map[string]struct",
			fields: fields{
				goneList: []Tomb{
					NewTomb(&struct {
						GonerFlag
						c map[GonerId]Point `gone:"*"`
					}{}).SetId("line"),
					NewTomb(&Point{x: 1, y: 2}).SetId("point-a"),
					NewTomb(&Point{x: -1, y: -2}).SetId("point-b"),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cemetery{
				Logger:  &defaultLogger{},
				tombMap: make(map[GonerId]Tomb),
			}

			for _, tomb := range tt.fields.goneList {
				c.Bury(tomb.GetGoner(), tomb.GetId())
			}

			if err := c.revive(); (err != nil) != tt.wantErr {
				t.Errorf("revive() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
