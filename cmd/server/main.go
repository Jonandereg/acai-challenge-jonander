package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/acai-travel/tech-challenge/internal/chat"
	"github.com/acai-travel/tech-challenge/internal/chat/assistant"
	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/httpx"
	"github.com/acai-travel/tech-challenge/internal/mongox"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/acai-travel/tech-challenge/internal/telemetry"
	"github.com/gorilla/mux"
	"github.com/twitchtv/twirp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	mongo := mongox.MustConnect()

	repo := model.New(mongo)
	assist := assistant.New()

	server := chat.NewServer(repo, assist)
	shutdown, err := telemetry.Init(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer shutdown(context.Background())

	// Configure handler
	handler := mux.NewRouter()
	handler.Use(
		httpx.Logger(),
		httpx.Recovery(),
	)

	handler.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "Hi, my name is Clippy!")
	})

	handler.PathPrefix("/twirp/").Handler(pb.NewChatServiceServer(server, twirp.WithServerJSONSkipDefaults(true)))
	traced := otelhttp.NewHandler(
		handler,
		"http.server",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)

	// Start the server
	slog.Info("Starting the server...")
	if err := http.ListenAndServe(":8080", traced); err != nil {
		panic(err)
	}
}
