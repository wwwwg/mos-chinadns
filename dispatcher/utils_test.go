//     Copyright (C) 2020, IrineSistiana
//
//     This file is part of mos-chinadns.
//
//     mos-chinadns is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     mos-chinadns is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with this program.  If not, see <https://www.gnu.org/licenses/>.
package dispatcher

import (
	"bytes"
	"github.com/miekg/dns"
	"io"
	"reflect"
	"testing"
)

func Test_readMsgFromTCP(t *testing.T) {
	q := new(dns.Msg).SetQuestion("www.google.com.", dns.TypeA)
	bb := new(bytes.Buffer)
	data, err := q.Pack()
	if err != nil {
		t.Fatal(err)
	}
	bb.WriteByte(byte(len(data) >> 8))
	bb.WriteByte(byte(len(data)))
	bb.Write(data)
	type args struct {
		c io.Reader
	}
	tests := []struct {
		name               string
		args               args
		wantM              *dns.Msg
		wantBrokenDataLeft int
		wantN              int
		wantErr            bool
	}{
		{"normal read", args{bb}, q, 0, len(data) + 2, false},
		{"short read", args{bytes.NewBuffer(bb.Bytes()[:2+13])}, nil, len(data) - 13, 2 + 13, true},
		{"broken length header", args{bytes.NewBuffer(bb.Bytes()[:1])}, nil, unknownBrokenDataSize, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotM, gotBrokenDataLeft, gotN, err := readMsgFromTCP(tt.args.c)
			if (err != nil) != tt.wantErr {
				t.Errorf("readMsgFromTCP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotM, tt.wantM) {
				t.Errorf("readMsgFromTCP() gotM = %v, want %v", gotM, tt.wantM)
			}
			if gotBrokenDataLeft != tt.wantBrokenDataLeft {
				t.Errorf("readMsgFromTCP() gotBrokenDataLeft = %v, want %v", gotBrokenDataLeft, tt.wantBrokenDataLeft)
			}
			if gotN != tt.wantN {
				t.Errorf("readMsgFromTCP() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}

func Test_writeMsgToTCP(t *testing.T) {
	q := new(dns.Msg).SetQuestion("www.google.com.", dns.TypeA)
	bb := new(bytes.Buffer)
	data, err := q.Pack()
	if err != nil {
		t.Fatal(err)
	}
	bb.WriteByte(byte(len(data) >> 8))
	bb.WriteByte(byte(len(data)))
	bb.Write(data)

	type args struct {
		m *dns.Msg
	}
	tests := []struct {
		name    string
		args    args
		wantC   string
		wantN   int
		wantErr bool
	}{
		{"write", args{q}, bb.String(), len(data) + 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &bytes.Buffer{}
			gotN, err := writeMsgToTCP(c, tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeMsgToTCP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotC := c.String(); gotC != tt.wantC {
				t.Errorf("writeMsgToTCP() gotC = %v, want %v", gotC, tt.wantC)
			}
			if gotN != tt.wantN {
				t.Errorf("writeMsgToTCP() gotN = %v, want %v", gotN, tt.wantN)
			}
		})
	}
}
