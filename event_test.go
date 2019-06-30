package main

import (
	"testing"
	"time"
)

func TestEvent_validate(t *testing.T) {
	type fields struct {
		Event     string
		Timestamp time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "Test Ok", fields: fields{Event: "buy", Timestamp: time.Now()}, wantErr: false},
		{name: "Test Missing Event", fields: fields{Timestamp: time.Now()}, wantErr: true},
		{name: "Test Timestamp Event", fields: fields{Event: "buy"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Event{
				Event:     tt.fields.Event,
				Timestamp: tt.fields.Timestamp,
			}
			if err := e.validate(); (err != nil) != tt.wantErr {
				t.Errorf("Event.validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
