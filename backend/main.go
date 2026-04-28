package main

import (
	"fmt"
	"log"
	"net/http"

	"gorm.io/gorm"

	"goji.io"
	"goji.io/pat"
)

var DB *gorm.DB

func helloWorld(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to your new Go backend!")
}

func server_run(mux *goji.Mux) {
	err := http.ListenAndServe("0.0.0.0:8000", mux)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func main() {
	initDB()
	mux := goji.NewMux()
	privateMux := goji.SubMux()
	privateMux.Use(authMiddleware)

	mux.HandleFunc(pat.Get("/"), helloWorld)
	mux.HandleFunc(pat.Post("/api/register"), registerUser)
	mux.HandleFunc(pat.Post("/api/login"), loginUser)
	mux.HandleFunc(pat.Post("/api/accounts/token/refresh/"), refreshToken)

	privateMux.HandleFunc(pat.Get("/verify_login/"), verifyUser)
	privateMux.HandleFunc(pat.Post("/logout/"), logoutUser)
	privateMux.HandleFunc(pat.Patch("/bio/update"), updateBio)
	privateMux.HandleFunc(pat.Get("/bio/get"), getBio)
	privateMux.HandleFunc(pat.Get("/profiles/feed"), getProfileFeed)
	privateMux.HandleFunc(pat.Post("/bio/create"), createBio)
	privateMux.HandleFunc(pat.Patch("/user/update"), updateUser)
	privateMux.HandleFunc(pat.Patch("/user/password/update"), updatePassword)
	privateMux.HandleFunc(pat.Post("/avatar/upload"), uploadAvatar)
	privateMux.HandleFunc(pat.Post("/photo/upload"), uploadPhoto)

	mux.Handle(pat.New("/uploads/*"), http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))
	mux.Handle(pat.New("/api/accounts/*"), privateMux)
	fmt.Println("Starting Go server on port 8000...")
	server_run(mux)
}
