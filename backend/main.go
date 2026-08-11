package main

import (
	"backend/api"
	"backend/chat"
	"backend/middleware"
	"backend/notification"
	"backend/realtime"
	"backend/user"
	"errors"
	"log"
	"net/http"
	"time"
)

func runServer(handler http.Handler) {
	server := &http.Server{
		Addr:              "0.0.0.0:8000",
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       60 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}
	log.Printf("Starting Go server on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("Server failed: %v", err)
	}
}

func emailer_create() *user.SMTPEmailSender {
	emailSender, err := user.NewSMTPEmailSender(
		user.SMTPConfig{
			Host:              env("SMTP_HOST"),
			Port:              env("SMTP_PORT"),
			UserName:          env("SMTP_USERNAME"),
			Password:          env("SMTP_PASSWORD"),
			From:              env("SMTP_FROM"),
			PublicBackendURL:  env("PUBLIC_BACKEND_URL"),
			PublicFrontendURL: env("FRONTEND_URL"),
		})
	if err != nil {
		log.Fatalf("Initialize email sender: %v", err)
	}
	return emailSender
}

func main() {
	database := initDB()
	defer database.Close()

	hub := realtime.NewHub()

	emailSender := emailer_create()

	chatModule := chat.NewModule(database)
	userModule := user.NewModule(database, emailSender)
	notificationModule := notification.NewModule(database)

	hub.SetPresenceRecorder(userModule.Service)
	go hub.Run()

	notificationModule.Service.SetPublisher(hub)
	userModule.Handler.SetUserNotifier(notificationModule.Service)
	userModule.Handler.SetUserPresence(hub)

	websocketHandler := realtime.NewWebsocketHandler(hub, chatModule.Service, notificationModule.Service)

	mux := http.NewServeMux()
	privateMux := http.NewServeMux()

	loginLimiter := middleware.NewRateLimiter(5, time.Minute)
	registerLimiter := middleware.NewRateLimiter(3, time.Hour)
	resendVerificationLimiter := middleware.NewRateLimiter(3, 15*time.Minute)
	forgotPasswordLimiter := middleware.NewRateLimiter(3, 15*time.Minute)
	resetPasswordLimiter := middleware.NewRateLimiter(5, 15*time.Minute)

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	mux.Handle("POST /api/register", registerLimiter.Limit(http.HandlerFunc(userModule.Handler.RegisterUser)))
	mux.Handle("POST /api/login", loginLimiter.Limit(http.HandlerFunc(userModule.Handler.LoginUser)))
	mux.Handle("POST /api/email-verification/resend", resendVerificationLimiter.Limit(http.HandlerFunc(userModule.Handler.ResendVerificationEmail)))
	mux.Handle("POST /api/password/forgot", forgotPasswordLimiter.Limit(http.HandlerFunc(userModule.Handler.RequestPasswordReset)))
	mux.Handle("POST /api/password/reset", resetPasswordLimiter.Limit(http.HandlerFunc(userModule.Handler.ResetPassword)))
	mux.HandleFunc("GET /api/verify-email", userModule.Handler.VerifyEmail)
	mux.HandleFunc("POST /api/accounts/token/refresh", api.RefreshToken(userModule.Service))
	mux.HandleFunc("POST /api/accounts/logout/", userModule.Handler.LogoutUser)

	privateMux.HandleFunc("GET /verify_login", userModule.Handler.VerifyUser)
	privateMux.HandleFunc("GET /profile", userModule.Handler.GetProfile)
	privateMux.HandleFunc("GET /profile/views", userModule.Handler.ListProfileViewers)
	privateMux.HandleFunc("GET /profile/likes", userModule.Handler.ListProfileLikers)
	privateMux.HandleFunc("GET /profiles/feed", userModule.Handler.GetProfileFeed)
	privateMux.HandleFunc("GET /profiles/search", userModule.Handler.SearchProfiles)
	privateMux.HandleFunc("GET /profiles/{targetUserID}", userModule.Handler.GetPublicProfile)
	privateMux.HandleFunc("GET /matches", userModule.Handler.ListMatches)
	privateMux.HandleFunc("GET /ws", websocketHandler.Connect)
	privateMux.HandleFunc("GET /conversations", chatModule.Handler.ListConversations)
	privateMux.HandleFunc("GET /conversations/{conversationID}/messages", chatModule.Handler.ListConversationMessages)
	privateMux.HandleFunc("GET /notifications", notificationModule.Handler.ListNotifications)
	privateMux.HandleFunc("GET /interests", userModule.Handler.ListInterestTags)

	privateMux.HandleFunc("POST /profile/complete", userModule.Handler.CompleteProfile)
	privateMux.HandleFunc("POST /avatar", userModule.Handler.UploadAvatar)
	privateMux.HandleFunc("POST /photos", userModule.Handler.UploadPhoto)

	privateMux.HandleFunc("PATCH /profile", userModule.Handler.UpdateProfile)
	privateMux.HandleFunc("PATCH /user", userModule.Handler.UpdateUser)
	privateMux.HandleFunc("PATCH /user/password", userModule.Handler.UpdatePassword)
	privateMux.HandleFunc("PATCH /messages/{messageID}/read", chatModule.Handler.MarkMessageRead)
	privateMux.HandleFunc("PATCH /notifications/{notificationID}/read", notificationModule.Handler.MarkNotificationRead)

	privateMux.HandleFunc("PUT /profiles/{targetUserID}/decision", userModule.Handler.SaveProfileDecision)
	privateMux.HandleFunc("PUT /profiles/{targetUserID}/block", userModule.Handler.BlockUser)
	privateMux.HandleFunc("PUT /profiles/{targetUserID}/report", userModule.Handler.ReportUser)

	privateMux.HandleFunc("DELETE /photos/{photoID}", userModule.Handler.DeletePhoto)
	privateMux.HandleFunc("DELETE /profiles/{targetUserID}/block", userModule.Handler.UnblockUser)

	mux.Handle("/api/accounts/", http.StripPrefix("/api/accounts", middleware.AuthMiddleware(privateMux)))

	runServer(mux)
}
