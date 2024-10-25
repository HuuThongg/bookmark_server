package api

import (
	"bookmark/util"
	"bookmark/vultr"
	"io"
	"log"
	"net/http"
	"os"
)

func (h *API) UploadHeroImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	file, handler, err := r.FormFile("heroImg")
	if err != nil {
		log.Println(err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	defer file.Close()

	dst, err := os.Create(handler.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	imgFile, err := os.Open(handler.Filename)
	if err != nil {
		log.Printf("UploadHeroImage: failed to open file: %v", err)
		util.Response(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	heroImgScr := vultr.UploadHeroImage(imgFile)

	util.JsonResponse(w, heroImgScr)
}
