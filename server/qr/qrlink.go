package qr

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"github.com/skip2/go-qrcode"
	"html/template"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func New(maxEntries int, lifeTime time.Duration, host string) *Generator {
	log.Printf("Creating QR code generator with maxEntries=%d and lifeTime=%v", maxEntries, lifeTime)
	gen := &Generator{
		maxEntries: maxEntries,
		lifeTime:   lifeTime,
		host:       host,
		store:      make(map[string]codeEntry),
	}
	gen.cleanupOldEntries()
	return gen
}

func DecodeSourceCode(b64Source string) (string, error) {
	b, err := base64.URLEncoding.WithPadding(-1).DecodeString(b64Source)
	if err != nil {
		return "", fmt.Errorf("error decoding base64: %w", err)
	} else {
		byteReader := bytes.NewReader(b)
		zlibReader, err := zlib.NewReader(byteReader)
		if err != nil {
			return "", fmt.Errorf("error reading zlib header: %w", err)
		}
		defer zlibReader.Close()
		out, err := io.ReadAll(zlibReader)
		if err != nil {
			return "", fmt.Errorf("error reading zlib data: %w", err)
		} else {
			return string(out), nil
		}
	}
}

func (qr *Generator) add(code string) string {
	qr.mutex.Lock()
	defer qr.mutex.Unlock()

	if len(qr.store) >= qr.maxEntries {
		var oldestId string
		var oldestTime time.Time
		for id, entry := range qr.store {
			if oldestId == "" {
				oldestId = id
				oldestTime = entry.createdAt
			} else if entry.createdAt.Before(oldestTime) {
				oldestId = id
				oldestTime = entry.createdAt
			}
		}
		if oldestId != "" {
			log.Println("Deleting oldest QR code entry with id:", oldestId)
			delete(qr.store, oldestId)
		}
	}

	for id, entry := range qr.store {
		if entry.code == code {
			return id
		}
	}

	id := generateUniqueID()
	qr.store[id] = codeEntry{
		code:      code,
		createdAt: time.Now(),
	}
	log.Println("Generated QR code for id:", id, "-", len(qr.store), "of", qr.maxEntries)
	return id
}

func (qr *Generator) Get(id string) (string, error) {
	qr.mutex.Lock()
	defer qr.mutex.Unlock()
	if entry, exists := qr.store[id]; exists {
		return entry.code, nil
	}
	return "", fmt.Errorf("code not found for id: %s", id)
}

func generateUniqueID() string {
	// create 10 random characters for the unique ID
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 10)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

type codeEntry struct {
	code      string
	createdAt time.Time
}

type Generator struct {
	mutex sync.Mutex
	store map[string]codeEntry

	maxEntries int
	lifeTime   time.Duration
	host       string
}

func (qr *Generator) CreateHandler(temp *template.Template) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		b64code := query.Get("c")
		code, err := DecodeSourceCode(b64code)
		if err != nil {
			http.Error(w, fmt.Sprintf("error decoding source code: %v", err), http.StatusBadRequest)
			return
		}
		host := qr.host
		if host == "" {
			host = "https://" + r.Host
		}
		id := qr.add(code)

		qrCode, err := qrcode.Encode(host+"/?qr="+id, qrcode.Medium, 512)
		if err != nil {
			http.Error(w, fmt.Sprintf("error generating QR code: %v", err), http.StatusInternalServerError)
			return
		}

		base64Image := base64.StdEncoding.EncodeToString(qrCode)
		temp.Execute(w, base64Image)
	}
}

func (qr *Generator) cleanupOldEntries() {
	ticker := time.NewTicker(qr.lifeTime)
	go func() {
		log.Println("Starting cleanup of old QR code entries every", qr.lifeTime)
		for range ticker.C {
			deleted := 0
			qr.mutex.Lock()
			for id, entry := range qr.store {
				if time.Since(entry.createdAt) > qr.lifeTime {
					log.Println("Deleting old QR code entry with id:", id)
					delete(qr.store, id)
					deleted++
				}
			}
			if deleted > 0 && len(qr.store) == 0 {
				log.Println("No more QR code entries left.")
			}
			qr.mutex.Unlock()
		}
	}()
}
