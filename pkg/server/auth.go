package server

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

func (h *handler) createToken(w http.ResponseWriter, r *http.Request) {
	user := r.PostFormValue("username")
	pass := r.PostFormValue("password")

	if token, err := h.config.Authorization.Authorizer.CreateToken(user, pass); err == nil {
		http.SetCookie(w, &http.Cookie{
			Name:     "jwt",
			Value:    token,
			HttpOnly: true,
			Path:     "/",
			SameSite: http.SameSiteLaxMode,
		})

		log.Infof("Token created for user %s", user)
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte(http.StatusText(http.StatusOK)))
		if err != nil {
			log.Errorf("Error while creating token: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		log.Errorf("Error while creating token: %v", err)
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
}

func (h *handler) deleteToken(w http.ResponseWriter, _ *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(0, 0),
	})
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(http.StatusText(http.StatusOK))); err != nil {
		log.Errorf("Error while deleting token: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
