package handlers

import "net/http"

func isHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func redirect(w http.ResponseWriter, r *http.Request, url string) {
	if isHTMX(r) {
		w.Header().Set("HX-Redirect", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, http.StatusSeeOther)
}
