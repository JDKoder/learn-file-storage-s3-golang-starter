package main

import (
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

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
	mimeType, _, err := mime.ParseMediaType(contentType)
	log.Printf("mimeType %s, %t", mimeType, mimeType == "image/png")
	if err != nil {
		log.Printf("Problem parsing media type with contentType %s", contentType)
		w.WriteHeader(http.StatusInternalServerError)
	}
	if mimeType != "image/jpeg" && mimeType != "image/png" {
		respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported thumbnail type %s", mimeType), nil)
		return
	}

	log.Printf("Content-Type: %s", mimeType)
	//imageData, err := io.ReadAll(file)
	//encodedImage := base64.StdEncoding.EncodeToString(imageData)
	//dataURL := fmt.Sprintf("data:%s;base64,%s",  contentType, encodedImage)
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
	//thumbnailURL := fmt.Sprintf("http://localhost:%s/api/thumbnails/%s", cfg.port, videoIDString)
	//video.ThumbnailURL = &thumbnailURL
	//video.ThumbnailURL = &dataURL
	imageType := strings.Split(contentType, "/")[1]
	newFileName := fmt.Sprintf("%s.%s", videoIDString, imageType)
	thumbnailFilePath := filepath.Join(cfg.assetsRoot, newFileName)
	createdFile, err := os.Create(thumbnailFilePath)
	if err != nil {
		log.Printf("File creation failed with file path, %s", thumbnailFilePath)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	written, err := io.Copy(createdFile, file)
	log.Printf("Wrote %d bytes", written)
	if err != nil {
		log.Printf("Somethign went wrong adding file, %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	thumbnailURL := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, newFileName)
	//videoThumbnails[video.ID] = thumbnail{data:  written, mediaType: contentType}

	video.ThumbnailURL = &thumbnailURL
	cfg.db.UpdateVideo(video)
	respondWithJSON(w, http.StatusOK, video)
}
