package router

import (
	"bookmark/api"
	cm "bookmark/middleware"
	"bookmark/util"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func Router(l *zerolog.Logger, v *validator.Validate, db *pgxpool.Pool, config *util.Config, rdb *redis.Client) *chi.Mux {
	r := chi.NewRouter()

	a := api.NewAPI(l, v, db, config, rdb)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://bookmark-ui.vercel.app", "http://localhost:5173", "https://bookmarking.app", "https://ca94dd7c.bookmark-ui.pages.dev"}, // Allow both production and local dev origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true, // Allow credentials like cookies, tokens
		MaxAge:           300,  // Cache preflight response for 5 minutes
	}))

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.AllowContentEncoding("application/json", "application/x-www-form-urlencoded"))
	r.Use(middleware.CleanPath)
	r.Use(middleware.RedirectSlashes)
	r.Use(cm.LogUserLocationOnDailyRequest)
	r.Route("/public", func(r chi.Router) {
		r.Use(httprate.LimitByIP(100, time.Minute))
		r.Get("/proxy", a.ProxyHandler)
		r.Post("/checkIfIsAuthenticated", a.CheckIfIsAuthenticated)
		//
		// r.Post("/continueWithGoogle", a.ContinueWithGoogle)
		//
		r.Post("/refreshToken", a.RefreshToken)
		//
		// r.Post("/sendOTP", a.SendOTP)
		//
		// r.Post("/verifyOTP", a.VerifyOTP)
		//
		// r.Get("/getUserMessages", a.GetAllUserMessages)
		//
		// r.Get("/getLinksByAccountID/{accountID}", a.GetLinksByUserID)
		//
		r.Post("/requestResetPasswordLink", a.RequestResetPasswordLink)
		r.Patch("/updatePassword", a.UpdatePassword)
		r.Post("/uploadHeroImage", a.UploadHeroImage)
		// r.Get("/getCollectionAndInviterNames/{inviteToken}", h.GetCollectionAndInviterNames)
		// r.Post("/acceptInvite", a.AcceptInvite)

		r.Route("/account", func(r chi.Router) {
			// r.Post("/signup", a.SignUp)
			r.Post("/signup", a.NewAccount)
			// r.Post("/", a.ContinueWithGoogle)
			r.Post("/create", a.NewAccount)
			r.Get("/getAllAccounts", a.GetAllAccounts)
			r.Post("/signin", a.SignIn)
		})
	})

	r.Route("/private", func(r chi.Router) {
		r.Use(cm.AuthenticateRequest())
		r.Route("/getLinksAndFolders/{accountID}/{folderID}", func(r chi.Router) {
			r.Use(cm.AuthorizeReadRequestOnCollection())
			r.Get("/", a.GetLinksAndFolders)
		})
		r.Route("/folder", func(r chi.Router) {
			r.Route("/create", func(r chi.Router) {
				r.Use(cm.AuthorizeCreateFolderRequest())
				r.Post("/", a.CreateFolder)
			})
			r.Get("/all", a.GetAllFolders)
			r.Post("/new-child-folder", a.CreateChildFolder)
			r.Patch("/star", a.StarFolders)
			r.Patch("/unstar", a.UnstarFolders)
			r.Patch("/rename", a.RenameFolder)
			r.Patch("/moveFoldersToTrash", a.MoveFoldersToTrash)
			r.Patch("/moveFolders", a.MoveFolders)
			r.Patch("/moveFoldersToRoot", a.MoveFoldersToRoot)
			r.Patch("/toggle-folder-starred", a.ToggleFolderStarred)
			r.Patch("/restoreFoldersFromTrash", a.RestoreFoldersFromTrash)
			r.Delete("/deleteFoldersForever", a.DeleteFoldersForever)
			r.Get("/getRootFoldersByUserID", a.GetRootFolders)
			r.Get("/getFolderChildren/{folderID}/{accountID}", a.GetFolderChildren)
			r.Get("/getFolderAncestors/{folderID}", a.GetFolderAncestors)
			r.Get("/searchFolders/{query}", a.SearchFolders)
			r.Get("/getTreeFolders", a.GetSortedTreeFolders)
			r.Patch("/updateOrder", a.UpdateFolderSort)
		})
		r.Route("/link", func(r chi.Router) {
			r.Post("/add", a.AddLinkV2)
			r.Options("/add", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			r.Patch("/rename", a.RenameLink)
			r.Patch("/move", a.MoveLinks)
			r.Patch("/moveLinksToTrash", a.MoveLinksToTrash)
			r.Patch("/restoreLinksFromTrash", a.RestoreLinksFromTrash)
			r.Delete("/deleteLinksForever", a.DeleteLinksForever)
			r.Get("/getRootLinks/{accountID}", a.GetRootLinks)
			r.Get("/get_folder_links/{accountID}/{folderID}", a.GetFolderLinks)
			r.Get("/searchLinks/{query}", a.SearchLinks)
			r.Get("/all/{accountID}", a.GetAllLinks)
			r.Patch("/takeNote", a.AddNote)
			r.Patch("/changeTitle", a.ChangeLinkTitle)
			r.Patch("/changeLinkURL", a.ChangeLinkURL)
			r.Get("/{linkID}", a.GetSingleLink)
			r.Get("/getAllDeletedLinks", a.GetDeletedLinks)
			r.Patch("/changeLinkDesc", a.ChangeLinkDesc)
		})
		r.Route("/tag", func(r chi.Router) {
			r.Post("/", a.AddTag)
			r.Delete("/", a.DeleteTag)
			r.Get("/", a.GetTagByLinkId)
			r.Get("/stats", a.GetTagStats)
		})
		r.Route("/tags", func(r chi.Router) {
			r.Post("/", a.AddTagsForLinks)
		})

	})
	r.Get("/", a.Hello)
	return r
}
