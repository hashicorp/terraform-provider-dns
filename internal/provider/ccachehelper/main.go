package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/jcmturner/gofork/encoding/asn1"
	"github.com/jcmturner/gokrb5/v8/client"
	"github.com/jcmturner/gokrb5/v8/config"
	"github.com/jcmturner/gokrb5/v8/iana/nametype"
	"github.com/jcmturner/gokrb5/v8/messages"
	"github.com/jcmturner/gokrb5/v8/types"
)

type byteWriter struct {
	buf []byte
}

func (w *byteWriter) writeInt8(v int8) {
	w.buf = append(w.buf, byte(v))
}

func (w *byteWriter) writeInt16(v int16) {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(v))
	w.buf = append(w.buf, b...)
}

func (w *byteWriter) writeInt32(v int32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(v))
	w.buf = append(w.buf, b...)
}

func (w *byteWriter) writeData(data []byte) {
	w.writeInt32(int32(len(data)))
	w.buf = append(w.buf, data...)
}

func (w *byteWriter) writeTimestamp(t time.Time) {
	w.writeInt32(int32(t.Unix()))
}

func (w *byteWriter) writePrincipal(nameType, numComponents int32, realm string, components []string) {
	w.writeInt32(nameType)
	w.writeInt32(numComponents)
	w.writeData([]byte(realm))
	for _, comp := range components {
		w.writeData([]byte(comp))
	}
}

func (w *byteWriter) writeCredential(clientName, clientRealm string, serverName, serverRealm string, key types.EncryptionKey, authTime, startTime, endTime, renewTill time.Time, flags asn1.BitString, ticket, secondTicket []byte) {
	w.writePrincipal(nametype.KRB_NT_PRINCIPAL, 1, clientRealm, []string{clientName})
	w.writePrincipal(nametype.KRB_NT_SRV_INST, 1, serverRealm, []string{serverName, serverRealm})
	w.writeInt16(int16(key.KeyType))
	w.writeData(key.KeyValue)
	w.writeTimestamp(authTime)
	w.writeTimestamp(startTime)
	w.writeTimestamp(endTime)
	w.writeTimestamp(renewTill)
	w.writeInt8(0)
	w.buf = append(w.buf, flags.Bytes...)
	w.writeInt32(0)
	w.writeInt32(0)
	w.writeData(ticket)
	w.writeData(secondTicket)
}

func getSessionData(cl *client.Client) (realm string, authTime, endTime, renewTill time.Time, tgt messages.Ticket, sessionKey types.EncryptionKey, err error) {
	v := reflect.ValueOf(cl).Elem()

	sessionsField := v.FieldByName("sessions")
	if !sessionsField.IsValid() {
		return "", time.Time{}, time.Time{}, time.Time{}, messages.Ticket{}, types.EncryptionKey{}, fmt.Errorf("could not find sessions field on client")
	}

	entriesField := sessionsField.Elem().FieldByName("Entries")
	if !entriesField.IsValid() {
		return "", time.Time{}, time.Time{}, time.Time{}, messages.Ticket{}, types.EncryptionKey{}, fmt.Errorf("could not find Entries field on sessions")
	}

	for _, key := range entriesField.MapKeys() {
		sessionVal := entriesField.MapIndex(key).Elem()

		r := sessionVal.FieldByName("realm").String()
		at := sessionVal.FieldByName("authTime").Interface().(time.Time)
		et := sessionVal.FieldByName("endTime").Interface().(time.Time)
		rt := sessionVal.FieldByName("renewTill").Interface().(time.Time)
		t := sessionVal.FieldByName("tgt").Interface().(messages.Ticket)
		sk := sessionVal.FieldByName("sessionKey").Interface().(types.EncryptionKey)

		return r, at, et, rt, t, sk, nil
	}

	return "", time.Time{}, time.Time{}, time.Time{}, messages.Ticket{}, types.EncryptionKey{}, fmt.Errorf("no session found")
}

func main() {
	realm := os.Getenv("DNS_UPDATE_REALM")
	if realm == "" {
		realm = "EXAMPLE.COM"
	}
	username := os.Getenv("DNS_UPDATE_USERNAME")
	if username == "" {
		fmt.Fprintln(os.Stderr, "DNS_UPDATE_USERNAME is required")
		os.Exit(1)
	}
	krb5ConfigPath := os.Getenv("KRB5_CONFIG")
	if krb5ConfigPath == "" {
		fmt.Fprintln(os.Stderr, "KRB5_CONFIG is required")
		os.Exit(1)
	}

	cfg, err := config.Load(krb5ConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load krb5 config: %v\n", err)
		os.Exit(1)
	}

	password := os.Getenv("DNS_UPDATE_PASSWORD")
	if password == "" {
		fmt.Fprintln(os.Stderr, "DNS_UPDATE_PASSWORD is required")
		os.Exit(1)
	}

	cl := client.NewWithPassword(username, realm, password, cfg,
		client.DisablePAFXFAST(true),
	)

	if err := cl.Login(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to login to KDC: %v\n", err)
		os.Exit(1)
	}
	defer cl.Destroy()

	sRealm, authTime, endTime, renewTill, tgt, sessionKey, err := getSessionData(cl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get session data: %v\n", err)
		os.Exit(1)
	}

	ticketBytes, err := tgt.Marshal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal TGT: %v\n", err)
		os.Exit(1)
	}

	var flagsBytes [4]byte
	binary.BigEndian.PutUint32(flagsBytes[:], 0x400000)
	flags := asn1.BitString{
		Bytes:     flagsBytes[:],
		BitLength: 32,
	}

	var w byteWriter
	w.writeInt8(5)
	w.writeInt8(4)

	w.writePrincipal(
		nametype.KRB_NT_PRINCIPAL, 1,
		realm,
		[]string{username},
	)

	w.writeCredential(
		username, realm,
		"krbtgt", sRealm,
		sessionKey,
		authTime, authTime, endTime, renewTill,
		flags,
		ticketBytes, nil,
	)

	cachePath := filepath.Join(os.TempDir(), "krb5cc_dnshelper")
	if err := os.WriteFile(cachePath, w.buf, 0600); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write credential cache: %v\n", err)
		os.Exit(1)
	}

	fmt.Print(cachePath)
}
