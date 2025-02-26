package router

import (
	"fmt"
	"net/http"

	"sme-backend/src/core/middlewares"
	"sme-backend/src/core/version"
	admins_handler "sme-backend/src/modules/admins"
	appointments_handler "sme-backend/src/modules/appointments"
	auth_handler "sme-backend/src/modules/auth"
	bank_accounts_handler "sme-backend/src/modules/bank_accounts"
	commissions_handler "sme-backend/src/modules/commissons"
	"sme-backend/src/modules/notifications"
	"sme-backend/src/modules/payments"
	posts_handler "sme-backend/src/modules/posts"
	searches_handler "sme-backend/src/modules/searches"
	services_handler "sme-backend/src/modules/services"
	users_handler "sme-backend/src/modules/users"
	"sme-backend/src/modules/webhooks"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func SetupRoutes(app *chi.Mux) {
	// Do not change the order of middlewares.
	app.Use(middleware.RequestID)
	app.Use(middleware.StripSlashes)
	app.Use(middleware.RealIP)
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)

	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		request_id := middleware.GetReqID(r.Context())
		msg := fmt.Sprintf("Hello there, your request ID is: %s", request_id)
		w.Write([]byte(msg))
	})

	app.Get("/version", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(version.BUILD_VERSION))
	})

	app.Route("/admin", func(router chi.Router) {
		router.With(middlewares.PartiallyProtected).Post("/signin", auth_handler.AdminSignIn)
		router.With(middlewares.AdminProtected).Get("/users", admins_handler.GetPlatformUsers)
		router.With(middlewares.AdminProtected).Get("/users/joining-requests", admins_handler.GetNewEntrepreneursJoiningRequests)
		router.With(middlewares.AdminProtected).Put("/users/{joining_requestor_id}/approval", admins_handler.ApproveOrRejectEntrepreneur)
		router.With(middlewares.AdminProtected).Get("/payments/clearance-entrepreneurs", admins_handler.GetPaymentClearanceEntrepreneurs)
		router.With(middlewares.AdminProtected).Get("/payments/status/clear", admins_handler.GetClearingPaymentDetails)
		router.With(middlewares.AdminProtected).Put("/payments/status/clear", admins_handler.MarkAsPaymentCleared)

		router.With(middlewares.AdminProtected).Get("/analytics/joinings", admins_handler.GetWeeklyJoinedUsersCount)
		router.With(middlewares.AdminProtected).Get("/analytics/earnings", admins_handler.GetDailyWiseEarnings)

		// Commissons
		router.With(middlewares.AdminProtected).Get("/commissons", commissions_handler.GetCommissonsDetails)
		router.With(middlewares.AdminProtected).Post("/commissons", commissions_handler.UpdateCommissonsDetails)
	})

	app.With(middlewares.PartiallyProtected).Route("/auth", func(router chi.Router) {
		router.Post("/phone/verify", auth_handler.VerifyOTP)
		router.Post("/phone/send-otp", auth_handler.SendOTP)
		router.Post("/phone/resend", auth_handler.ResendOtp)
		router.Post("/signup", auth_handler.SignUp)
		router.Post("/{user_id}/upload/verification-docs", auth_handler.UploadVerificationDocs)
		router.Get("/jwt/verify", auth_handler.VerifyJwtTokenExpiration)
		router.Post("/signin/bypass", auth_handler.SignInByPass)
	})

	app.Route("/appointments", func(router chi.Router) {
		router.With(middlewares.Protected).Get("/price", appointments_handler.GetAppointmentPrice)
		router.With(middlewares.Protected).Post("/book", appointments_handler.BookAppointment)
		router.With(middlewares.Protected).Post("/{appointment_id}/completed", appointments_handler.MarkAsCompleted)
		router.With(middlewares.PartiallyProtected).Get("/available-days", appointments_handler.GetAppointmentsEnabledDays)
		router.With(middlewares.PartiallyProtected).Get("/available-timings", appointments_handler.GetAppointmentsEnabledDayTimings)
	})

	app.Route("/bank-accounts", func(router chi.Router) {
		router.With(middlewares.Protected).Get("/", bank_accounts_handler.GetBankAccount)
		router.With(middlewares.Protected).Post("/", bank_accounts_handler.CreateUpdateBankAccount)
	})

	app.With(middlewares.Protected).Route("/payments", func(router chi.Router) {
		router.Get("/verify/status", payments.VerifyPaymentStatus)
	})

	app.Route("/posts", func(router chi.Router) {
		router.With(middlewares.PartiallyProtected).Get("/", posts_handler.GetPosts)
		router.With(middlewares.Protected).Post("/", posts_handler.CreateUpdatePost)

		router.With(middlewares.Protected).Delete("/{post_id}", posts_handler.DeletePost)

		router.With(middlewares.Protected).Post("/{post_id}/medias", posts_handler.UploadPostMedia)
		router.With(middlewares.Protected).Delete("/{post_id}/medias/{media_id}", posts_handler.DeletePostMedia)
	})

	app.Route("/services", func(router chi.Router) {
		router.With(middlewares.PartiallyProtected).Get("/", services_handler.GetServices)
		router.With(middlewares.PartiallyProtected).Get("/{service_id}", services_handler.GetServiceByID)
		router.With(middlewares.Protected).Post("/", services_handler.CreateService)

		router.With(middlewares.Protected).Put("/{service_id}", services_handler.UpdateService)
		router.With(middlewares.Protected).Delete("/{service_id}", services_handler.DeleteService)

		router.With(middlewares.Protected).Post("/{service_id}/medias", services_handler.UploadServiceMedia)
		router.With(middlewares.Protected).Delete("/{service_id}/medias/{media_id}", services_handler.DeleteServiceMedia)

		router.With(middlewares.Protected).Get("/{service_id}/days", services_handler.GetServiceDays)
	})

	app.With(middlewares.Protected).Route("/services-days", func(router chi.Router) {
		router.Patch("/{service_day_id}", services_handler.ChangeServiceDayEnableStatus)
		router.Get("/{service_day_id}/timings", services_handler.GetServiceDayTimings)
		router.Put("/{service_day_id}/timings", services_handler.AddUpdateServiceTimings)
		router.Delete("/{service_day_id}/timings/{service_timing_id}", services_handler.DeleteServiceTimings)
	})

	app.With(middlewares.Protected).Route("/notifications", func(router chi.Router) {
		router.Put("/{noitifiction_id}/mark-as-read", notifications.MarkAsRead)
	})

	app.Route("/users", func(router chi.Router) {
		router.With(middlewares.Protected).Get("/appointments", users_handler.GetUserAppointments)
		router.With(middlewares.PartiallyProtected).Get("/{profile_id}/details", users_handler.GetUserDetailsByUserId)
		router.With(middlewares.Protected).Put("/details", users_handler.UpdateUserDetails)
		router.With(middlewares.Protected).Post("/photo/upload", users_handler.UploadPhoto)
		router.With(middlewares.Protected).Get("/notifications", notifications.GetUserNotifications)
		router.With(middlewares.Protected).Get("/notifications/unread-count", notifications.UnreadCount)
		router.With(middlewares.Protected).Get("/favourites", users_handler.GetFavouriteUsers)
		router.With(middlewares.Protected).Post("/{profile_id}/favourite", users_handler.AddUserToFavourite)
		router.With(middlewares.Protected).Delete("/{profile_id}/favourite", users_handler.RemoveUserFromFavourite)
	})

	app.With(middlewares.PartiallyProtected).Route("/searches", func(router chi.Router) {
		router.Get("/", searches_handler.SearchServicesAndPostsAndUsers)
	})

	app.With(middlewares.PartiallyProtected).Route("/webhooks", func(router chi.Router) {
		router.Post("/payment/razorpay", webhooks.RozarPayVerifyPayment)
	})

	// V1 API version
	app.Route(fmt.Sprintf("/%s", version.V1_API_VERSION), func(router chi.Router) {
		router.Route("/admin", func(router chi.Router) {
			router.With(middlewares.PartiallyProtected).Post("/signin", auth_handler.AdminSignIn)
			router.With(middlewares.AdminProtected).Get("/users", admins_handler.GetPlatformUsers)
			router.With(middlewares.AdminProtected).Get("/users/joining-requests", admins_handler.GetNewEntrepreneursJoiningRequests)
			router.With(middlewares.AdminProtected).Put("/users/{joining_requestor_id}/approval", admins_handler.ApproveOrRejectEntrepreneur)
			router.With(middlewares.AdminProtected).Get("/payments/clearance-entrepreneurs", admins_handler.GetPaymentClearanceEntrepreneurs)
			router.With(middlewares.AdminProtected).Get("/payments/status/clear", admins_handler.GetClearingPaymentDetails)
			router.With(middlewares.AdminProtected).Put("/payments/status/clear", admins_handler.MarkAsPaymentCleared)

			router.With(middlewares.AdminProtected).Get("/analytics/joinings", admins_handler.GetWeeklyJoinedUsersCount)
			router.With(middlewares.AdminProtected).Get("/analytics/earnings", admins_handler.GetDailyWiseEarnings)

			// Commissons
			router.With(middlewares.AdminProtected).Get("/commissons", commissions_handler.GetCommissonsDetails)
			router.With(middlewares.AdminProtected).Post("/commissons", commissions_handler.UpdateCommissonsDetails)
		})

		router.With(middlewares.PartiallyProtected).Route("/auth", func(router chi.Router) {
			router.Post("/phone/verify", auth_handler.VerifyOTP)
			router.Post("/phone/send-otp", auth_handler.SendOTP)
			router.Post("/phone/resend", auth_handler.ResendOtp)
			router.Post("/signup", auth_handler.SignUp)
			router.Post("/{user_id}/upload/verification-docs", auth_handler.UploadVerificationDocs)
			router.Get("/jwt/verify", auth_handler.VerifyJwtTokenExpiration)
			router.Post("/signin/bypass", auth_handler.SignInByPass)
		})

		router.Route("/appointments", func(router chi.Router) {
			router.With(middlewares.Protected).Get("/price", appointments_handler.GetAppointmentPrice)
			router.With(middlewares.Protected).Post("/book", appointments_handler.BookAppointment)
			router.With(middlewares.Protected).Post("/{appointment_id}/completed", appointments_handler.MarkAsCompleted)
			router.With(middlewares.PartiallyProtected).Get("/available-days", appointments_handler.GetAppointmentsEnabledDays)
			router.With(middlewares.PartiallyProtected).Get("/available-timings", appointments_handler.GetAppointmentsEnabledDayTimings)
		})

		router.Route("/bank-accounts", func(router chi.Router) {
			router.With(middlewares.Protected).Get("/", bank_accounts_handler.GetBankAccount)
			router.With(middlewares.Protected).Post("/", bank_accounts_handler.CreateUpdateBankAccount)
		})

		router.With(middlewares.Protected).Route("/payments", func(router chi.Router) {
			router.Get("/verify/status", payments.VerifyPaymentStatus)
		})

		router.Route("/posts", func(router chi.Router) {
			router.With(middlewares.PartiallyProtected).Get("/", posts_handler.GetPosts)
			router.With(middlewares.Protected).Post("/", posts_handler.CreateUpdatePost)

			router.With(middlewares.Protected).Delete("/{post_id}", posts_handler.DeletePost)

			router.With(middlewares.Protected).Post("/{post_id}/medias", posts_handler.UploadPostMedia)
			router.With(middlewares.Protected).Delete("/{post_id}/medias/{media_id}", posts_handler.DeletePostMedia)
		})

		router.Route("/services", func(router chi.Router) {
			router.With(middlewares.PartiallyProtected).Get("/", services_handler.GetServices)
			router.With(middlewares.PartiallyProtected).Get("/{service_id}", services_handler.GetServiceByID)
			router.With(middlewares.Protected).Post("/", services_handler.CreateService)

			router.With(middlewares.Protected).Put("/{service_id}", services_handler.UpdateService)
			router.With(middlewares.Protected).Delete("/{service_id}", services_handler.DeleteService)

			router.With(middlewares.Protected).Post("/{service_id}/medias", services_handler.UploadServiceMedia)
			router.With(middlewares.Protected).Delete("/{service_id}/medias/{media_id}", services_handler.DeleteServiceMedia)

			router.With(middlewares.Protected).Get("/{service_id}/days", services_handler.GetServiceDays)
		})

		router.With(middlewares.Protected).Route("/services-days", func(router chi.Router) {
			router.Patch("/{service_day_id}", services_handler.ChangeServiceDayEnableStatus)
			router.Get("/{service_day_id}/timings", services_handler.GetServiceDayTimings)
			router.Put("/{service_day_id}/timings", services_handler.AddUpdateServiceTimings)
			router.Delete("/{service_day_id}/timings/{service_timing_id}", services_handler.DeleteServiceTimings)
		})

		router.With(middlewares.Protected).Route("/notifications", func(router chi.Router) {
			router.Put("/{noitifiction_id}/mark-as-read", notifications.MarkAsRead)
		})

		router.Route("/users", func(router chi.Router) {
			router.With(middlewares.Protected).Get("/appointments", users_handler.GetUserAppointments)
			router.With(middlewares.PartiallyProtected).Get("/{profile_id}/details", users_handler.GetUserDetailsByUserId)
			router.With(middlewares.Protected).Put("/details", users_handler.UpdateUserDetails)
			router.With(middlewares.Protected).Post("/photo/upload", users_handler.UploadPhoto)
			router.With(middlewares.Protected).Get("/notifications", notifications.GetUserNotifications)
			router.With(middlewares.Protected).Get("/notifications/unread-count", notifications.UnreadCount)
			router.With(middlewares.Protected).Get("/favourites", users_handler.GetFavouriteUsers)
			router.With(middlewares.Protected).Post("/{profile_id}/favourite", users_handler.AddUserToFavourite)
			router.With(middlewares.Protected).Delete("/{profile_id}/favourite", users_handler.RemoveUserFromFavourite)
		})

		router.With(middlewares.PartiallyProtected).Route("/searches", func(router chi.Router) {
			router.Get("/", searches_handler.SearchServicesAndPostsAndUsers)
		})

		router.With(middlewares.PartiallyProtected).Route("/webhooks", func(router chi.Router) {
			router.Post("/payment/razorpay", webhooks.RozarPayVerifyPayment)
		})
	})
}
