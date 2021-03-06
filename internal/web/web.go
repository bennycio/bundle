package web

import (
	"net/http"

	"github.com/bennycio/bundle/internal"
	"github.com/rs/cors"
)

func NewWebServer() *http.Server {

	mux := http.NewServeMux()
	rootHandler := http.HandlerFunc(rootHandlerFunc)
	aboutHandler := http.HandlerFunc(aboutHandlerFunc)
	signupHandler := http.HandlerFunc(signupHandlerFunc)
	loginHandler := http.HandlerFunc(loginHandlerFunc)
	logoutHandler := http.HandlerFunc(logoutHandlerFunc)
	pluginHandler := http.HandlerFunc(pluginsHandlerFunc)
	thumbnailHandler := http.HandlerFunc(thumbnailHandlerFunc)
	profileHandler := http.HandlerFunc(profileHandlerFunc)
	stripeAuthHandler := http.HandlerFunc(stripeAuthHandlerFunc)
	stripeReturnHandler := http.HandlerFunc(stripeReturnHandlerFunc)
	purchasePluginHandler := http.HandlerFunc(purchasePluginHandlerFunc)
	premiumHandler := http.HandlerFunc(premiumHandlerFunc)

	mux.Handle("/", rootHandler)
	mux.Handle("/about", aboutHandler)
	mux.Handle("/plugins", pluginHandler)
	mux.Handle("/plugins/", pluginHandler)
	mux.Handle("/plugins/thumbnails", thumbnailHandler)
	mux.Handle("/plugins/purchase", purchasePluginHandler)
	mux.Handle("/plugins/premium", premiumHandler)
	mux.Handle("/profile", loginGate(profileHandler))
	mux.Handle("/stripe/auth", stripeAuthHandler)
	mux.Handle("/stripe/return", stripeReturnHandler)
	mux.Handle("/login", loginHandler)
	mux.Handle("/logout", logoutHandler)
	mux.Handle("/signup", signupHandler)
	mux.Handle("/public/", http.StripPrefix("/public", http.FileServer(http.Dir("assets/public"))))

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"localhost:8080", "bundlemc.io"},
		AllowedMethods: []string{
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	return internal.MakeServerFromMux(handler)
}
