package main

import (
	"context"
	"fmt"
	"net/http"
	"path"

	"github.com/open-policy-agent/opa/rego"

	"github.com/gorilla/mux"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("home")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func EnabledHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("enabled")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func DisabledHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("disabled")); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func PolicyMiddleware(f http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		rBasePath := path.Base(request.URL.Path)

		r := rego.New(
			rego.Query(fmt.Sprintf("x = data.example.%s", rBasePath)),
			rego.Load([]string{"./example.rego"}, nil))

		ctx := context.Background()
		query, err := r.PrepareForEval(ctx)
		if err != nil {
			panic(err)
		}

		rs, err := query.Eval(context.Background())
		if err != nil {
			panic(err)
		}

		var allow bool
		for _, r := range rs {
			allow = r.Bindings["x"].(bool)
		}

		if !allow {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		f.ServeHTTP(writer, request)
	})
}

func main() {
	r := mux.NewRouter()
	r.Use(PolicyMiddleware)
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/enabled", EnabledHandler)
	r.HandleFunc("/disabled", DisabledHandler)
	http.Handle("/", r)

	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}
}
