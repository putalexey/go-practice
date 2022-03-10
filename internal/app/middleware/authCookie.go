package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/google/uuid"
)

type AuthKey string

var UIDKey = AuthKey("UID")

// AuthCookie creates middleware that will create cookie with user authentication.
// Middleware adds user id to the request context with key middleware.UIDKey. user id is UUID (Version 4)
// keyString is used to encrypt cookie value.
func AuthCookie(cookieName string, keyString string) func(http.Handler) http.Handler {
	tmp := sha256.Sum256([]byte(keyString))
	key := tmp[:]

	return func(next http.Handler) http.Handler {
		handler := authCookieHandler{
			next:       next,
			cookieName: cookieName,
			key:        key,
		}
		return handler
	}
}

type authCookieHandler struct {
	next       http.Handler
	cookieName string
	key        []byte
}

func (h authCookieHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		userCtx context.Context
		uid     string
		err     error
	)

	// if cookie exists, but can't be decoded `http.ErrNoCookie` will be returned too
	uid, err = h.findUIDInCookies(r) // r.Cookie(cookieName)
	if err != nil {
		if err != http.ErrNoCookie {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		uid, err = randUID()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		authCookie, err := h.newCookie(uid)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, authCookie)
	}

	userCtx = context.WithValue(r.Context(), UIDKey, uid)
	h.next.ServeHTTP(w, r.WithContext(userCtx))
}

func randUID() (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

// findUIDInCookies gets cookie and tries to decrypt it, returns http.ErrNoCookie on cookie absence or decrypt error
func (h authCookieHandler) findUIDInCookies(r *http.Request) (string, error) {
	authCookie, err := r.Cookie(h.cookieName)
	if err != nil {
		return "", err
	}

	uid, err := h.decryptUID(authCookie.Value)
	if err != nil {
		return "", http.ErrNoCookie
	}

	return uid, nil
}

func (h authCookieHandler) newCookie(uid string) (*http.Cookie, error) {
	encryptedUID, err := h.encryptUID(uid)
	if err != nil {
		return nil, err
	}

	cook := http.Cookie{
		Name:     h.cookieName,
		Value:    encryptedUID,
		Path:     "/",
		HttpOnly: true,
		Raw:      "",
		Unparsed: nil,
	}
	return &cook, nil
}

func (h authCookieHandler) encryptUID(uid string) (string, error) {
	aesgcm, err := h.prepareGcm()
	if err != nil {
		return "", err
	}
	nonce := h.key[0:aesgcm.NonceSize()]
	encrypted := aesgcm.Seal(nil, nonce, []byte(uid), nil)

	return hex.EncodeToString(encrypted), nil
}

func (h authCookieHandler) decryptUID(encrypted string) (string, error) {
	aesgcm, err := h.prepareGcm()
	if err != nil {
		return "", err
	}
	nonce := h.key[0:aesgcm.NonceSize()]

	decoded, err := hex.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	uid, err := aesgcm.Open(nil, nonce, decoded, nil)
	if err != nil {
		return "", err
	}

	return string(uid), nil
}

func (h authCookieHandler) prepareGcm() (cipher.AEAD, error) {
	aesblock, err := aes.NewCipher(h.key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}

	return aesgcm, nil
}
