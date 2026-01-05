package main

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

const maxMemory = 10 << 20 //10 megabytes or 10.48576*10^6

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}
	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	r.ParseMultipartForm(maxMemory)
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		log.Printf("FormFile for 'thumbnail' failed: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	contentType := header.Header.Get("Content-Type")
	imageData, err := io.ReadAll(file)
	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		log.Printf("could not retrieve video given id %s", videoIDString)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if video.UserID != userID {
		log.Printf("video does not belong to this user.")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	videoThumbnails[video.ID] = thumbnail{data: imageData, mediaType: contentType}
	thumbnailURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoIDString)
	video.ThumbnailURL = &thumbnailURL
	cfg.db.UpdateVideo(video)
	respondWithJSON(w, http.StatusOK, video)
}

