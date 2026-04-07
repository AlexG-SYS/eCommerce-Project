package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/AlexG-SYS/eCommerce-Project/internal/data"
	"golang.org/x/crypto/bcrypt"
)

func (h *Handler) CreateProfileHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email       string `json:"email"`
		Password    string `json:"password"`
		FullName    string `json:"full_name"`
		Phone       string `json:"phone"`
		Address     string `json:"address"`
		District    string `json:"district"`
		TownVillage string `json:"town_village"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	user := &data.User{Email: input.Email, Password: string(hashedPassword), Role: "Customer"}
	profile := &data.Profile{
		FullName:    input.FullName,
		Phone:       input.Phone,
		Address:     input.Address,
		District:    input.District,
		TownVillage: input.TownVillage,
	}

	//validate the input
	if errs := data.ValidateUser(user); len(errs) > 0 {
		h.App.ErrorJSON(w, http.StatusBadRequest, errs["error"])
		return
	}
	if errs := data.ValidateProfile(profile); len(errs) > 0 {
		h.App.ErrorJSON(w, http.StatusBadRequest, errs["error"])
		return
	}

	// Call the new Insert method that handles the transaction
	if err := h.Models.Users.Insert(user, profile); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"user": user, "profile": profile}, nil)

	// 1. Generate an activation token valid for 3 days
	token, err := h.Models.Tokens.GenerateToken(user.UserID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 2. Save the token to the database!
	err = h.Models.Tokens.Insert(token)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 3. Send the email in a background goroutine so the user gets an immediate response
	go func() {
		defer func() {
			if err := recover(); err != nil {
				h.App.Logger.Error("Email background worker panic", "error", err)
			}
		}()

		data := map[string]any{
			"activationToken": token.Plaintext,
			"userID":          user.UserID,
		}

		// 3.1 Attempt to send
		err := h.App.Mailer.Send(user.Email, "user_welcome.tmpl", data)

		// 3.2 LOG THE RESULT
		if err != nil {
			// Log the failure with details
			h.App.Logger.Error("Failed to send activation email",
				"error", err,
				"recipient", user.Email,
			)
		} else {
			// 3.3 LOG THE SUCCESS
			h.App.Logger.Info("Email sent successfully",
				"recipient", user.Email,
				"template", "user_welcome.tmpl",
			)
		}
	}()

	h.App.WriteJSON(w, http.StatusAccepted, map[string]string{"message": "check your email to activate account"}, nil)
}

func (h *Handler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusNotFound, "invalid record id")
		return
	}

	user, err := h.Models.Profile.Get(id)
	if err != nil {
		// Distinguish between a DB error and a missing record
		if err.Error() == "record not found" {
			h.App.ErrorJSON(w, http.StatusNotFound, "profile not found")
		} else {
			h.App.ServerError(w, r, err)
		}
		return
	}

	currentUser := h.App.ContextGetUser(r)

	// Security Check: Admin sees all, Customers see only their own.
	if currentUser.Role != "Admin" && user.UserID != currentUser.UserID {
		h.App.Logger.Warn("Unauthorized profile access attempt",
			"by_user", currentUser.UserID,
			"target_user", user.UserID,
		)
		h.App.ErrorJSON(w, http.StatusForbidden, "you can only access your own profile")
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"user": user}, nil)
}

func (h *Handler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Get ID from URL
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id < 1 {
		h.App.ErrorJSON(w, http.StatusNotFound, "invalid id")
		return
	}

	// 2. Fetch existing profile first
	// Note: You'll need a GetProfileByID method in your ProfileModel
	user, err := h.Models.Profile.Get(id)
	if err != nil {
		// If your Get method returns a specific "not found" error string
		if err.Error() == "record not found" {
			h.App.ErrorJSON(w, http.StatusNotFound, "profile not found")
		} else {
			h.App.ServerError(w, r, err)
		}
		return
	}

	profile := user.Profile

	// Admin can update anyone, Customers can ONLY update themselves
	currentUser := h.App.ContextGetUser(r)

	if currentUser.Role != "Admin" && user.UserID != currentUser.UserID {
		h.App.Logger.Warn("Unauthorized UPDATE attempt",
			"by_user", currentUser.UserID,
			"target_user", user.UserID)
		h.App.ErrorJSON(w, http.StatusForbidden, "you can only update your own profile")
		return
	}

	// 3. Read JSON into a temporary anonymous struct with pointers
	var input struct {
		FullName    *string `json:"full_name"`
		Phone       *string `json:"phone"`
		Address     *string `json:"address"`
		District    *string `json:"district"`
		TownVillage *string `json:"town_village"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	// 4. Update only the fields that were provided in the JSON
	if input.FullName != nil {
		profile.FullName = *input.FullName
	}
	if input.Phone != nil {
		profile.Phone = *input.Phone
	}
	if input.Address != nil {
		profile.Address = *input.Address
	}
	if input.District != nil {
		profile.District = *input.District
	}
	if input.TownVillage != nil {
		profile.TownVillage = *input.TownVillage
	}

	// 5. VALIDATE before saving
	vErrs := data.ValidateProfile(profile)
	if len(vErrs) > 0 {
		h.App.WriteJSON(w, http.StatusUnprocessableEntity, map[string]any{"errors": vErrs}, nil)
		return
	}

	// 6. Save the updated version
	if err := h.Models.Profile.Update(profile); err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]any{"user": user}, nil)
}

func (h *Handler) ActivateUserHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Parse the plaintext token from the query string
	qs := r.URL.Query()
	tokenPlaintext := qs.Get("token")

	// 2. Validate token format (e.g., length)
	if len(tokenPlaintext) != 26 {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid activation token L")
		return
	}

	// 3. Retrieve the user associated with the token hash
	user, err := h.Models.Tokens.GetForToken(data.ScopeActivation, tokenPlaintext)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusNotFound, "invalid or expired token")
		return
	}

	// 4. Update user status to activated
	user.Activated = true
	err = h.Models.Users.Update(user)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	// 5. Delete all activation tokens for this user
	err = h.Models.Tokens.DeleteAllForUser(data.ScopeActivation, user.UserID)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusOK, map[string]string{"message": "account activated"}, nil)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := h.App.ReadJSON(w, r, &input); err != nil {
		h.App.ErrorJSON(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	// 1. Retrieve user by email
	user, err := h.Models.Users.GetByEmail(input.Email)
	if err != nil {
		h.App.ErrorJSON(w, http.StatusUnauthorized, "invalid credentials email")
		return
	}

	// 2. Verify password hash
	match, err := user.PasswordMatches(input.Password)
	if err != nil || !match {
		h.App.ErrorJSON(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// 3. Check if account is activated
	if !user.Activated {
		h.App.ErrorJSON(w, http.StatusForbidden, "your account must be activated to login")
		return
	}

	// 4. Generate a persistent login token (valid for 24 hours)
	token, err := h.Models.Tokens.GenerateToken(user.UserID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	err = h.Models.Tokens.Insert(token)
	if err != nil {
		h.App.ServerError(w, r, err)
		return
	}

	h.App.WriteJSON(w, http.StatusCreated, map[string]any{"authentication_token": token.Plaintext, "user": user}, nil)
}
