package server

import (
	"testing"
)

func TestServer_findHandler(t *testing.T) {
	handlers := map[string]HandleFunc{
		"/payments/{play}": func(req *Request) {
			t.Logf("%+v", req)
		},
	}
	params := map[string]string{}
	handler, ok := findHandler("/payments/123", handlers, params)
	if !ok {
		t.Fatalf("want true")
	}
	if handler == nil {
		t.Fatalf("handler is nil")
	}
	value, ok := params["play"]
	if !ok {
		t.Fatalf("no play value")
	}
	if value != "123" {
		t.Fatalf("wrong value, want 123, got %s", value)
	}

	handler(&Request{
		Conn:        nil,
		QueryParams: map[string][]string{"hello": []string{"1", "32"}},
		PathParams:  nil,
	})

	handler, ok = findHandler("/credit/123", handlers, nil)
	if ok != false {
		t.Fatalf("want false")
	}
	if handler != nil {
		t.Fatalf("handler is not nil")
	}

	handlers["/category/{id}/partners/{partner_id}"] = func(req *Request) {
		t.Logf("%+v", req)
	}
	handler, ok = findHandler("/category/321/partners/456", handlers, params)
	if !ok {
		t.Fatalf("want true")
	}
	if handler == nil {
		t.Fatalf("handler is nil")
	}
	value, ok = params["id"]
	if !ok {
		t.Fatalf("no id value")
	}
	if value != "321" {
		t.Fatalf("wrong value, want 321, got %s", value)
	}
	value, ok = params["partner_id"]
	if !ok {
		t.Fatalf("no partner_id value")
	}
	if value != "456" {
		t.Fatalf("wrong value, want 456, got %s", value)
	}
	handler(&Request{
		Conn:        nil,
		QueryParams: nil,
		PathParams:  params,
	})
	/*type fields struct {
		addr     string
		mu       sync.RWMutex
		handlers map[string]HandleFunc
	}
	type args struct {
		path string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   HandleFunc
		want1  bool
	}{
		{
			name: "1",
			fields: fields{
				addr: "",
				mu:   sync.RWMutex{},
				handlers: map[string]HandleFunc{
					"/payments/{id}": func(req *Request) {},
				},
			},
			args:  args{},
			want:  nil,
			want1: false,
		},
	},
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				addr:     tt.fields.addr,
				mu:       tt.fields.mu,
				handlers: tt.fields.handlers,
			}
			got, got1 := s.route(tt.args.path)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("route() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("route() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}*/
}
