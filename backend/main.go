package main

import (
	"backend/api"
	"backend/chat"
	"backend/middleware"
	"backend/realtime"
	"backend/user"
	"fmt"
	"log"
	"net/http"
)

func server_run(mux *http.ServeMux) {
	fmt.Println("Starting Go server on port 8000...")
	err := http.ListenAndServe("0.0.0.0:8000", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func main() {
	database := initDB()
	defer database.Close()

	hub := realtime.NewHub()
	go hub.Run()

	chatModule := chat.NewModule(database)
	userModule := user.NewModule(database)

	websocketHandler := realtime.NewWebsocketHandler(hub, chatModule.Service)

	mux := http.NewServeMux()
	privateMux := http.NewServeMux()

	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	mux.HandleFunc("POST /api/register", userModule.Handler.RegisterUser)
	mux.HandleFunc("POST /api/login", userModule.Handler.LoginUser)
	mux.HandleFunc("POST /api/accounts/token/refresh", api.RefreshToken)

	privateMux.HandleFunc("GET /verify_login", userModule.Handler.VerifyUser)
	privateMux.HandleFunc("GET /profile", userModule.Handler.GetProfile)
	privateMux.HandleFunc("GET /profiles/feed", userModule.Handler.GetProfileFeed)
	privateMux.HandleFunc("GET /matches", userModule.Handler.ListMatches)
	privateMux.HandleFunc("GET /ws", websocketHandler.Connect)
	privateMux.HandleFunc("GET /conversations", chatModule.Handler.ListConversations)
	privateMux.HandleFunc("GET /conversations/{conversationID}/messages", chatModule.Handler.ListConversationMessages)
	// privateMux.HandleFunc("GET /notifications", listNotifications)

	privateMux.HandleFunc("POST /logout/", userModule.Handler.LogoutUser)
	privateMux.HandleFunc("POST /profile/complete", userModule.Handler.CompleteProfile)
	privateMux.HandleFunc("POST /avatar", userModule.Handler.UploadAvatar)
	privateMux.HandleFunc("POST /photos", userModule.Handler.UploadPhoto)
	// privateMux.HandleFunc("POST /notifications/send", notificationHandler(hub))

	privateMux.HandleFunc("PATCH /profile", userModule.Handler.UpdateProfile)
	privateMux.HandleFunc("PATCH /user", userModule.Handler.UpdateUser)
	privateMux.HandleFunc("PATCH /user/password", userModule.Handler.UpdatePassword)
	privateMux.HandleFunc("PATCH /messages/{messageID}/read", chatModule.Handler.MarkMessageRead)
	// privateMux.HandleFunc("PATCH /notifications/:notificationID/read", markNotificationRead)

	privateMux.HandleFunc("PUT /profiles/{targetUserID}/decision", userModule.Handler.SaveProfileDecision)

	privateMux.HandleFunc("DELETE /photos/{photoID}", userModule.Handler.DeletePhoto)

	mux.Handle("/api/accounts/", http.StripPrefix("/api/accounts", middleware.AuthMiddleware(privateMux)))

	server_run(mux)
}
