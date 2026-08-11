package main

import (
	"backend/account"
	"backend/api"
	"backend/chat"
	"backend/discovery"
	"backend/middleware"
	"backend/notification"
	"backend/profile"
	"backend/realtime"
	"backend/relationship"
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

func createEmailSender() *account.SMTPEmailSender {
	emailSender, err := account.NewSMTPEmailSender(
		account.SMTPConfig{
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

	emailSender := createEmailSender()

	chatModule := chat.NewModule(database)
	accountModule := account.NewModule(database, emailSender)
	profileModule := profile.NewModule(database)
	discoveryModule := discovery.NewModule(database)
	relationshipModule := relationship.NewModule(database)
	notificationModule := notification.NewModule(database)

	hub.SetPresenceRecorder(profileModule.Service)
	go hub.Run()

	notificationModule.Service.SetPublisher(hub)
	relationshipModule.Handler.SetUserNotifier(notificationModule.Service)
	relationshipModule.Handler.SetUserPresence(hub)

	websocketHandler := realtime.NewWebsocketHandler(hub, chatModule.Service, notificationModule.Service)

	mux := http.NewServeMux()
	privateMux := http.NewServeMux()

	loginLimiter := middleware.NewRateLimiter(5, time.Minute)
	registerLimiter := middleware.NewRateLimiter(3, time.Hour)
	resendVerificationLimiter := middleware.NewRateLimiter(3, 15*time.Minute)
	forgotPasswordLimiter := middleware.NewRateLimiter(3, 15*time.Minute)
	resetPasswordLimiter := middleware.NewRateLimiter(5, 15*time.Minute)

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	mux.Handle("POST /api/register", registerLimiter.Limit(http.HandlerFunc(accountModule.Handler.RegisterUser)))
	mux.Handle("POST /api/login", loginLimiter.Limit(http.HandlerFunc(accountModule.Handler.LoginUser)))
	mux.Handle("POST /api/email-verification/resend", resendVerificationLimiter.Limit(http.HandlerFunc(accountModule.Handler.ResendVerificationEmail)))
	mux.Handle("POST /api/password/forgot", forgotPasswordLimiter.Limit(http.HandlerFunc(accountModule.Handler.RequestPasswordReset)))
	mux.Handle("POST /api/password/reset", resetPasswordLimiter.Limit(http.HandlerFunc(accountModule.Handler.ResetPassword)))
	mux.HandleFunc("GET /api/verify-email", accountModule.Handler.VerifyEmail)
	mux.HandleFunc("POST /api/accounts/token/refresh", api.RefreshToken(accountModule.Service))
	mux.HandleFunc("POST /api/accounts/logout/", accountModule.Handler.LogoutUser)

	privateMux.HandleFunc("GET /verify_login", accountModule.Handler.VerifyUser)
	privateMux.HandleFunc("GET /profile", profileModule.Handler.GetProfile)
	privateMux.HandleFunc("GET /profile/views", relationshipModule.Handler.ListProfileViewers)
	privateMux.HandleFunc("GET /profile/likes", relationshipModule.Handler.ListProfileLikers)
	privateMux.HandleFunc("GET /profiles/feed", discoveryModule.Handler.GetProfileFeed)
	privateMux.HandleFunc("GET /profiles/search", discoveryModule.Handler.SearchProfiles)
	privateMux.HandleFunc("GET /profiles/{targetUserID}", relationshipModule.Handler.GetPublicProfile)
	privateMux.HandleFunc("GET /matches", relationshipModule.Handler.ListMatches)
	privateMux.HandleFunc("GET /ws", websocketHandler.Connect)
	privateMux.HandleFunc("GET /conversations", chatModule.Handler.ListConversations)
	privateMux.HandleFunc("GET /conversations/{conversationID}/messages", chatModule.Handler.ListConversationMessages)
	privateMux.HandleFunc("GET /notifications", notificationModule.Handler.ListNotifications)
	privateMux.HandleFunc("GET /interests", profileModule.Handler.ListInterestTags)

	privateMux.HandleFunc("POST /profile/complete", profileModule.Handler.CompleteProfile)
	privateMux.HandleFunc("POST /avatar", profileModule.Handler.UploadAvatar)
	privateMux.HandleFunc("POST /photos", profileModule.Handler.UploadPhoto)

	privateMux.HandleFunc("PATCH /profile", profileModule.Handler.UpdateProfile)
	privateMux.HandleFunc("PATCH /user", accountModule.Handler.UpdateUser)
	privateMux.HandleFunc("PATCH /user/password", accountModule.Handler.UpdatePassword)
	privateMux.HandleFunc("PATCH /messages/{messageID}/read", chatModule.Handler.MarkMessageRead)
	privateMux.HandleFunc("PATCH /notifications/{notificationID}/read", notificationModule.Handler.MarkNotificationRead)

	privateMux.HandleFunc("PUT /profiles/{targetUserID}/decision", relationshipModule.Handler.SaveProfileDecision)
	privateMux.HandleFunc("PUT /profiles/{targetUserID}/block", relationshipModule.Handler.BlockUser)
	privateMux.HandleFunc("PUT /profiles/{targetUserID}/report", relationshipModule.Handler.ReportUser)

	privateMux.HandleFunc("DELETE /photos/{photoID}", profileModule.Handler.DeletePhoto)
	privateMux.HandleFunc("DELETE /profiles/{targetUserID}/block", relationshipModule.Handler.UnblockUser)

	mux.Handle("/api/accounts/", http.StripPrefix("/api/accounts", middleware.AuthMiddleware(privateMux)))

	runServer(mux)
}
