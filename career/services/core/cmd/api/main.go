package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/social-platform/services/core/internal/ai"
	"github.com/social-platform/services/core/internal/audit"
	"github.com/social-platform/services/core/internal/career"
	"github.com/social-platform/services/core/internal/identity"
	"github.com/social-platform/services/core/internal/linkedin"
	"github.com/social-platform/services/core/internal/policy"
	"github.com/social-platform/services/core/internal/provider"
	"github.com/social-platform/services/core/internal/quota"
	"github.com/social-platform/services/core/internal/search"
	"github.com/social-platform/services/core/internal/storage"
	"github.com/social-platform/services/core/internal/workflow"
	"github.com/social-platform/services/core/internal/workspace"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
}

type SectionMeta struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type SignupInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"fullName"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateWorkspaceInput struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type AddMemberInput struct {
	WorkspaceID string `json:"workspaceId"`
	UserID      string `json:"userId"`
	Role        string `json:"role"`
}

type CreateGrantInput struct {
	WorkspaceID  string `json:"workspaceId"`
	GranteeID    string `json:"granteeId"`
	ResourceType string `json:"resourceType"`
	ResourceID   string `json:"resourceId"`
	Permission   string `json:"permission"`
	DurationSec  int    `json:"durationSec"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev_jwt_secret_change_in_production_min_32_chars"
	}

	identityStore := identity.NewMemoryStore()
	authService := identity.NewAuthService(identityStore, jwtSecret)

	workspaceStore := workspace.NewMemoryStore()
	workspaceService := workspace.NewWorkspaceService(workspaceStore)

	storageDir := os.Getenv("STORAGE_DIR")
	if storageDir == "" {
		storageDir = "data/uploads"
	}
	storageProvider, err := storage.NewLocalStorageProvider(storageDir, []byte(jwtSecret), "http://localhost:"+port)
	if err != nil {
		log.Fatalf("[core-api] Failed to initialize storage provider: %v", err)
	}
	uploadRepo := storage.NewMemoryUploadRepository()
	uploadService := storage.NewUploadService(uploadRepo, storageProvider)

	// Workflow, Run Ledger and Outbox initialization
	runLedgerRepo := workflow.NewMemoryRunLedgerRepository()
	runLedgerService := workflow.NewRunLedgerService(runLedgerRepo)
	streamEngine := workflow.NewMemoryStreamEngine()
	outboxRepo := workflow.NewMemoryOutboxRepository()
	outboxService := workflow.NewOutboxService(outboxRepo)
	relayCtx, relayCancel := context.WithCancel(context.Background())
	defer relayCancel()
	outboxRelay := workflow.NewOutboxRelay(outboxRepo, streamEngine)
	go outboxRelay.Start(relayCtx, 500*time.Millisecond, 50)

	// Scheduler and SSE initialization (FND-008)
	schedulerRepo := workflow.NewMemorySchedulerRepository()
	schedulerService := workflow.NewSchedulerService(schedulerRepo)
	eventStore := workflow.NewMemoryRunEventStore()
	sseBroadcaster := workflow.NewSSEBroadcaster(eventStore)

	// Provider manifests and Credential Vault (FND-009)
	vaultKey := []byte(jwtSecret)
	if len(vaultKey) < 32 {
		vaultKey = append(vaultKey, make([]byte, 32-len(vaultKey))...)
	} else if len(vaultKey) > 32 {
		vaultKey = vaultKey[:32]
	}
	capabilityRegistry := provider.NewCapabilityRegistry()
	vaultRepo := provider.NewMemoryCredentialVaultRepository()
	vaultService, err := provider.NewCredentialVaultService(vaultRepo, vaultKey)
	if err != nil {
		log.Fatalf("[core-api] Failed to initialize credential vault: %v", err)
	}

	// Immutable Approval Gate (FND-010)
	approvalRepo := policy.NewMemoryApprovalRepository()
	approvalService := policy.NewApprovalService(approvalRepo)

	// Account-wide Budgets, Reservations and Nested Quotas (FND-011)
	quotaStore := quota.NewMemoryQuotaStore()
	quotaService := quota.NewQuotaService(quotaStore, capabilityRegistry)

	// Scoped Search Projections and Global Facade (FND-012)
	searchBackend := search.NewMemorySearchBackend()
	searchFacade := search.NewSearchFacade(searchBackend)

	// Audit, Notifications, Consent and Redacted Diagnostics (FND-013)
	auditService := audit.NewAuditService(vaultService, runLedgerService, searchFacade)

	// AI Gateway Service with Fact Grounding and Injection Defense (FND-015)
	aiGatewayService := ai.NewAIGatewayService(ai.NewMockDeterministicAdapter())

	// Career Service for Document Ingestion and Verified Facts (IMP-CAR-01)
	careerRepo := career.NewMemoryCareerRepository()
	careerService := career.NewCareerService(careerRepo)

	// LinkedIn Service for Capability Inspection and Integrations (IMP-LI-01)
	linkedinRepo := linkedin.NewMemoryRepository()
	linkedinService := linkedin.NewService(linkedinRepo)

	mux := http.NewServeMux()

	// 1. Health Probe
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "ok",
			Version:   "0.1.0",
			Timestamp: time.Now().UTC(),
		})
	})

	// 2. Sections Overview
	mux.HandleFunc("/api/v1/sections", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sections := []SectionMeta{
			{
				ID:          "career",
				Name:        "Career",
				Description: "Verified profile, ATS resume generation, job search, and application tracking",
			},
			{
				ID:          "linkedin",
				Name:        "LinkedIn",
				Description: "Profile analysis, genuine relationship outreach, and approved OIDC sync",
			},
			{
				ID:          "social",
				Name:        "Social",
				Description: "Content creation, campaign research, schedule planning, and approval inbox",
			},
		}
		_ = json.NewEncoder(w).Encode(sections)
	})

	// 3. Auth Routes
	mux.HandleFunc("/api/v1/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in SignupInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		res, err := authService.SignUp(r.Context(), in.Email, in.Password, in.FullName)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "SIGNUP_FAILED", "message": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("/api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in LoginInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		res, err := authService.Login(r.Context(), in.Email, in.Password, r.RemoteAddr, r.UserAgent())
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "AUTH_FAILED", "message": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	})

	mux.HandleFunc("/api/v1/auth/logout", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user != nil {
			_ = authService.RevokeAllSessions(r.Context(), user.ID)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "logged_out"})
	}))

	mux.HandleFunc("/api/v1/auth/me", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(user)
	}))

	// 4. Workspace & Resource Grant Routes (AT-011)
	mux.HandleFunc("/api/v1/workspaces", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if r.Method == http.MethodPost {
			var in CreateWorkspaceInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			ws, err := workspaceService.CreateWorkspace(r.Context(), user.ID, in.Name, in.Slug)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(ws)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/workspaces/members", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in AddMemberInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		role := identity.WorkspaceRole(in.Role)
		if role == "" {
			role = identity.WorkspaceRoleUser
		}
		err := workspaceService.AddMember(r.Context(), user.ID, in.WorkspaceID, in.UserID, role)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "FORBIDDEN", "message": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "member_added"})
	}))

	mux.HandleFunc("/api/v1/grants", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var in CreateGrantInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		duration := time.Duration(in.DurationSec) * time.Second
		grant, err := workspaceService.CreateGrant(
			r.Context(),
			user.ID,
			in.WorkspaceID,
			in.GranteeID,
			in.ResourceType,
			in.ResourceID,
			workspace.Permission(in.Permission),
			duration,
		)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "GRANT_FAILED", "message": err.Error()})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(grant)
	}))

	// 5. Upload & Storage Routes (FND-006)
	mux.HandleFunc("/api/v1/uploads/request", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req storage.UploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		resp, err := uploadService.RequestUpload(r.Context(), user.ID, req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "UPLOAD_REQUEST_FAILED", "message": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/api/v1/uploads/commit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		uploadID := r.URL.Query().Get("upload_id")
		if uploadID == "" {
			http.Error(w, "Missing upload_id parameter", http.StatusBadRequest)
			return
		}

		rec, err := uploadService.ProcessAndCommitUpload(r.Context(), uploadID, r.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "COMMIT_FAILED", "message": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rec)
	}))

	mux.HandleFunc("/api/v1/uploads/download", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		uploadID := r.URL.Query().Get("upload_id")
		if uploadID == "" {
			http.Error(w, "Missing upload_id parameter", http.StatusBadRequest)
			return
		}

		resp, err := uploadService.GetDownloadURL(r.Context(), user.ID, uploadID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "DOWNLOAD_DENIED", "message": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	mux.HandleFunc("/api/v1/uploads/delete", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		uploadID := r.URL.Query().Get("upload_id")
		if uploadID == "" {
			http.Error(w, "Missing upload_id parameter", http.StatusBadRequest)
			return
		}

		if err := uploadService.DeleteUpload(r.Context(), user.ID, uploadID); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "DELETE_FAILED", "message": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	}))

	// Direct Local Storage Handlers with Signed Tokens
	mux.HandleFunc("/api/v1/storage/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		token := r.URL.Query().Get("token")
		if err := storageProvider.ValidateSignedToken(token, "upload", ""); err != nil {
			http.Error(w, "Invalid or expired storage token", http.StatusForbidden)
			return
		}

		parts := strings.Split(token, ":")
		key := parts[1]

		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		if err := storageProvider.Put(r.Context(), key, r.Body, r.ContentLength, contentType); err != nil {
			http.Error(w, fmt.Sprintf("Storage write failed: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"uploaded"}`))
	})

	mux.HandleFunc("/api/v1/storage/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		token := r.URL.Query().Get("token")
		if err := storageProvider.ValidateSignedToken(token, "download", ""); err != nil {
			http.Error(w, "Invalid or expired storage token", http.StatusForbidden)
			return
		}

		parts := strings.Split(token, ":")
		key := parts[1]

		rc, size, contentType, err := storageProvider.Get(r.Context(), key)
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		defer rc.Close()

		filename := r.URL.Query().Get("filename")
		if filename == "" {
			filename = storage.SanitizeFilename(key)
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", size))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		_, _ = io.Copy(w, rc)
	})

	// 5. Workflow Runs API (FND-007)
	mux.HandleFunc("/api/v1/runs", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())

		if r.Method == http.MethodPost {
			var input struct {
				WorkspaceID    string                 `json:"workspaceId"`
				ActionType     string                 `json:"actionType"`
				TargetURL      string                 `json:"targetUrl"`
				Parameters     map[string]interface{} `json:"parameters"`
				IdempotencyKey string                 `json:"idempotencyKey"`
			}
			if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
				http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
				return
			}
			if input.WorkspaceID == "" || input.ActionType == "" || input.IdempotencyKey == "" {
				http.Error(w, "workspaceId, actionType, and idempotencyKey are required", http.StatusBadRequest)
				return
			}

			rawParams, _ := json.Marshal(input.Parameters)

			actionRun := &workflow.ActionRun{
				WorkspaceID:    input.WorkspaceID,
				ActorID:        user.ID,
				TaskType:       input.ActionType,
				InputPayload:   rawParams,
				IdempotencyKey: input.IdempotencyKey,
			}

			run, isNew, err := runLedgerService.RegisterOrGetRun(r.Context(), actionRun)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to register run: %v", err), http.StatusInternalServerError)
				return
			}

			if isNew {
				// Enqueue to outbox for durable streaming
				_, _ = outboxService.RecordEvent(r.Context(), "action_run", run.ID, input.ActionType, rawParams)
			}

			w.Header().Set("Content-Type", "application/json")
			if isNew {
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			_ = json.NewEncoder(w).Encode(run)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/runs/", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		runID := strings.TrimPrefix(r.URL.Path, "/api/v1/runs/")
		if runID == "" {
			http.Error(w, "Run ID required", http.StatusBadRequest)
			return
		}

		run, err := runLedgerService.GetRun(r.Context(), runID)
		if err != nil {
			if err == workflow.ErrRunNotFound {
				http.Error(w, "Run not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		attempts, _ := runLedgerService.GetAttempts(r.Context(), runID)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"run":      run,
			"attempts": attempts,
		})
	}))

	// 6. Pause, Resume, Cancel & SSE Routes (FND-008)
	mux.HandleFunc("/api/v1/runs/control", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var input struct {
			RunID  string `json:"runId"`
			Action string `json:"action"` // "pause", "resume", "cancel"
			Reason string `json:"reason,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid input: "+err.Error(), http.StatusBadRequest)
			return
		}
		if input.RunID == "" || input.Action == "" {
			http.Error(w, "runId and action are required", http.StatusBadRequest)
			return
		}

		var updated *workflow.ActionRun
		var err error
		switch input.Action {
		case "pause":
			updated, err = runLedgerService.PauseRun(r.Context(), input.RunID)
			if err == nil {
				_, _ = sseBroadcaster.Broadcast(r.Context(), input.RunID, "paused", json.RawMessage(`{"state":"paused"}`))
			}
		case "resume":
			updated, err = runLedgerService.ResumeRun(r.Context(), input.RunID)
			if err == nil {
				_, _ = sseBroadcaster.Broadcast(r.Context(), input.RunID, "resumed", json.RawMessage(`{"state":"queued"}`))
			}
		case "cancel":
			updated, err = runLedgerService.CancelRun(r.Context(), input.RunID, input.Reason)
			if err == nil {
				_, _ = sseBroadcaster.Broadcast(r.Context(), input.RunID, "cancelled", json.RawMessage(`{"state":"cancelled"}`))
			}
		default:
			http.Error(w, "Unknown action: "+input.Action, http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// SSE Live Run Events with Reconnection Replay (REQ-019)
	mux.HandleFunc("/api/v1/runs/events", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		runID := r.URL.Query().Get("runId")
		if runID == "" {
			http.Error(w, "runId query param required", http.StatusBadRequest)
			return
		}

		lastEventID := r.Header.Get("Last-Event-ID")
		if lastEventID == "" {
			lastEventID = r.URL.Query().Get("lastEventId")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		ch, history, cleanup, err := sseBroadcaster.Subscribe(r.Context(), runID, lastEventID, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cleanup()

		// 1. Replay historical events since Last-Event-ID
		for _, evt := range history {
			_, _ = w.Write(workflow.FormatSSE(evt))
		}
		flusher.Flush()

		// 2. Stream live events until client disconnects
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				_, _ = w.Write(workflow.FormatKeepalive())
				flusher.Flush()
			case evt, ok := <-ch:
				if !ok {
					return
				}
				_, _ = w.Write(workflow.FormatSSE(evt))
				flusher.Flush()
			}
		}
	})

	// 7. Schedules API (FND-008)
	mux.HandleFunc("/api/v1/schedules", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var in struct {
				WorkspaceID string                 `json:"workspaceId"`
				TaskType    string                 `json:"taskType"`
				Payload     map[string]interface{} `json:"payload"`
				IntervalSec int                    `json:"intervalSec"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			if in.WorkspaceID == "" || in.TaskType == "" {
				http.Error(w, "workspaceId and taskType are required", http.StatusBadRequest)
				return
			}

			payloadBytes, _ := json.Marshal(in.Payload)
			interval := time.Duration(in.IntervalSec) * time.Second
			if interval <= 0 {
				interval = 1 * time.Hour
			}

			job, err := schedulerService.RegisterSchedule(r.Context(), &workflow.ScheduleJob{
				WorkspaceID: in.WorkspaceID,
				TaskType:    in.TaskType,
				Payload:     payloadBytes,
				Interval:    interval,
				NextRunAt:   time.Now().UTC(),
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(job)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 8. Providers and Credential Vault API (FND-009)
	mux.HandleFunc("/api/v1/providers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		manifests := capabilityRegistry.ListManifests()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(manifests)
	})

	mux.HandleFunc("/api/v1/credentials", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())

		if r.Method == http.MethodPost {
			var in struct {
				WorkspaceID       string   `json:"workspaceId"`
				ProviderID        string   `json:"providerId"`
				AccountIdentifier string   `json:"accountIdentifier"`
				Secret            string   `json:"secret"`
				Scopes            []string `json:"scopes"`
				ExpiresInSec      int      `json:"expiresInSec,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			if in.WorkspaceID == "" || in.ProviderID == "" || in.Secret == "" {
				http.Error(w, "workspaceId, providerId, and secret are required", http.StatusBadRequest)
				return
			}

			// Validate provider exists
			if _, err := capabilityRegistry.GetManifest(in.ProviderID); err != nil {
				http.Error(w, "Unsupported provider: "+in.ProviderID, http.StatusBadRequest)
				return
			}

			var expiresAt *time.Time
			if in.ExpiresInSec > 0 {
				t := time.Now().UTC().Add(time.Duration(in.ExpiresInSec) * time.Second)
				expiresAt = &t
			}

			cred, err := vaultService.StoreCredential(
				r.Context(),
				in.WorkspaceID,
				user.ID,
				in.ProviderID,
				in.AccountIdentifier,
				in.Secret,
				in.Scopes,
				expiresAt,
			)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to store credential: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(cred)
			return
		}

		if r.Method == http.MethodGet {
			workspaceID := r.URL.Query().Get("workspaceId")
			if workspaceID == "" {
				http.Error(w, "workspaceId query param required", http.StatusBadRequest)
				return
			}

			creds, err := vaultService.ListWorkspaceCredentials(r.Context(), workspaceID, user.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(creds)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/credentials/", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())

		if r.Method == http.MethodDelete {
			id := strings.TrimPrefix(r.URL.Path, "/api/v1/credentials/")
			if id == "" {
				http.Error(w, "Credential ID required", http.StatusBadRequest)
				return
			}

			err := vaultService.RevokeCredential(r.Context(), id, user.ID)
			if err != nil {
				if err == provider.ErrAccessDenied {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				if err == provider.ErrCredentialNotFound {
					http.Error(w, "Credential not found", http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "revoked"})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 9. Immutable Approval Gate API (FND-010)
	mux.HandleFunc("/api/v1/approvals/request", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var in struct {
			WorkspaceID     string                 `json:"workspaceId"`
			ActionType      string                 `json:"actionType"`
			TargetRecipient string                 `json:"targetRecipient"`
			Payload         map[string]interface{} `json:"payload"`
			DocumentHash    string                 `json:"documentHash,omitempty"`
			TTLSec          int                    `json:"ttlSec,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		if in.WorkspaceID == "" || in.ActionType == "" {
			http.Error(w, "workspaceId and actionType are required", http.StatusBadRequest)
			return
		}

		rawPayload, _ := json.Marshal(in.Payload)
		ttl := time.Duration(in.TTLSec) * time.Second
		if ttl <= 0 {
			ttl = 24 * time.Hour
		}

		req, err := approvalService.CreateApprovalRequest(r.Context(), policy.RequestApprovalInput{
			WorkspaceID:     in.WorkspaceID,
			ActorID:         user.ID,
			ActionType:      in.ActionType,
			TargetRecipient: in.TargetRecipient,
			Payload:         rawPayload,
			DocumentHash:    in.DocumentHash,
			TTL:             ttl,
		})
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create approval request: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(req)
	}))

	mux.HandleFunc("/api/v1/approvals/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var in struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ID == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		req, err := approvalService.Approve(r.Context(), in.ID, user.ID)
		if err != nil {
			if err == policy.ErrApprovalExpired {
				http.Error(w, "Approval request has expired", http.StatusGone)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(req)
	}))

	mux.HandleFunc("/api/v1/approvals/reject", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var in struct {
			ID     string `json:"id"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.ID == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		req, err := approvalService.Reject(r.Context(), in.ID, user.ID, in.Reason)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(req)
	}))

	mux.HandleFunc("/api/v1/approvals", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspaceId")
		if workspaceID == "" {
			http.Error(w, "workspaceId query param required", http.StatusBadRequest)
			return
		}

		list, err := approvalService.ListApprovals(r.Context(), workspaceID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(list)
	}))

	// 17. Quota and Nested Budget Endpoints (FND-011)
	mux.Handle("/api/v1/quotas/reserve", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req quota.ReservationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		res, err := quotaService.Reserve(r.Context(), req)
		if err != nil {
			status := http.StatusInternalServerError
			switch err {
			case quota.ErrQuotaExceeded, quota.ErrNestedQuotaExceeded:
				status = http.StatusTooManyRequests
			case quota.ErrRateLimited:
				status = http.StatusTooManyRequests
			case quota.ErrActionCapabilityUnauthorized:
				status = http.StatusForbidden
			case quota.ErrAccountPaused, quota.ErrChallengeRequired:
				status = http.StatusServiceUnavailable
			case quota.ErrStoreUnavailable:
				status = http.StatusServiceUnavailable
			}
			http.Error(w, err.Error(), status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(res)
	}))

	mux.Handle("/api/v1/quotas/settle", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			ReservationID string `json:"reservation_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ReservationID == "" {
			http.Error(w, "reservation_id required", http.StatusBadRequest)
			return
		}

		if err := quotaService.Settle(r.Context(), body.ReservationID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "settled"})
	}))

	mux.Handle("/api/v1/quotas/release", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			ReservationID string `json:"reservation_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ReservationID == "" {
			http.Error(w, "reservation_id required", http.StatusBadRequest)
			return
		}

		if err := quotaService.Release(r.Context(), body.ReservationID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "released"})
	}))

	mux.Handle("/api/v1/quotas/retain", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			ReservationID string `json:"reservation_id"`
			Reason        string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ReservationID == "" {
			http.Error(w, "reservation_id required", http.StatusBadRequest)
			return
		}

		if err := quotaService.RetainAmbiguous(r.Context(), body.ReservationID, body.Reason); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "retained"})
	}))

	mux.Handle("/api/v1/quotas/usage", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		accountID := r.URL.Query().Get("accountId")
		metric := quota.ActionMetric(r.URL.Query().Get("metric"))
		if accountID == "" || metric == "" {
			http.Error(w, "accountId and metric query parameters required", http.StatusBadRequest)
			return
		}

		summary, err := quotaService.GetUsageSummary(r.Context(), accountID, metric)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(summary)
	}))

	mux.Handle("/api/v1/accounts/status", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			AccountID         string `json:"account_id"`
			Provider          string `json:"provider"`
			Action            string `json:"action"` // pause, resume, challenge, rate_limit
			Reason            string `json:"reason"`
			RetryAfterSeconds int    `json:"retry_after_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.AccountID == "" {
			http.Error(w, "account_id required", http.StatusBadRequest)
			return
		}

		var err error
		switch body.Action {
		case "pause":
			err = quotaService.PauseAccount(r.Context(), body.AccountID, body.Provider, body.Reason)
		case "resume":
			err = quotaService.ResumeAccount(r.Context(), body.AccountID, body.Provider)
		case "challenge":
			err = quotaService.RecordChallenge(r.Context(), body.AccountID, body.Provider, body.Reason)
		case "rate_limit":
			retryAfter := time.Now().UTC().Add(time.Duration(body.RetryAfterSeconds) * time.Second)
			err = quotaService.RecordRateLimit(r.Context(), body.AccountID, body.Provider, retryAfter)
		default:
			http.Error(w, "Invalid action: must be pause, resume, challenge or rate_limit", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	}))

	// 18. Scoped Search Projections and Global Facade (FND-012)
	mux.Handle("/api/v1/search/query", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req search.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid search request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		result, err := searchFacade.Search(r.Context(), req)
		if err != nil {
			status := http.StatusInternalServerError
			if err == search.ErrWorkspaceScopeRequired {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))

	mux.Handle("/api/v1/search/autocomplete", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspaceId")
		index := search.IndexName(r.URL.Query().Get("index"))
		prefix := r.URL.Query().Get("prefix")
		field := r.URL.Query().Get("field")

		if workspaceID == "" || index == "" || field == "" {
			http.Error(w, "workspaceId, index and field query parameters are required", http.StatusBadRequest)
			return
		}

		result, err := searchFacade.Autocomplete(r.Context(), search.AutocompleteRequest{
			WorkspaceID: workspaceID,
			Index:       index,
			Prefix:      prefix,
			Field:       field,
			Limit:       10,
		})
		if err != nil {
			status := http.StatusInternalServerError
			if err == search.ErrWorkspaceScopeRequired {
				status = http.StatusBadRequest
			}
			http.Error(w, err.Error(), status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))

	mux.Handle("/api/v1/search/diagnostics", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		index := search.IndexName(r.URL.Query().Get("index"))
		if index == "" {
			index = search.IndexJobs
		}

		diag, err := searchFacade.GetDiagnostics(r.Context(), index)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(diag)
	}))

	mux.Handle("/api/v1/search/project", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Index    search.IndexName       `json:"index"`
			Document map[string]interface{} `json:"document"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := searchFacade.ProjectDocument(r.Context(), body.Index, body.Document); err != nil {
			status := http.StatusBadRequest
			if err == search.ErrRawDocumentIndexProhibited {
				status = http.StatusForbidden
			}
			http.Error(w, err.Error(), status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "projected"})
	}))

	mux.Handle("/api/v1/search/delete", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Index       search.IndexName `json:"index"`
			WorkspaceID string           `json:"workspace_id"`
			DocumentID  string           `json:"document_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := searchFacade.DeleteDocument(r.Context(), body.Index, body.WorkspaceID, body.DocumentID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
	}))

	// 14. Audit, Notifications, Consent and Account Deletion Routes (FND-013)
	mux.Handle("/api/v1/audit/events", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		actorID := r.URL.Query().Get("actor_id")
		limit := 50
		if lStr := r.URL.Query().Get("limit"); lStr != "" {
			if parsedLimit, err := strconv.Atoi(lStr); err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		events, err := auditService.ListEvents(r.Context(), workspaceID, actorID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if limit > 0 && len(events) > limit {
			events = events[:limit]
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(events)
	}))

	mux.Handle("/api/v1/notifications", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		userID := r.URL.Query().Get("user_id")

		notifications, err := auditService.ListNotifications(r.Context(), workspaceID, userID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(notifications)
	}))

	mux.Handle("/api/v1/notifications/preferences", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var prefs audit.NotificationPreferences
		if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
			http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := auditService.UpdatePreferences(r.Context(), prefs); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	}))

	mux.Handle("/api/v1/consent/support", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			WorkspaceID     string `json:"workspace_id"`
			GranterUserID   string `json:"granter_user_id"`
			SupportUserID   string `json:"support_user_id"`
			Scope           string `json:"scope"`
			DurationSeconds int    `json:"duration_seconds"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		expiresAt := time.Now().UTC().Add(2 * time.Hour)
		if body.DurationSeconds > 0 {
			expiresAt = time.Now().UTC().Add(time.Duration(body.DurationSeconds) * time.Second)
		}
		consent, err := auditService.GrantSupportConsent(r.Context(), audit.SupportConsent{
			WorkspaceID:   body.WorkspaceID,
			GranterUserID: body.GranterUserID,
			SupportUserID: body.SupportUserID,
			Scope:         body.Scope,
			ExpiresAt:     expiresAt,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(consent)
	}))

	mux.Handle("/api/v1/accounts/delete", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			WorkspaceID      string   `json:"workspace_id"`
			ActorID          string   `json:"actor_id"`
			AccountID        string   `json:"account_id"`
			AssociatedDocIDs []string `json:"associated_doc_ids"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		report, err := auditService.CascadeAccountDeletion(
			r.Context(),
			body.WorkspaceID,
			body.ActorID,
			body.AccountID,
			body.AssociatedDocIDs,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// 17. AI Gateway Endpoints (FND-015: REQ-016, REQ-021, AT-003, AT-019)
	mux.HandleFunc("/api/v1/ai/models", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		models, _ := aiGatewayService.ListModels(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(models)
	})

	mux.HandleFunc("/api/v1/ai/estimate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req ai.GenerationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}
		estimate, err := aiGatewayService.Estimate(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(estimate)
	}))

	mux.HandleFunc("/api/v1/ai/generate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req ai.GenerationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		res, err := aiGatewayService.Generate(r.Context(), req)
		if err != nil {
			status := http.StatusInternalServerError
			switch err {
			case ai.ErrModelConsentRequired:
				status = http.StatusForbidden
			case ai.ErrFabricatedFactsDetected:
				status = http.StatusUnprocessableEntity
			case ai.ErrNeedsUserInput:
				status = http.StatusUnprocessableEntity
			case ai.ErrPromptInjectionBlocked:
				status = http.StatusBadRequest
			case ai.ErrBudgetExceeded:
				status = http.StatusPaymentRequired
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":   err.Error(),
				"status":  status,
				"taskKey": req.TaskKey,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// 18. Career Profile & Document Ingestion Routes (IMP-CAR-01, REQ-002, AT-001, AT-020)
	mux.HandleFunc("/api/v1/career/upload", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Support multipart/form-data upload (up to 10MB default)
		if err := r.ParseMultipartForm(10 * 1024 * 1024); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to parse upload form or file exceeds 10MB"})
			return
		}

		file, handler, err := r.FormFile("file")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Missing 'file' field in multipart form"})
			return
		}
		defer file.Close()

		fileData, err := io.ReadAll(file)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to read file data"})
			return
		}

		mimeType := handler.Header.Get("Content-Type")
		if mimeType == "" || mimeType == "application/octet-stream" {
			if strings.HasSuffix(strings.ToLower(handler.Filename), ".pdf") {
				mimeType = "application/pdf"
			} else if strings.HasSuffix(strings.ToLower(handler.Filename), ".docx") {
				mimeType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
			}
		}

		draft, err := careerService.ProcessUploadedDocument(
			r.Context(),
			user.ID,
			fmt.Sprintf("up_%d", time.Now().UnixNano()),
			handler.Filename,
			mimeType,
			fileData,
		)
		if err != nil {
			status := http.StatusBadRequest
			if strings.Contains(err.Error(), "quarantine") {
				status = http.StatusUnprocessableEntity
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(draft)
	}))

	mux.HandleFunc("/api/v1/career/drafts/", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Pattern: /api/v1/career/drafts/{id} or /api/v1/career/drafts/{id}/confirm
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/career/drafts/")
		parts := strings.Split(strings.Trim(path, "/"), "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Draft ID required", http.StatusBadRequest)
			return
		}
		draftID := parts[0]

		if len(parts) == 2 && parts[1] == "confirm" {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			var in career.ConfirmDraftInput
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&in)
			}
			facts, err := careerService.ConfirmDraft(r.Context(), user.ID, draftID, in)
			if err != nil {
				status := http.StatusBadRequest
				if err == career.ErrDraftNotFound {
					status = http.StatusNotFound
				} else if err == career.ErrUnauthorized {
					status = http.StatusForbidden
				} else if err == career.ErrAlreadyConfirmed {
					status = http.StatusConflict
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":          "confirmed",
				"confirmed_facts": facts,
			})
			return
		}

		if r.Method == http.MethodGet {
			draft, err := careerService.GetExtractionDraft(r.Context(), user.ID, draftID)
			if err != nil {
				status := http.StatusNotFound
				if err == career.ErrUnauthorized {
					status = http.StatusForbidden
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(status)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(draft)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/career/facts", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			category := r.URL.Query().Get("category")
			facts, err := careerService.GetUserFacts(r.Context(), user.ID, category)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(facts)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/career/facts/manual", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in career.CreateManualFactInput
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
			return
		}

		fact, err := careerService.CreateManualFact(r.Context(), user.ID, in)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(fact)
	}))

	// Master Career Profile Endpoints (IMP-CAR-02, REQ-002, REQ-003, REQ-018)
	mux.HandleFunc("/api/v1/career/profile", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			profile, err := careerService.GetMasterProfile(r.Context(), user.ID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(profile)
			return
		}

		if r.Method == http.MethodPut {
			var p career.MasterCareerProfile
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid profile JSON payload"})
				return
			}
			p.UserID = user.ID
			saved, err := careerService.SaveMasterProfile(r.Context(), user.ID, &p, user.ID, "user update via API")
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(saved)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/career/profile/section/", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPatch && r.Method != http.MethodPut && r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		section := strings.TrimPrefix(r.URL.Path, "/api/v1/career/profile/section/")
		if section == "" {
			http.Error(w, "Section name required", http.StatusBadRequest)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		updated, err := careerService.UpdateProfileSection(r.Context(), user.ID, section, bodyBytes, user.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	mux.HandleFunc("/api/v1/career/profile/consent", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var consent career.UserCareerConsent
		if err := json.NewDecoder(r.Body).Decode(&consent); err != nil {
			http.Error(w, "Invalid consent payload", http.StatusBadRequest)
			return
		}

		updated, err := careerService.UpdateCareerConsent(r.Context(), user.ID, consent)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	mux.HandleFunc("/api/v1/career/profile/privacy", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var priv career.CareerPrivacySettings
		if err := json.NewDecoder(r.Body).Decode(&priv); err != nil {
			http.Error(w, "Invalid privacy payload", http.StatusBadRequest)
			return
		}

		updated, err := careerService.UpdatePrivacySettings(r.Context(), user.ID, priv)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	mux.HandleFunc("/api/v1/career/profile/admin/", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		targetUserID := strings.TrimPrefix(r.URL.Path, "/api/v1/career/profile/admin/")
		if targetUserID == "" {
			http.Error(w, "Target User ID required", http.StatusBadRequest)
			return
		}

		// Enforce REQ-018: Workspace Admin check
		workspaceID := r.URL.Query().Get("workspaceId")
		isWorkspaceAdmin := user.PlatformRole == identity.PlatformRoleSuperAdmin
		if !isWorkspaceAdmin && workspaceID != "" {
			isWorkspaceAdmin = workspaceService.IsAdmin(r.Context(), workspaceID, user.ID)
		}
		profile, err := careerService.GetMasterProfileWithPrivacyCheck(r.Context(), user.ID, targetUserID, isWorkspaceAdmin)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if errors.Is(err, career.ErrPrivateProfileAccessDenied) {
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]string{
					"code":    "PRIVATE_PROFILE_ISOLATED",
					"message": "Personal career profile is isolated. Admin access to workspace does not grant access (REQ-018).",
				})
				return
			}
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(profile)
	}))

	// Career Preferences & Exclusions Endpoints (IMP-CAR-03, REQ-003, AT-003)
	mux.HandleFunc("/api/v1/career/preferences", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			pref, err := careerService.GetCareerPreferences(r.Context(), user.ID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(pref)
			return
		}

		if r.Method == http.MethodPut {
			var pref career.CareerPreferences
			if err := json.NewDecoder(r.Body).Decode(&pref); err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid preferences JSON payload"})
				return
			}
			pref.UserID = user.ID
			saved, err := careerService.SaveCareerPreferences(r.Context(), user.ID, &pref)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(saved)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/career/preferences/check-exclusion", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			CompanyName string `json:"company_name"`
			JobTitle    string `json:"job_title"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}

		excluded, reason, err := careerService.CheckJobMatchExclusion(r.Context(), user.ID, in.CompanyName, in.JobTitle, in.Description)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"excluded": excluded,
			"reason":   reason,
		})
	}))

	// Master Resume Generation & AT-002 Parse-Back Verification Endpoints (IMP-CAR-04, CAR-04, REQ-003, AT-002)
	mux.HandleFunc("/api/v1/career/resume/generate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Format   career.MasterResumeFormat `json:"format"`   // "pdf", "docx", "txt"
			Template career.ResumeTemplateType `json:"template"` // "single_column_modern", "single_column_classic", "single_column_minimal"
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request JSON payload", http.StatusBadRequest)
			return
		}

		if in.Format == "" {
			in.Format = career.ResumeFormatPDF
		}
		if in.Template == "" {
			in.Template = career.TemplateSingleColumnModern
		}

		res, vResult, err := careerService.GenerateMasterResume(r.Context(), user.ID, in.Format, in.Template)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":        err.Error(),
				"verification": vResult,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"resume":       res,
			"verification": vResult,
		})
	}))

	mux.HandleFunc("/api/v1/career/resumes", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		resumes, err := careerService.ListGeneratedResumes(r.Context(), user.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resumes)
	}))

	mux.HandleFunc("/api/v1/career/resume/download", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		resumeID := r.URL.Query().Get("id")
		if resumeID == "" {
			http.Error(w, "Missing resume id parameter", http.StatusBadRequest)
			return
		}

		res, err := careerService.GetGeneratedResume(r.Context(), user.ID, resumeID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", res.MimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, res.FileName))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(res.ByteContent)))
		_, _ = w.Write(res.ByteContent)
	}))

	mux.HandleFunc("/api/v1/career/resume/verify-parse-back", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			ResumeID string `json:"resume_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid body", http.StatusBadRequest)
			return
		}

		vResult, err := careerService.VerifyResumeParseBack(r.Context(), user.ID, in.ResumeID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":        err.Error(),
				"verification": vResult,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(vResult)
	}))

	// Job-Specific Tailored Resume and Cover Letter Endpoints (IMP-CAR-05, REQ-016, AT-003, FND-010, AT-007)
	mux.HandleFunc("/api/v1/career/tailor/resume", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			resumeID := r.URL.Query().Get("id")
			if resumeID == "" {
				http.Error(w, "Missing resume id parameter", http.StatusBadRequest)
				return
			}
			res, err := careerService.GetTailoredResume(r.Context(), user.ID, resumeID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(res)
			return
		}

		if r.Method == http.MethodPost {
			var in struct {
				Job      career.JobTarget          `json:"job"`
				Format   career.MasterResumeFormat `json:"format"`
				Template career.ResumeTemplateType `json:"template"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request JSON payload", http.StatusBadRequest)
				return
			}
			if in.Format == "" {
				in.Format = career.ResumeFormatPDF
			}
			if in.Template == "" {
				in.Template = career.TemplateSingleColumnModern
			}

			tailored, err := careerService.CreateTailoredResume(r.Context(), user.ID, in.Job, in.Format, in.Template)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(tailored)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/career/tailor/resume/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			ResumeID      string `json:"resume_id"`
			ApprovalToken string `json:"approval_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		approved, err := careerService.ApproveTailoredResume(r.Context(), user.ID, in.ResumeID, in.ApprovalToken)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if errors.Is(err, career.ErrInvalidApprovalToken) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	mux.HandleFunc("/api/v1/career/tailor/resumes", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		resumes, err := careerService.ListTailoredResumes(r.Context(), user.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resumes)
	}))

	mux.HandleFunc("/api/v1/career/tailor/cover-letter", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			letterID := r.URL.Query().Get("id")
			if letterID == "" {
				http.Error(w, "Missing letter id parameter", http.StatusBadRequest)
				return
			}
			letter, err := careerService.GetCoverLetter(r.Context(), user.ID, letterID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(letter)
			return
		}

		if r.Method == http.MethodPost {
			var in struct {
				Job           career.JobTarget          `json:"job"`
				Format        career.MasterResumeFormat `json:"format"`
				RecipientName string                    `json:"recipient_name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request JSON payload", http.StatusBadRequest)
				return
			}
			if in.Format == "" {
				in.Format = career.ResumeFormatPDF
			}

			letter, err := careerService.CreateCoverLetter(r.Context(), user.ID, in.Job, in.Format, in.RecipientName)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(letter)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	mux.HandleFunc("/api/v1/career/tailor/cover-letter/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			LetterID      string `json:"letter_id"`
			ApprovalToken string `json:"approval_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		approved, err := careerService.ApproveCoverLetter(r.Context(), user.ID, in.LetterID, in.ApprovalToken)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if errors.Is(err, career.ErrInvalidApprovalToken) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	mux.HandleFunc("/api/v1/career/tailor/cover-letters", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		letters, err := careerService.ListCoverLetters(r.Context(), user.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(letters)
	}))

	mux.HandleFunc("/api/v1/career/tailor/download", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		docType := r.URL.Query().Get("type")
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		if docType == "cover_letter" {
			letter, err := careerService.GetCoverLetter(r.Context(), user.ID, id)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", letter.MimeType)
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, letter.FileName))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(letter.ByteContent)))
			_, _ = w.Write(letter.ByteContent)
			return
		}

		// Default: tailored resume
		tailored, err := careerService.GetTailoredResume(r.Context(), user.ID, id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		w.Header().Set("Content-Type", tailored.MimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, tailored.FileName))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(tailored.ByteContent)))
		_, _ = w.Write(tailored.ByteContent)
	}))

	// Resume Feedback & Quality Analysis (IMP-CAR-06)
	mux.HandleFunc("/api/v1/career/feedback/analyze", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		report, err := careerService.GenerateResumeFeedback(r.Context(), user.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	mux.HandleFunc("/api/v1/career/feedback/latest", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		report, err := careerService.GetLatestResumeFeedback(r.Context(), user.ID)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// Skills Demand & Market Intelligence Analysis (IMP-CAR-06, CAR-06, AT-028)
	mux.HandleFunc("/api/v1/career/skills-demand", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		role := r.URL.Query().Get("role")
		if role == "" {
			role = "backend"
		}

		analysis, err := careerService.GetSkillsDemandAnalysis(r.Context(), user.ID, role)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(analysis)
	}))

	// Multi-Board Job Discovery & Capability Reporting (IMP-CAR-07, REQ-005, AT-004, AT-010)
	mux.HandleFunc("/api/v1/career/discovery/search", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
		keywords := q.Get("keywords")
		location := q.Get("location")
		remoteOnly := q.Get("remote_only") == "true"

		var jobTypes []string
		if jt := q.Get("job_types"); jt != "" {
			for _, t := range strings.Split(jt, ",") {
				if trimmed := strings.TrimSpace(t); trimmed != "" {
					jobTypes = append(jobTypes, trimmed)
				}
			}
		}

		var sources []career.BoardSource
		if src := q.Get("sources"); src != "" {
			for _, s := range strings.Split(src, ",") {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					sources = append(sources, career.BoardSource(trimmed))
				}
			}
		}

		searchQuery := career.JobSearchQuery{
			Keywords:   keywords,
			Location:   location,
			RemoteOnly: remoteOnly,
			JobTypes:   jobTypes,
			Sources:    sources,
		}

		jobs, err := careerService.DiscoverJobs(r.Context(), searchQuery)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jobs)
	}))

	mux.HandleFunc("/api/v1/career/discovery/boards", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		caps := careerService.GetJobBoardCapabilities(r.Context())
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(caps)
	}))

	mux.HandleFunc("/api/v1/career/discovery/normalize-url", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var in struct {
			URL string `json:"url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		canonical := careerService.NormalizeJobURL(in.URL)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"original_url":  in.URL,
			"canonical_url": canonical,
		})
	}))

	// Fine-Grained Discovery Post-Filtering & Audit Reporting (IMP-CAR-08, CAR-08, AT-004)
	mux.HandleFunc("/api/v1/career/discovery/filter", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var in struct {
			Jobs   []career.DiscoveredJob   `json:"jobs"`
			Filter career.AdvancedJobFilter `json:"filter"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		filtered, report := careerService.FilterDiscoveredJobs(r.Context(), in.Jobs, in.Filter)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"filtered_jobs": filtered,
			"audit_report":  report,
		})
	}))

	// Explainable Job Matching & Candidate Fit Evaluation (IMP-CAR-09, CAR-09, AT-003, AT-028)
	mux.HandleFunc("/api/v1/career/match/evaluate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Job     career.DiscoveredJob   `json:"job"`
			Weights *career.ScoringWeights `json:"weights,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		result, err := careerService.EvaluateJobMatch(r.Context(), user.ID, in.Job, in.Weights)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))

	mux.HandleFunc("/api/v1/career/match/batch", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Jobs    []career.DiscoveredJob `json:"jobs"`
			Weights *career.ScoringWeights `json:"weights,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		results, err := careerService.EvaluateBatchJobMatches(r.Context(), user.ID, in.Jobs, in.Weights)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"matches": results,
			"total":   len(results),
		})
	}))

	mux.HandleFunc("/api/v1/career/match/weights", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		weights := careerService.GetDefaultScoringWeights()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"default_weights": weights,
			"description":     "Sum of weights must strictly equal 100%. Skills, TitleExperience, LocationWorkMode, Compensation.",
		})
	}))

	// Saved Jobs, Shortlists & Lifecycle (IMP-CAR-10, CAR-10)
	mux.HandleFunc("/api/v1/career/saved-jobs", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var in career.SaveJobInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			saved, err := careerService.SaveJob(r.Context(), user.ID, in)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				if errors.Is(err, career.ErrDuplicateSavedJob) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(saved)

		case http.MethodGet:
			id := r.URL.Query().Get("id")
			if id != "" {
				saved, err := careerService.GetSavedJob(r.Context(), user.ID, id)
				if err != nil {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusNotFound)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(saved)
				return
			}

			status := career.SavedJobStatus(r.URL.Query().Get("status"))
			tag := r.URL.Query().Get("tag")
			list, err := careerService.ListSavedJobs(r.Context(), user.ID, status, tag)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"saved_jobs": list,
				"total":      len(list),
			})

		case http.MethodPatch:
			id := r.URL.Query().Get("id")
			var in struct {
				ID string `json:"id,omitempty"`
				career.UpdateSavedJobInput
			}
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			if id == "" {
				id = in.ID
			}
			if id == "" {
				http.Error(w, "Missing saved job id", http.StatusBadRequest)
				return
			}

			updated, err := careerService.UpdateSavedJob(r.Context(), user.ID, id, in.UpdateSavedJobInput)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updated)

		case http.MethodDelete:
			id := r.URL.Query().Get("id")
			if id == "" {
				var in struct {
					ID string `json:"id"`
				}
				_ = json.NewDecoder(r.Body).Decode(&in)
				id = in.ID
			}
			if id == "" {
				http.Error(w, "Missing saved job id", http.StatusBadRequest)
				return
			}

			err := careerService.DeleteSavedJob(r.Context(), user.ID, id)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"deleted": true,
				"id":      id,
			})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Persistent Exclusion Blocklist ("Never see again" - CAR-10, C8)
	mux.HandleFunc("/api/v1/career/exclusions", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var in career.AddExclusionInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			exclusion, err := careerService.AddExclusion(r.Context(), user.ID, in)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(exclusion)

		case http.MethodGet:
			list, err := careerService.ListExclusions(r.Context(), user.ID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"exclusions": list,
				"total":      len(list),
			})

		case http.MethodDelete:
			id := r.URL.Query().Get("id")
			if id == "" {
				var in struct {
					ID string `json:"id"`
				}
				_ = json.NewDecoder(r.Body).Decode(&in)
				id = in.ID
			}
			if id == "" {
				http.Error(w, "Missing exclusion id", http.StatusBadRequest)
				return
			}

			err := careerService.DeleteExclusion(r.Context(), user.ID, id)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"deleted": true,
				"id":      id,
			})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Separate Historical Application Submission Ledger (CAR-10, AT-004, AT-006)
	mux.HandleFunc("/api/v1/career/applications", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodPost:
			var in career.RecordApplicationInput
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}
			record, err := careerService.RecordApplication(r.Context(), user.ID, in)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				if errors.Is(err, career.ErrDuplicateApplication) {
					w.WriteHeader(http.StatusConflict)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(record)

		case http.MethodGet:
			apps, err := careerService.ListApplications(r.Context(), user.ID)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"applications": apps,
				"total":        len(apps),
			})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Multi-Ledger Job Deduplication & Ambiguous Match Evaluation (CAR-10, AT-004)
	mux.HandleFunc("/api/v1/career/dedupe/check", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var in struct {
			Jobs []career.DiscoveredJob `json:"jobs"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		results, err := careerService.CheckJobDeduplication(r.Context(), user.ID, in.Jobs)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"results": results,
			"total":   len(results),
		})
	}))

	// --- Application Workflow & Human Review (IMP-CAR-11, CAR-11, REQ-005, REQ-015, AT-003, AT-005, AT-007) ---

	// POST /api/v1/career/apply/prepare
	mux.HandleFunc("/api/v1/career/apply/prepare", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			Job           career.DiscoveredJob            `json:"job"`
			ResumeID      string                          `json:"resume_id,omitempty"`
			CoverLetterID string                          `json:"cover_letter_id,omitempty"`
			ExecutionMode career.ApplicationExecutionMode `json:"execution_mode,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, err := careerService.PrepareApplicationWorkflow(
			r.Context(),
			user.ID,
			fmt.Sprintf("ws_%s", user.ID),
			req.Job,
			req.ResumeID,
			req.CoverLetterID,
			req.ExecutionMode,
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session": session,
		})
	}))

	// GET /api/v1/career/apply/sessions
	mux.HandleFunc("/api/v1/career/apply/sessions", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		status := career.ApplicationWorkflowStatus(r.URL.Query().Get("status"))
		sessions, err := careerService.ListApplicationWorkflowSessions(r.Context(), user.ID, status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions": sessions,
			"total":    len(sessions),
		})
	}))

	// GET /api/v1/career/apply/session
	mux.HandleFunc("/api/v1/career/apply/session", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		sessionID := r.URL.Query().Get("id")
		if sessionID == "" {
			http.Error(w, "missing id parameter", http.StatusBadRequest)
			return
		}

		session, err := careerService.GetApplicationWorkflowSession(r.Context(), user.ID, sessionID)
		if err != nil {
			if errors.Is(err, career.ErrWorkflowNotFound) {
				http.Error(w, "session not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Also check timeout if status is dispatched (AT-005)
		if session.Status == career.WorkflowStatusDispatched {
			session, _, _ = careerService.CheckApplicationWorkflowTimeout(r.Context(), user.ID, sessionID)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session": session,
		})
	}))

	// PATCH /api/v1/career/apply/session/answers
	mux.HandleFunc("/api/v1/career/apply/session/answers", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			SessionID string                             `json:"session_id"`
			Answers   []career.ApplicationQuestionAnswer `json:"answers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, approvalInvalidated, err := careerService.UpdateApplicationWorkflowAnswers(r.Context(), user.ID, req.SessionID, req.Answers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session":              session,
			"approval_invalidated": approvalInvalidated,
		})
	}))

	// POST /api/v1/career/apply/session/approve
	mux.HandleFunc("/api/v1/career/apply/session/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, err := careerService.ApproveApplicationWorkflow(r.Context(), user.ID, req.SessionID, user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session":        session,
			"approval_token": session.ApprovalToken,
		})
	}))

	// POST /api/v1/career/apply/session/dispatch
	mux.HandleFunc("/api/v1/career/apply/session/dispatch", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			SessionID     string                          `json:"session_id"`
			ExecutionMode career.ApplicationExecutionMode `json:"execution_mode,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, err := careerService.DispatchApplicationWorkflow(r.Context(), user.ID, req.SessionID, req.ExecutionMode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session": session,
		})
	}))

	// POST /api/v1/career/apply/session/confirm
	mux.HandleFunc("/api/v1/career/apply/session/confirm", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			SessionID string                   `json:"session_id"`
			Receipt   career.SubmissionReceipt `json:"receipt"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, appRecord, err := careerService.ConfirmApplicationWorkflowSubmission(r.Context(), user.ID, req.SessionID, req.Receipt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session":            session,
			"application_record": appRecord,
		})
	}))

	// POST /api/v1/career/apply/session/cancel
	mux.HandleFunc("/api/v1/career/apply/session/cancel", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			SessionID string `json:"session_id"`
			Reason    string `json:"reason,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		session, err := careerService.CancelApplicationWorkflow(r.Context(), user.ID, req.SessionID, req.Reason)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"session": session,
		})
	}))

	// POST /api/v1/career/fields/recognize
	mux.HandleFunc("/api/v1/career/fields/recognize", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req career.RecognizeFieldsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.RecognizeFormFields(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/fields/resolve
	mux.HandleFunc("/api/v1/career/fields/resolve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.ResolveAnswersRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.ResolveFormAnswers(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/fields/evaluate-conditions
	mux.HandleFunc("/api/v1/career/fields/evaluate-conditions", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req career.EvaluateConditionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.EvaluateConditionalFields(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// Multi-step Wizard, Shadow DOM & Loop Detection Endpoints (IMP-CAR-13, CAR-13, AT-003, AT-006, FND-009, SRC-C4)
	// POST /api/v1/career/wizard/init
	mux.HandleFunc("/api/v1/career/wizard/init", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req career.InitWizardRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.InitFormStateMachine(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/wizard/advance
	mux.HandleFunc("/api/v1/career/wizard/advance", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req career.AdvanceWizardStepRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.AdvanceWizardStep(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/wizard/previous
	mux.HandleFunc("/api/v1/career/wizard/previous", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req career.PreviousWizardStepRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.PreviousWizardStep(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/wizard/shadow-dom
	mux.HandleFunc("/api/v1/career/wizard/shadow-dom", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req career.TraverseShadowDOMRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.TraverseShadowDOM(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// External and Indeed applications Endpoints (IMP-CAR-14, CAR-14, AT-005, AT-006, AT-010, SRC-C3)
	// POST /api/v1/career/portal/classify
	mux.HandleFunc("/api/v1/career/portal/classify", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req career.ClassifyPortalRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.ClassifyPortal(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/portal/prepare-external
	mux.HandleFunc("/api/v1/career/portal/prepare-external", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.PrepareExternalBundleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.PrepareExternalBundle(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/portal/dispatch-external
	mux.HandleFunc("/api/v1/career/portal/dispatch-external", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.DispatchExternalApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.DispatchExternalApplication(r.Context(), req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/portal/confirm-external
	mux.HandleFunc("/api/v1/career/portal/confirm-external", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.ConfirmExternalApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.ConfirmExternalApplication(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// Application Status, Verified Applied Mark & Reconciliation Endpoints (IMP-CAR-15, CAR-15, REQ-005, REQ-016, AT-005, AT-006)
	// POST /api/v1/career/applications/verify-applied
	mux.HandleFunc("/api/v1/career/applications/verify-applied", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.VerifyAppliedMarkRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.VerifyAppliedMark(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/applications/reconcile
	mux.HandleFunc("/api/v1/career/applications/reconcile", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.ReconcileApplicationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.ReconcileApplication(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/applications/stage
	mux.HandleFunc("/api/v1/career/applications/stage", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.UpdateApplicationStageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.UpdateApplicationStage(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// GET /api/v1/career/applications/audit-ledger
	mux.HandleFunc("/api/v1/career/applications/audit-ledger", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		recordID := r.URL.Query().Get("record_id")
		if recordID == "" {
			http.Error(w, "missing record_id parameter", http.StatusBadRequest)
			return
		}

		res, err := careerService.GetApplicationAuditLedger(r.Context(), user.ID, recordID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/applications/expire-check
	mux.HandleFunc("/api/v1/career/applications/expire-check", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		transitioned, err := careerService.CheckAndExpireDispatchedSessions(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"transitioned_count": len(transitioned),
			"transitioned":       transitioned,
		})
	}))

	// Per-Application Evidence / Document Bundle Endpoints (IMP-CAR-16, CAR-16, AT-005, AT-011, SRC-C2, SRC-C3, SRC-C4)
	// POST /api/v1/career/applications/evidence
	mux.HandleFunc("/api/v1/career/applications/evidence", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.CreateEvidenceBundleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.CreateEvidenceBundle(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(res)
	}))

	// GET /api/v1/career/applications/evidence/get
	mux.HandleFunc("/api/v1/career/applications/evidence/get", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		bundleID := r.URL.Query().Get("bundle_id")
		if bundleID == "" {
			http.Error(w, "missing bundle_id query parameter", http.StatusBadRequest)
			return
		}

		isAdmin := r.URL.Query().Get("is_admin") == "true"
		hasGrant := r.URL.Query().Get("has_grant") == "true"

		req := career.GetEvidenceBundleRequest{
			BundleID:         bundleID,
			RequestingUserID: user.ID,
			IsWorkspaceAdmin: isAdmin,
			HasExplicitGrant: hasGrant,
		}

		res, err := careerService.GetEvidenceBundle(r.Context(), req)
		if err != nil {
			if errors.Is(err, career.ErrWorkspaceAdminAccessDenied) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			if errors.Is(err, career.ErrEvidenceBundleNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/applications/evidence/verify
	mux.HandleFunc("/api/v1/career/applications/evidence/verify", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.VerifyBundleIntegrityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		res, err := careerService.VerifyEvidenceBundleIntegrity(r.Context(), req.BundleID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// GET /api/v1/career/applications/evidence/by-app
	mux.HandleFunc("/api/v1/career/applications/evidence/by-app", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		appID := r.URL.Query().Get("app_id")
		if appID == "" {
			http.Error(w, "missing app_id query parameter", http.StatusBadRequest)
			return
		}

		isAdmin := r.URL.Query().Get("is_admin") == "true"
		hasGrant := r.URL.Query().Get("has_grant") == "true"

		res, err := careerService.GetEvidenceBundleByApplicationID(r.Context(), appID, user.ID, isAdmin, hasGrant)
		if err != nil {
			if errors.Is(err, career.ErrWorkspaceAdminAccessDenied) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			if errors.Is(err, career.ErrEvidenceBundleNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// Google Sheets Export & Sync Endpoints (IMP-CAR-17, CAR-17, REQ-006, AT-013, AT-014, SRC-C4)
	// POST /api/v1/career/export/sheets/config
	mux.HandleFunc("/api/v1/career/export/sheets/config", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			cfg, err := careerService.GetGoogleSheetsSyncConfig(r.Context(), user.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cfg)
			return
		}

		var cfg career.SheetsSyncConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		saved, err := careerService.ConfigureGoogleSheetsSync(r.Context(), user.ID, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(saved)
	}))

	// GET /api/v1/career/export/sheets/preview
	mux.HandleFunc("/api/v1/career/export/sheets/preview", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		preview, err := careerService.GetGoogleSheetsPreview(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(preview)
	}))

	// POST /api/v1/career/export/sheets/sync
	mux.HandleFunc("/api/v1/career/export/sheets/sync", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ExistingSheetRows [][]string `json:"existing_sheet_rows"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Body can be empty for clean initial sync
			req.ExistingSheetRows = nil
		}

		result, reconciledRows, err := careerService.SyncApplicationsToGoogleSheets(r.Context(), user.ID, req.ExistingSheetRows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result":          result,
			"reconciled_rows": reconciledRows,
		})
	}))

	// GET /api/v1/career/export/sheets/csv
	mux.HandleFunc("/api/v1/career/export/sheets/csv", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		csvData, err := careerService.ExportApplicationsToCSV(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", "attachment; filename=\"career_applications.csv\"")
		_, _ = w.Write([]byte(csvData))
	}))

	// GET /api/v1/career/export/csv (IMP-CAR-18, CAR-18, REQ-006, AT-013, FND-006)
	mux.HandleFunc("/api/v1/career/export/csv", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		q := r.URL.Query()
		datasetType := career.ExportDatasetType(q.Get("dataset"))
		if datasetType == "" {
			datasetType = career.ExportDatasetApplications
		}

		withBOM := q.Get("with_bom") == "true"
		includeHeaders := q.Get("include_headers") != "false"
		stageFilter := q.Get("stage")
		platformFilter := q.Get("platform")

		var customCols []string
		if colsParam := q.Get("columns"); colsParam != "" {
			for _, col := range strings.Split(colsParam, ",") {
				c := strings.TrimSpace(col)
				if c != "" {
					customCols = append(customCols, c)
				}
			}
		}

		var fromDate, toDate *time.Time
		if fd := q.Get("from_date"); fd != "" {
			if t, err := time.Parse(time.RFC3339, fd); err == nil {
				fromDate = &t
			}
		}
		if td := q.Get("to_date"); td != "" {
			if t, err := time.Parse(time.RFC3339, td); err == nil {
				toDate = &t
			}
		}

		opts := career.ExportFilterOptions{
			DatasetType:     datasetType,
			IncludeHeaders:  includeHeaders,
			WithBOM:         withBOM,
			StageFilter:     stageFilter,
			PlatformFilter:  platformFilter,
			StartDate:       fromDate,
			EndDate:         toDate,
			SelectedColumns: customCols,
		}

		manifest, err := careerService.ExportDatasetCSV(r.Context(), user.ID, datasetType, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Return JSON preview if requested
		if q.Get("as_json") == "true" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(manifest)
			return
		}

		w.Header().Set("Content-Type", "text/csv; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", manifest.Filename))
		_, _ = w.Write([]byte(manifest.CSVContent))
	}))

	// POST /api/v1/career/export/custom (IMP-CAR-18, CAR-18, AT-013)
	mux.HandleFunc("/api/v1/career/export/custom", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var opts career.ExportFilterOptions
		if err := json.NewDecoder(r.Body).Decode(&opts); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if opts.DatasetType == "" {
			opts.DatasetType = career.ExportDatasetApplications
		}

		manifest, err := careerService.ExportDatasetCSV(r.Context(), user.ID, opts.DatasetType, opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(manifest)
	}))

	// GET /api/v1/career/export/history (IMP-CAR-18, EXP-003, AT-022)
	mux.HandleFunc("/api/v1/career/export/history", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		history, err := careerService.GetExportAuditHistory(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(history)
	}))

	// GET & POST /api/v1/career/contacts (IMP-CAR-18, CAR-18)
	mux.HandleFunc("/api/v1/career/contacts", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			contacts, err := careerService.ListRecruiterContacts(r.Context(), user.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(contacts)

		case http.MethodPost:
			var contact career.RecruiterContact
			if err := json.NewDecoder(r.Body).Decode(&contact); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			contact.UserID = user.ID
			if err := careerService.SaveRecruiterContact(r.Context(), &contact); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(contact)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// --- Daily Reports & Reminders Endpoints (IMP-CAR-19, CAR-19, AT-018) ---

	// GET /api/v1/career/reports/daily
	mux.HandleFunc("/api/v1/career/reports/daily", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = user.ID
		}

		date := r.URL.Query().Get("date")
		report, err := careerService.GetDailyReport(r.Context(), user.ID, workspaceID, date)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// GET & PUT /api/v1/career/reports/config
	mux.HandleFunc("/api/v1/career/reports/config", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = user.ID
		}

		switch r.Method {
		case http.MethodGet:
			cfg, err := careerService.GetDailyReportConfig(r.Context(), user.ID, workspaceID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(cfg)

		case http.MethodPut:
			var req career.DailyReportConfig
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			updated, err := careerService.UpdateDailyReportConfig(r.Context(), user.ID, workspaceID, &req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updated)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// GET & POST /api/v1/career/reminders
	mux.HandleFunc("/api/v1/career/reminders", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = user.ID
		}

		switch r.Method {
		case http.MethodGet:
			statusFilter := career.ReminderStatus(r.URL.Query().Get("status"))
			reminders, err := careerService.ListReminders(r.Context(), user.ID, statusFilter)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(reminders)

		case http.MethodPost:
			var req career.CareerReminder
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			created, err := careerService.CreateCustomReminder(r.Context(), user.ID, workspaceID, &req)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// POST /api/v1/career/reminders/action
	mux.HandleFunc("/api/v1/career/reminders/action", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ReminderID  string                `json:"reminder_id"`
			Action      career.ReminderStatus `json:"action"` // "snoozed", "dismissed", "completed"
			SnoozeHours int                   `json:"snooze_hours,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := careerService.UpdateReminderStatus(r.Context(), user.ID, req.ReminderID, req.Action, req.SnoozeHours); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "success",
			"reminder_id": req.ReminderID,
			"action":      req.Action,
		})
	}))

	// --- Hiring Posts & Recruiter Leads Routes (IMP-CAR-20, CAR-20, AT-019) ---

	// POST /api/v1/career/leads/ingest-post
	mux.HandleFunc("/api/v1/career/leads/ingest-post", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req career.IngestHiringPostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		post, lead, err := careerService.IngestHiringPost(r.Context(), user.ID, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"post": post,
			"lead": lead,
		})
	}))

	// GET /api/v1/career/leads/posts
	mux.HandleFunc("/api/v1/career/leads/posts", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		posts, err := careerService.ListHiringPosts(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"posts": posts,
			"count": len(posts),
		})
	}))

	// GET /api/v1/career/leads
	mux.HandleFunc("/api/v1/career/leads", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		filterStatus := career.LeadReviewStatus(r.URL.Query().Get("status"))
		leads, err := careerService.ListRecruiterLeads(r.Context(), user.ID, filterStatus)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"leads": leads,
			"count": len(leads),
		})
	}))

	// POST /api/v1/career/leads/review (CAR-20, REQ-015)
	mux.HandleFunc("/api/v1/career/leads/review", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			LeadID string                  `json:"lead_id"`
			Action career.LeadReviewStatus `json:"action"` // "approved", "rejected"
			Notes  string                  `json:"notes,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		lead, err := careerService.ReviewRecruiterLead(r.Context(), user.ID, req.LeadID, req.Action, user.ID, req.Notes)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lead)
	}))

	// POST /api/v1/career/leads/draft
	mux.HandleFunc("/api/v1/career/leads/draft", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			LeadID       string                      `json:"lead_id"`
			TemplateType career.OutreachTemplateType `json:"template_type"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		draft, err := careerService.GenerateLeadOutreach(r.Context(), user.ID, req.LeadID, req.TemplateType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(draft)
	}))

	// POST /api/v1/career/leads/action
	mux.HandleFunc("/api/v1/career/leads/action", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			LeadID string                    `json:"lead_id"`
			Action career.LeadOutreachStatus `json:"action"` // "copied_to_clipboard", "contacted", etc.
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		lead, err := careerService.RecordLeadOutreachAction(r.Context(), user.ID, req.LeadID, req.Action)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lead)
	}))

	// POST /api/v1/career/leads/link-app
	mux.HandleFunc("/api/v1/career/leads/link-app", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			LeadID        string `json:"lead_id"`
			ApplicationID string `json:"application_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		lead, err := careerService.LinkLeadApplication(r.Context(), user.ID, req.LeadID, req.ApplicationID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lead)
	}))

	// --- Run Controls & Crash Recovery Endpoints (IMP-CAR-21, CAR-21, AT-006, AT-021, REQ-017, FND-011) ---

	// POST /api/v1/career/runs/create
	mux.HandleFunc("/api/v1/career/runs/create", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			WorkspaceID string               `json:"workspace_id"`
			RunType     career.CareerRunType `json:"run_type"`
			TotalItems  int                  `json:"total_items"`
			HourlyLimit int                  `json:"hourly_limit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		run, err := careerService.CreateCareerRun(r.Context(), user.ID, req.WorkspaceID, req.RunType, req.TotalItems, req.HourlyLimit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(run)
	}))

	// GET /api/v1/career/runs
	mux.HandleFunc("/api/v1/career/runs", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		runs, err := careerService.ListCareerRuns(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"runs":  runs,
			"count": len(runs),
		})
	}))

	// Router for /api/v1/career/runs/ detail & sub-actions
	mux.HandleFunc("/api/v1/career/runs/", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/v1/career/runs/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "missing run id", http.StatusBadRequest)
			return
		}
		runID := parts[0]

		if len(parts) == 1 {
			// GET /api/v1/career/runs/{id}
			if r.Method != http.MethodGet {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			run, err := careerService.GetCareerRun(r.Context(), user.ID, runID)
			if err != nil {
				if errors.Is(err, career.ErrRunNotFound) {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(run)
			return
		}

		subAction := parts[1]
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		switch subAction {
		case "control":
			var req struct {
				Action career.RunControlAction `json:"action"` // pause, resume, cancel, drain
				Reason string                  `json:"reason,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			run, err := careerService.ControlCareerRun(r.Context(), user.ID, runID, req.Action, req.Reason)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(run)

		case "commit":
			var req struct {
				Token       int64                   `json:"token"`
				Stage       string                  `json:"stage"`
				Safety      career.StageSafetyLevel `json:"safety"`
				Succeeded   bool                    `json:"succeeded"`
				Diagnostics string                  `json:"diagnostics,omitempty"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			run, err := careerService.CommitRunAttempt(r.Context(), user.ID, runID, req.Token, req.Stage, req.Safety, req.Succeeded, req.Diagnostics)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(run)

		case "recover":
			run, recovered, err := careerService.RecoverCareerRun(r.Context(), user.ID, runID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"run":       run,
				"recovered": recovered,
			})

		case "reconcile":
			var req struct {
				Resolution    string `json:"resolution"` // confirm_completed, abandon_attempt, force_retry
				MarkSucceeded bool   `json:"mark_succeeded"`
				Notes         string `json:"notes"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}
			run, err := careerService.ReconcileCareerRun(r.Context(), user.ID, runID, req.Resolution, req.MarkSucceeded, req.Notes)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(run)

		default:
			http.Error(w, "not found", http.StatusNotFound)
		}
	}))

	// --- PII Controls & LLM Provider Choices Endpoints (IMP-CAR-22, CAR-22, REQ-023, AT-019) ---

	// POST /api/v1/career/pii/redact
	mux.HandleFunc("/api/v1/career/pii/redact", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			Text   string                     `json:"text"`
			Policy *career.PIIRedactionPolicy `json:"policy,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		policy := career.DefaultPIIRedactionPolicy()
		if req.Policy != nil {
			policy = *req.Policy
		}

		res, err := careerService.RedactCandidatePII(r.Context(), user.ID, req.Text, policy)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// POST /api/v1/career/pii/rehydrate
	mux.HandleFunc("/api/v1/career/pii/rehydrate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			AnonymizedText string            `json:"anonymized_text"`
			TokenMap       map[string]string `json:"token_map"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		rehydrated := careerService.RehydrateCandidateText(req.AnonymizedText, req.TokenMap)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"rehydrated_text": rehydrated,
		})
	}))

	// GET /api/v1/career/pii/consents
	mux.HandleFunc("/api/v1/career/pii/consents", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		consents, err := careerService.ListProviderConsents(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"consents": consents,
			"count":    len(consents),
		})
	}))

	// POST /api/v1/career/pii/consents/grant
	mux.HandleFunc("/api/v1/career/pii/consents/grant", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ProviderID           string   `json:"provider_id"`
			ProviderName         string   `json:"provider_name"`
			AllowedTasks         []string `json:"allowed_tasks"`
			ZeroTrainingAffirmed bool     `json:"zero_training_affirmed"`
			RetentionDays        int      `json:"retention_days"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		consent, err := careerService.GrantProviderConsent(r.Context(), user.ID, req.ProviderID, req.ProviderName, req.AllowedTasks, req.ZeroTrainingAffirmed, req.RetentionDays)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(consent)
	}))

	// POST /api/v1/career/pii/consents/revoke
	mux.HandleFunc("/api/v1/career/pii/consents/revoke", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ProviderID string `json:"provider_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		consent, err := careerService.RevokeProviderConsent(r.Context(), user.ID, req.ProviderID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(consent)
	}))

	// =========================================================================
	// 32. Interview Preparation & Outcomes API (IMP-CAR-23, CAR-23, AT-003)
	// =========================================================================

	// POST & GET /api/v1/career/interviews/prep
	mux.HandleFunc("/api/v1/career/interviews/prep", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				ApplicationRecordID string              `json:"application_record_id"`
				JobTitle            string              `json:"job_title"`
				RequiredSkills      []string            `json:"required_skills"`
				Company             career.CompanyBrief `json:"company"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			pack, err := careerService.GenerateInterviewPrep(r.Context(), req.ApplicationRecordID, user.ID, req.Company, req.JobTitle, req.RequiredSkills)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(pack)
			return
		}

		if r.Method == http.MethodGet {
			packID := r.URL.Query().Get("pack_id")
			appID := r.URL.Query().Get("app_id")

			if packID != "" {
				pack, err := careerService.GetInterviewPrep(r.Context(), packID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(pack)
				return
			}

			if appID != "" {
				pack, err := careerService.GetInterviewPrepByApp(r.Context(), appID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(pack)
				return
			}

			http.Error(w, "app_id or pack_id query param required", http.StatusBadRequest)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// POST & GET /api/v1/career/interviews/schedule
	mux.HandleFunc("/api/v1/career/interviews/schedule", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				ApplicationRecordID string    `json:"application_record_id"`
				RoundID             string    `json:"round_id"`
				StageName           string    `json:"stage_name"`
				ScheduledAt         time.Time `json:"scheduled_at"`
				Timezone            string    `json:"timezone"`
				Format              string    `json:"format"`
				MeetingLink         string    `json:"meeting_link"`
				InterviewerName     string    `json:"interviewer_name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			sched, err := careerService.ScheduleInterviewRound(r.Context(), req.ApplicationRecordID, req.RoundID, req.StageName, req.ScheduledAt, req.Timezone, req.Format, req.MeetingLink, req.InterviewerName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(sched)
			return
		}

		if r.Method == http.MethodGet {
			appID := r.URL.Query().Get("app_id")
			schedules, err := careerService.ListInterviewSchedules(r.Context(), appID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"schedules": schedules,
				"count":     len(schedules),
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// POST & GET /api/v1/career/interviews/notes
	mux.HandleFunc("/api/v1/career/interviews/notes", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				ApplicationRecordID       string   `json:"application_record_id"`
				RoundID                   string   `json:"round_id"`
				PreInterviewNotes         string   `json:"pre_interview_notes"`
				QuestionsAskedByCandidate []string `json:"questions_asked_by_candidate"`
				PostInterviewReflections  string   `json:"post_interview_reflections"`
				InterviewerName           string   `json:"interviewer_name"`
				InterviewerTitle          string   `json:"interviewer_title"`
				SelfRating                int      `json:"self_rating"`
				FollowUpActions           string   `json:"follow_up_actions"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid request body", http.StatusBadRequest)
				return
			}

			note, err := careerService.RecordInterviewSessionNote(r.Context(), req.ApplicationRecordID, req.RoundID, user.ID, req.PreInterviewNotes, req.QuestionsAskedByCandidate, req.PostInterviewReflections, req.InterviewerName, req.InterviewerTitle, req.SelfRating, req.FollowUpActions)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(note)
			return
		}

		if r.Method == http.MethodGet {
			roundID := r.URL.Query().Get("round_id")
			appID := r.URL.Query().Get("app_id")

			if roundID != "" {
				note, err := careerService.GetInterviewSessionNote(r.Context(), roundID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(note)
				return
			}

			notes, err := careerService.ListInterviewSessionNotes(r.Context(), appID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"notes": notes,
				"count": len(notes),
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// POST /api/v1/career/interviews/outcome
	mux.HandleFunc("/api/v1/career/interviews/outcome", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ScheduleID string              `json:"schedule_id"`
			Outcome    career.RoundOutcome `json:"outcome"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := careerService.UpdateRoundOutcome(r.Context(), req.ScheduleID, req.Outcome); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "updated",
			"schedule_id": req.ScheduleID,
			"outcome":     req.Outcome,
		})
	}))

	// GET /api/v1/career/interviews/funnel
	mux.HandleFunc("/api/v1/career/interviews/funnel", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		funnel, err := careerService.GetOutcomeFunnel(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(funnel)
	}))

	// POST /api/v1/career/consistency/audit
	mux.HandleFunc("/api/v1/career/consistency/audit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		report, err := careerService.RunConsistencyAudit(r.Context(), user.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// GET /api/v1/career/consistency/latest
	mux.HandleFunc("/api/v1/career/consistency/latest", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		report, err := careerService.GetLatestConsistencyReport(r.Context(), user.ID)
		if err != nil {
			if errors.Is(err, career.ErrReportNotFound) {
				http.Error(w, "No consistency report found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// POST /api/v1/career/consistency/resolve
	mux.HandleFunc("/api/v1/career/consistency/resolve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ReportID   string                    `json:"report_id"`
			MismatchID string                    `json:"mismatch_id"`
			Decision   career.ResolutionDecision `json:"decision"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.ReportID == "" || req.MismatchID == "" {
			http.Error(w, "report_id and mismatch_id are required", http.StatusBadRequest)
			return
		}

		report, err := careerService.ResolveConsistencyMismatch(r.Context(), user.ID, req.ReportID, req.MismatchID, req.Decision)
		if err != nil {
			status := http.StatusBadRequest
			if errors.Is(err, career.ErrReportNotFound) {
				status = http.StatusNotFound
			}
			http.Error(w, err.Error(), status)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// POST /api/v1/career/consistency/linkedin-snapshot
	mux.HandleFunc("/api/v1/career/consistency/linkedin-snapshot", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var snap career.LinkedInProfileSnapshot
		if err := json.NewDecoder(r.Body).Decode(&snap); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		snap.UserID = user.ID
		if snap.SnapshotID == "" {
			snap.SnapshotID = fmt.Sprintf("snap-%d", time.Now().UnixNano())
		}
		if snap.ImportedAt.IsZero() {
			snap.ImportedAt = time.Now().UTC()
		}

		if err := careerService.SaveLinkedInSnapshot(r.Context(), &snap); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "saved",
			"snapshot_id": snap.SnapshotID,
		})
	}))

	// LinkedIn Capabilities and Integration Endpoints (IMP-LI-01, LI-01, REQ-004, REQ-021, AT-010, AT-016)
	// GET /api/v1/linkedin/capabilities
	mux.HandleFunc("/api/v1/linkedin/capabilities", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		conn, err := linkedinService.InspectCapabilities(r.Context(), user.ID)
		if err != nil {
			if errors.Is(err, linkedin.ErrConnectionNotFound) {
				http.Error(w, "LinkedIn connection not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(conn)
	}))

	// POST /api/v1/linkedin/connect
	mux.HandleFunc("/api/v1/linkedin/connect", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			MemberID       string   `json:"member_id"`
			DisplayName    string   `json:"display_name"`
			Email          string   `json:"email"`
			GrantedScopes  []string `json:"granted_scopes"`
			RawToken       string   `json:"raw_token"`
			ExpiresInHours int      `json:"expires_in_hours"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.MemberID == "" {
			req.MemberID = fmt.Sprintf("urn:li:person:%s", user.ID)
		}
		if req.DisplayName == "" {
			req.DisplayName = user.Email
		}
		if req.Email == "" {
			req.Email = user.Email
		}
		if len(req.GrantedScopes) == 0 {
			req.GrantedScopes = []string{"openid", "profile", "email"}
		}
		if req.ExpiresInHours == 0 {
			req.ExpiresInHours = 720
		}
		if req.RawToken == "" {
			req.RawToken = "tok_oauth_dev_default_sample_secret_key"
		}

		conn, err := linkedinService.ConnectAccount(r.Context(), user.ID, req.MemberID, req.DisplayName, req.Email, req.GrantedScopes, req.RawToken, req.ExpiresInHours)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(conn)
	}))

	// POST /api/v1/linkedin/disconnect
	mux.HandleFunc("/api/v1/linkedin/disconnect", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := linkedinService.DisconnectAccount(r.Context(), user.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "disconnected",
			"user_id": user.ID,
		})
	}))

	// POST /api/v1/linkedin/check-action
	mux.HandleFunc("/api/v1/linkedin/check-action", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var body struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Action == "" {
			http.Error(w, "action is required", http.StatusBadRequest)
			return
		}

		allowed, capObj, err := linkedinService.CheckActionPermission(r.Context(), user.ID, body.Action)
		response := map[string]interface{}{
			"action":     body.Action,
			"allowed":    allowed,
			"capability": capObj,
		}
		if err != nil {
			response["error"] = err.Error()
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))

	// POST /api/v1/linkedin/profile/import-zip
	mux.HandleFunc("/api/v1/linkedin/profile/import-zip", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var zipBytes []byte
		// Support both multipart/form-data and direct octet-stream body
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
				return
			}
			file, _, err := r.FormFile("archive")
			if err != nil {
				file, _, err = r.FormFile("file")
			}
			if err != nil {
				http.Error(w, "archive or file form field is required", http.StatusBadRequest)
				return
			}
			defer file.Close()
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, file); err != nil {
				http.Error(w, "Failed to read uploaded archive", http.StatusInternalServerError)
				return
			}
			zipBytes = buf.Bytes()
		} else {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r.Body); err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			zipBytes = buf.Bytes()
		}

		if len(zipBytes) == 0 {
			http.Error(w, "Archive file cannot be empty", http.StatusBadRequest)
			return
		}

		imported, err := linkedinService.ImportProfileArchive(r.Context(), user.ID, zipBytes)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse LinkedIn zip archive: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(imported)
	}))

	// POST /api/v1/linkedin/profile/import-text
	mux.HandleFunc("/api/v1/linkedin/profile/import-text", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var body struct {
			RawText string `json:"raw_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.RawText) == "" {
			http.Error(w, "raw_text is required", http.StatusBadRequest)
			return
		}

		imported, err := linkedinService.ImportPastedProfile(r.Context(), user.ID, body.RawText)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse pasted text: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(imported)
	}))

	// POST /api/v1/linkedin/profile/import-api
	mux.HandleFunc("/api/v1/linkedin/profile/import-api", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var bodyBytes bytes.Buffer
		if _, err := io.Copy(&bodyBytes, r.Body); err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		imported, err := linkedinService.ImportEnterpriseProfile(r.Context(), user.ID, bodyBytes.Bytes())
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to parse enterprise payload: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(imported)
	}))

	// POST /api/v1/linkedin/profile/import-oidc-fallback
	mux.HandleFunc("/api/v1/linkedin/profile/import-oidc-fallback", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var claims map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&claims); err != nil {
			http.Error(w, "Valid claims JSON object is required", http.StatusBadRequest)
			return
		}

		imported, err := linkedinService.ImportOIDCProfileFallback(r.Context(), user.ID, claims)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to process OIDC fallback: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(imported)
	}))

	// GET /api/v1/linkedin/profile/latest
	mux.HandleFunc("/api/v1/linkedin/profile/latest", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		imported, err := linkedinService.GetLatestProfileImport(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "No imported LinkedIn profile found for user", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(imported)
	}))

	// POST /api/v1/linkedin/profile/merge-to-career
	mux.HandleFunc("/api/v1/linkedin/profile/merge-to-career", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ImportID         string `json:"import_id"`
			MergeHeadline    bool   `json:"merge_headline"`
			MergeSummary     bool   `json:"merge_summary"`
			MergeExperiences bool   `json:"merge_experiences"`
			MergeEducation   bool   `json:"merge_education"`
			MergeSkills      bool   `json:"merge_skills"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var imported *linkedin.LinkedInImportedProfile
		var err error
		if req.ImportID != "" {
			imported, err = linkedinService.GetProfileImport(r.Context(), req.ImportID)
		} else {
			imported, err = linkedinService.GetLatestProfileImport(r.Context(), user.ID)
		}
		if err != nil {
			http.Error(w, "Imported profile not found", http.StatusNotFound)
			return
		}

		// Convert to career.LinkedInProfileSnapshot for CAR-24 consistency checks
		snapExperiences := make([]career.ExperienceItem, 0, len(imported.Experiences))
		for _, e := range imported.Experiences {
			snapExperiences = append(snapExperiences, career.ExperienceItem{
				ID:          fmt.Sprintf("li-exp-%d", time.Now().UnixNano()),
				Title:       e.Title,
				Company:     e.CompanyName,
				Location:    e.Location,
				IsCurrent:   e.IsCurrent,
				Description: e.Description,
				Confirmed:   true,
			})
		}
		snapEducation := make([]career.EducationItem, 0, len(imported.Education))
		for _, ed := range imported.Education {
			var highlights []string
			if strings.TrimSpace(ed.Notes) != "" {
				highlights = append(highlights, ed.Notes)
			}
			snapEducation = append(snapEducation, career.EducationItem{
				ID:           fmt.Sprintf("li-edu-%d", time.Now().UnixNano()),
				Institution:  ed.SchoolName,
				Degree:       ed.DegreeName,
				FieldOfStudy: ed.FieldOfStudy,
				Highlights:   highlights,
				Confirmed:    true,
			})
		}

		snapshot := &career.LinkedInProfileSnapshot{
			SnapshotID:  fmt.Sprintf("li-snap-%d", time.Now().UnixNano()),
			UserID:      user.ID,
			Headline:    imported.Headline,
			Summary:     imported.Summary,
			Experiences: snapExperiences,
			Education:   snapEducation,
			Skills:      imported.Skills,
			ImportedAt:  time.Now().UTC(),
			SourceMode:  string(imported.SourceMode),
		}

		if err := careerService.SaveLinkedInSnapshot(r.Context(), snapshot); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save LinkedIn snapshot: %v", err), http.StatusInternalServerError)
			return
		}

		// Merge into MasterCareerProfile if requested
		masterProfile, profErr := careerService.GetMasterProfile(r.Context(), user.ID)
		if profErr == nil && masterProfile != nil {
			if req.MergeHeadline && imported.Headline != "" && masterProfile.Contact.Headline == "" {
				masterProfile.Contact.Headline = imported.Headline
			}
			if req.MergeSummary && imported.Summary != "" && masterProfile.Contact.Summary == "" {
				masterProfile.Contact.Summary = imported.Summary
			}
			if req.MergeExperiences && len(snapExperiences) > 0 {
				masterProfile.Experiences = append(masterProfile.Experiences, snapExperiences...)
			}
			if req.MergeEducation && len(snapEducation) > 0 {
				masterProfile.Education = append(masterProfile.Education, snapEducation...)
			}
			if req.MergeSkills && len(imported.Skills) > 0 {
				for _, sk := range imported.Skills {
					masterProfile.Skills = append(masterProfile.Skills, career.SkillItem{
						ID:          fmt.Sprintf("sk-%d", time.Now().UnixNano()),
						Name:        sk,
						Proficiency: career.ProficiencyIntermediate,
						Confirmed:   true,
					})
				}
			}
			_, _ = careerService.SaveMasterProfile(r.Context(), user.ID, masterProfile, user.ID, "merged from linkedin import")
		}

		_, _ = linkedinService.MarkProfileImportMerged(r.Context(), imported.ImportID)

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "merged",
			"import_id":    imported.ImportID,
			"snapshot_id":  snapshot.SnapshotID,
			"source_mode":  imported.SourceMode,
			"merged_at":    time.Now().UTC(),
			"experiences":  len(snapExperiences),
			"education":    len(snapEducation),
			"skills":       len(imported.Skills),
		})
	}))

	// POST /api/v1/linkedin/optimize/generate
	mux.HandleFunc("/api/v1/linkedin/optimize/generate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			TargetRole          string   `json:"target_role"`
			CurrentHeadline     string   `json:"current_headline"`
			CurrentAbout        string   `json:"current_about"`
			VerifiedSourceFacts []string `json:"verified_source_facts"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		// Gather existing facts from MasterProfile or LinkedInImportedProfile
		candidateName := "Candidate"
		currentHeadline := req.CurrentHeadline
		currentAbout := req.CurrentAbout
		var exps []linkedin.LinkedInImportedExperience
		var skills []string
		var facts []string = req.VerifiedSourceFacts

		// Fetch from latest LinkedIn import if available
		if imp, err := linkedinService.GetLatestProfileImport(r.Context(), user.ID); err == nil && imp != nil {
			if imp.DisplayName != "" {
				candidateName = imp.DisplayName
			}
			if currentHeadline == "" {
				currentHeadline = imp.Headline
			}
			if currentAbout == "" {
				currentAbout = imp.Summary
			}
			exps = append(exps, imp.Experiences...)
			skills = append(skills, imp.Skills...)
		}

		// Fetch from MasterCareerProfile if available
		if prof, err := careerService.GetMasterProfile(r.Context(), user.ID); err == nil && prof != nil {
			if prof.Contact.FullName != "" {
				candidateName = prof.Contact.FullName
			}
			if currentHeadline == "" {
				currentHeadline = prof.Contact.Headline
			}
			if currentAbout == "" {
				currentAbout = prof.Contact.Summary
			}
			if len(exps) == 0 {
				for _, e := range prof.Experiences {
					exps = append(exps, linkedin.LinkedInImportedExperience{
						CompanyName: e.Company,
						Title:       e.Title,
						Description: e.Description,
						IsCurrent:   e.IsCurrent,
					})
				}
			}
			if len(skills) == 0 {
				for _, sk := range prof.Skills {
					skills = append(skills, sk.Name)
				}
			}
		}

		// Add fallback facts if none provided
		if len(facts) == 0 {
			if currentHeadline != "" {
				facts = append(facts, fmt.Sprintf("Current headline: %s", currentHeadline))
			}
			if len(exps) > 0 {
				facts = append(facts, fmt.Sprintf("Extracted experience at %s as %s", exps[0].CompanyName, exps[0].Title))
			}
		}

		input := linkedin.CandidateOptimizationInput{
			UserID:              user.ID,
			CandidateName:       candidateName,
			TargetRole:          req.TargetRole,
			CurrentHeadline:     currentHeadline,
			CurrentAbout:        currentAbout,
			Experiences:         exps,
			Skills:              skills,
			VerifiedSourceFacts: facts,
		}

		report, err := linkedinService.GenerateOptimizationReport(r.Context(), input)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate optimization report: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(report)
	}))

	// GET /api/v1/linkedin/optimize/latest
	mux.HandleFunc("/api/v1/linkedin/optimize/latest", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		report, err := linkedinService.GetLatestOptimizationReport(r.Context(), user.ID)
		if err != nil {
			http.Error(w, "No optimization report found for user", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// POST /api/v1/linkedin/optimize/approve
	mux.HandleFunc("/api/v1/linkedin/optimize/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ReportID     string `json:"report_id"`
			SuggestionID string `json:"suggestion_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReportID == "" || req.SuggestionID == "" {
			http.Error(w, "report_id and suggestion_id are required", http.StatusBadRequest)
			return
		}

		sug, err := linkedinService.ApproveOptimizationSuggestion(r.Context(), req.ReportID, req.SuggestionID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to approve suggestion: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sug)
	}))

	// POST /api/v1/linkedin/optimize/reject
	mux.HandleFunc("/api/v1/linkedin/optimize/reject", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ReportID     string `json:"report_id"`
			SuggestionID string `json:"suggestion_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReportID == "" || req.SuggestionID == "" {
			http.Error(w, "report_id and suggestion_id are required", http.StatusBadRequest)
			return
		}

		sug, err := linkedinService.RejectOptimizationSuggestion(r.Context(), req.ReportID, req.SuggestionID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to reject suggestion: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sug)
	}))

	// POST /api/v1/linkedin/optimize/custom-edit
	mux.HandleFunc("/api/v1/linkedin/optimize/custom-edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			ReportID     string `json:"report_id"`
			SuggestionID string `json:"suggestion_id"`
			CustomText   string `json:"custom_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ReportID == "" || req.SuggestionID == "" || strings.TrimSpace(req.CustomText) == "" {
			http.Error(w, "report_id, suggestion_id, and custom_text are required", http.StatusBadRequest)
			return
		}

		sug, err := linkedinService.CustomEditOptimizationSuggestion(r.Context(), req.ReportID, req.SuggestionID, req.CustomText)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to record custom edit: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sug)
	}))

	// -------------------------------------------------------------------------
	// LinkedIn Records Engine (IMP-LI-04, AT-011, AT-012, FND-009, FND-012)
	// -------------------------------------------------------------------------

	// Person Records
	mux.HandleFunc("/api/v1/linkedin/records/person", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-default"
		}
		tenantID := user.ID // Scoped per authenticated tenant

		switch r.Method {
		case http.MethodGet:
			recordID := r.URL.Query().Get("record_id")
			if recordID != "" {
				record, err := linkedinService.GetPersonRecord(r.Context(), recordID, tenantID)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to fetch person record: %v", err), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(record)
				return
			}

			filter := linkedin.RecordFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Query:       r.URL.Query().Get("query"),
			}
			records, err := linkedinService.ListPersonRecords(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list person records: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(records)

		case http.MethodPost:
			var p linkedin.PersonRecord
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "Invalid person record payload", http.StatusBadRequest)
				return
			}
			if p.WorkspaceID == "" {
				p.WorkspaceID = workspaceID
			}
			p.TenantID = tenantID
			if p.RecordID == "" {
				p.RecordID = fmt.Sprintf("rec-p-%d", time.Now().UnixNano())
			}

			if err := linkedinService.UpsertPersonRecord(r.Context(), &p); err != nil {
				http.Error(w, fmt.Sprintf("Failed to upsert person record: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(p)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Company Records
	mux.HandleFunc("/api/v1/linkedin/records/company", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-default"
		}
		tenantID := user.ID

		switch r.Method {
		case http.MethodGet:
			recordID := r.URL.Query().Get("record_id")
			if recordID != "" {
				record, err := linkedinService.GetCompanyRecord(r.Context(), recordID, tenantID)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to fetch company record: %v", err), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(record)
				return
			}

			filter := linkedin.RecordFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Query:       r.URL.Query().Get("query"),
			}
			records, err := linkedinService.ListCompanyRecords(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list company records: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(records)

		case http.MethodPost:
			var c linkedin.CompanyRecord
			if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
				http.Error(w, "Invalid company record payload", http.StatusBadRequest)
				return
			}
			if c.WorkspaceID == "" {
				c.WorkspaceID = workspaceID
			}
			c.TenantID = tenantID
			if c.RecordID == "" {
				c.RecordID = fmt.Sprintf("rec-c-%d", time.Now().UnixNano())
			}

			if err := linkedinService.UpsertCompanyRecord(r.Context(), &c); err != nil {
				http.Error(w, fmt.Sprintf("Failed to upsert company record: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(c)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Job Records
	mux.HandleFunc("/api/v1/linkedin/records/job", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-default"
		}
		tenantID := user.ID

		switch r.Method {
		case http.MethodGet:
			recordID := r.URL.Query().Get("record_id")
			if recordID != "" {
				record, err := linkedinService.GetJobRecord(r.Context(), recordID, tenantID)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to fetch job record: %v", err), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(record)
				return
			}

			filter := linkedin.RecordFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Query:       r.URL.Query().Get("query"),
			}
			records, err := linkedinService.ListJobRecords(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list job records: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(records)

		case http.MethodPost:
			var j linkedin.JobRecord
			if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
				http.Error(w, "Invalid job record payload", http.StatusBadRequest)
				return
			}
			if j.WorkspaceID == "" {
				j.WorkspaceID = workspaceID
			}
			j.TenantID = tenantID
			if j.RecordID == "" {
				j.RecordID = fmt.Sprintf("rec-j-%d", time.Now().UnixNano())
			}

			if err := linkedinService.UpsertJobRecord(r.Context(), &j); err != nil {
				http.Error(w, fmt.Sprintf("Failed to upsert job record: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(j)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Post Records
	mux.HandleFunc("/api/v1/linkedin/records/post", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-default"
		}
		tenantID := user.ID

		switch r.Method {
		case http.MethodGet:
			recordID := r.URL.Query().Get("record_id")
			if recordID != "" {
				record, err := linkedinService.GetPostRecord(r.Context(), recordID, tenantID)
				if err != nil {
					http.Error(w, fmt.Sprintf("Failed to fetch post record: %v", err), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(record)
				return
			}

			filter := linkedin.RecordFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Query:       r.URL.Query().Get("query"),
			}
			records, err := linkedinService.ListPostRecords(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list post records: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(records)

		case http.MethodPost:
			var post linkedin.PostRecord
			if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
				http.Error(w, "Invalid post record payload", http.StatusBadRequest)
				return
			}
			if post.WorkspaceID == "" {
				post.WorkspaceID = workspaceID
			}
			post.TenantID = tenantID
			if post.RecordID == "" {
				post.RecordID = fmt.Sprintf("rec-post-%d", time.Now().UnixNano())
			}

			if err := linkedinService.UpsertPostRecord(r.Context(), &post); err != nil {
				http.Error(w, fmt.Sprintf("Failed to upsert post record: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(post)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// -------------------------------------------------------------------------
	// LinkedIn Professional Discovery Engine (IMP-LI-05, LI-05, AT-010, FND-011)
	// -------------------------------------------------------------------------

	// POST /api/v1/linkedin/discovery/search
	mux.HandleFunc("/api/v1/linkedin/discovery/search", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			WorkspaceID string                     `json:"workspace_id"`
			Criteria    linkedin.DiscoveryCriteria `json:"criteria"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid discovery search request payload", http.StatusBadRequest)
			return
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = "ws-default"
		}

		result, err := linkedinService.DiscoverEntities(r.Context(), req.Criteria, user.ID, req.WorkspaceID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Discovery failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))

	// POST /api/v1/linkedin/discovery/boolean-query
	mux.HandleFunc("/api/v1/linkedin/discovery/boolean-query", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var criteria linkedin.DiscoveryCriteria
		if err := json.NewDecoder(r.Body).Decode(&criteria); err != nil {
			http.Error(w, "Invalid criteria payload", http.StatusBadRequest)
			return
		}

		bq, err := linkedin.BuildBooleanQuery(criteria)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to build Boolean query: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(bq)
	}))

	// POST /api/v1/linkedin/discovery/normalize-url
	mux.HandleFunc("/api/v1/linkedin/discovery/normalize-url", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			RawURL string `json:"raw_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.RawURL) == "" {
			http.Error(w, "raw_url is required", http.StatusBadRequest)
			return
		}

		norm, err := linkedin.NormalizeLinkedInURL(req.RawURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("URL normalization failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(norm)
	}))

	// --- Recruiter & Hiring Lead Workspace Endpoints (IMP-LI-06, LI-06, AT-011) ---

	// POST & GET /api/v1/linkedin/recruiter/leads
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.Header.Get("X-Workspace-ID")
		if workspaceID == "" {
			workspaceID = "default-workspace"
		}

		if r.Method == http.MethodGet {
			filter := linkedin.RecruiterLeadFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Company:     r.URL.Query().Get("company"),
				Status:      linkedin.LeadStatus(r.URL.Query().Get("status")),
				SearchQuery: r.URL.Query().Get("q"),
			}
			leads, err := linkedinService.ListRecruiterLeads(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list leads: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(leads)
			return
		}

		if r.Method == http.MethodPost {
			var lead linkedin.RecruiterLead
			if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
				http.Error(w, "Invalid lead payload", http.StatusBadRequest)
				return
			}
			if lead.WorkspaceID == "" {
				lead.WorkspaceID = workspaceID
			}
			if lead.TenantID == "" {
				lead.TenantID = tenantID
			}

			created, err := linkedinService.CreateRecruiterLead(r.Context(), &lead)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create lead: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// POST /api/v1/linkedin/recruiter/leads/from-person
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/from-person", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.Header.Get("X-Workspace-ID")
		if workspaceID == "" {
			workspaceID = "default-workspace"
		}

		var req struct {
			RecordID string `json:"record_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.RecordID) == "" {
			http.Error(w, "record_id is required", http.StatusBadRequest)
			return
		}

		lead, err := linkedinService.CreateLeadFromPersonRecord(r.Context(), req.RecordID, workspaceID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create lead from person: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lead)
	}))

	// POST /api/v1/linkedin/recruiter/leads/from-post
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/from-post", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.Header.Get("X-Workspace-ID")
		if workspaceID == "" {
			workspaceID = "default-workspace"
		}

		var req struct {
			RecordID string `json:"record_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.RecordID) == "" {
			http.Error(w, "record_id is required", http.StatusBadRequest)
			return
		}

		lead, err := linkedinService.CreateLeadFromPostRecord(r.Context(), req.RecordID, workspaceID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create lead from post: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lead)
	}))

	// POST /api/v1/linkedin/recruiter/leads/note
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/note", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			LeadID  string `json:"lead_id"`
			Author  string `json:"author"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.LeadID) == "" || strings.TrimSpace(req.Content) == "" {
			http.Error(w, "lead_id and content are required", http.StatusBadRequest)
			return
		}

		note, err := linkedinService.AddLeadNote(r.Context(), req.LeadID, tenantID, req.Author, req.Content)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to add note: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(note)
	}))

	// POST /api/v1/linkedin/recruiter/leads/reminder
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/reminder", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			LeadID  string    `json:"lead_id"`
			DueDate time.Time `json:"due_date"`
			Message string    `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.LeadID) == "" || strings.TrimSpace(req.Message) == "" {
			http.Error(w, "lead_id and message are required", http.StatusBadRequest)
			return
		}

		rem, err := linkedinService.SetLeadReminder(r.Context(), req.LeadID, tenantID, req.DueDate, req.Message)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to set reminder: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rem)
	}))

	// POST /api/v1/linkedin/recruiter/leads/complete-reminder
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/complete-reminder", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			LeadID string `json:"lead_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.LeadID) == "" {
			http.Error(w, "lead_id is required", http.StatusBadRequest)
			return
		}

		if err := linkedinService.CompleteLeadReminder(r.Context(), req.LeadID, tenantID); err != nil {
			http.Error(w, fmt.Sprintf("Failed to complete reminder: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "completed"})
	}))

	// POST /api/v1/linkedin/recruiter/leads/status
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/status", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			LeadID        string                     `json:"lead_id"`
			Status        linkedin.LeadStatus        `json:"status"`
			OutreachStage linkedin.LeadOutreachStage `json:"outreach_stage"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.LeadID) == "" {
			http.Error(w, "lead_id is required", http.StatusBadRequest)
			return
		}

		lead, err := linkedinService.UpdateLeadStatus(r.Context(), req.LeadID, tenantID, req.Status, req.OutreachStage)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update lead status: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lead)
	}))

	// POST /api/v1/linkedin/recruiter/leads/link
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/link", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			LeadID        string `json:"lead_id"`
			JobID         string `json:"job_id"`
			ApplicationID string `json:"application_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.LeadID) == "" {
			http.Error(w, "lead_id is required", http.StatusBadRequest)
			return
		}

		lead, err := linkedinService.LinkLeadApplication(r.Context(), req.LeadID, tenantID, req.JobID, req.ApplicationID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to link application: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lead)
	}))

	// POST /api/v1/linkedin/recruiter/leads/draft-outreach
	mux.HandleFunc("/api/v1/linkedin/recruiter/leads/draft-outreach", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			LeadID        string `json:"lead_id"`
			CandidateName string `json:"candidate_name"`
			TargetRole    string `json:"target_role"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.LeadID) == "" {
			http.Error(w, "lead_id is required", http.StatusBadRequest)
			return
		}

		lead, err := linkedinService.GetRecruiterLead(r.Context(), req.LeadID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Lead not found: %v", err), http.StatusNotFound)
			return
		}

		draft := lead.GenerateOutreachDraft(req.CandidateName, req.TargetRole)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(draft)
	}))

	// --- Connection Note Drafts & Queue Endpoints (IMP-LI-07, LI-07, AT-007, AT-010, FND-010, FND-011, SRC-L1) ---

	// POST /api/v1/linkedin/queue/draft
	mux.HandleFunc("/api/v1/linkedin/queue/draft", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			CandidateName    string `json:"candidate_name"`
			TargetRole       string `json:"target_role"`
			RecipientName    string `json:"recipient_name"`
			RecipientCompany string `json:"recipient_company"`
			RecipientTitle   string `json:"recipient_title"`
			KeyOverlap       string `json:"key_overlap"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		note, factors := linkedinService.DraftConnectionNote(
			req.CandidateName,
			req.TargetRole,
			req.RecipientName,
			req.RecipientCompany,
			req.RecipientTitle,
			req.KeyOverlap,
		)

		charCount := len([]rune(note))
		resp := map[string]interface{}{
			"note_text":       note,
			"factors":         factors,
			"character_count": charCount,
			"within_limit":   charCount <= linkedin.LinkedInMaxNoteCharacters,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	// POST & GET /api/v1/linkedin/queue
	mux.HandleFunc("/api/v1/linkedin/queue", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		if r.Method == http.MethodGet {
			workspaceID := r.URL.Query().Get("workspace_id")
			status := r.URL.Query().Get("status")
			search := r.URL.Query().Get("query")

			filter := linkedin.ConnectionQueueFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Status:      linkedin.ConnectionQueueStatus(status),
				SearchQuery: search,
			}

			items, err := linkedinService.ListConnectionQueueItems(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list queue items: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(items)
			return
		}

		if r.Method == http.MethodPost {
			var item linkedin.ConnectionQueueItem
			if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
				http.Error(w, "Invalid request payload", http.StatusBadRequest)
				return
			}
			if item.TenantID == "" {
				item.TenantID = tenantID
			}

			created, err := linkedinService.EnqueueConnectionNote(r.Context(), &item)
			if err != nil {
				if errors.Is(err, linkedin.ErrDuplicateRecipient) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, fmt.Sprintf("Failed to enqueue connection note: %v", err), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// GET /api/v1/linkedin/queue/item
	mux.HandleFunc("/api/v1/linkedin/queue/item", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		itemID := r.URL.Query().Get("id")
		if strings.TrimSpace(itemID) == "" {
			http.Error(w, "id query parameter is required", http.StatusBadRequest)
			return
		}

		item, err := linkedinService.GetConnectionQueueItem(r.Context(), itemID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Item not found: %v", err), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(item)
	}))

	// POST /api/v1/linkedin/queue/approve
	mux.HandleFunc("/api/v1/linkedin/queue/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			ID       string `json:"id"`
			Approver string `json:"approver"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ID) == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if req.Approver == "" {
			req.Approver = "Candidate (operator)"
		}

		approved, err := linkedinService.ApproveConnectionQueueItem(r.Context(), req.ID, tenantID, req.Approver)
		if err != nil {
			if errors.Is(err, linkedin.ErrDailyBudgetExceeded) || errors.Is(err, linkedin.ErrWeeklyBudgetExceeded) {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to approve item: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// POST /api/v1/linkedin/queue/edit
	mux.HandleFunc("/api/v1/linkedin/queue/edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			ID       string `json:"id"`
			NoteText string `json:"note_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ID) == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		updated, err := linkedinService.UpdateConnectionNoteText(r.Context(), req.ID, tenantID, req.NoteText)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update note: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// POST /api/v1/linkedin/queue/copy
	mux.HandleFunc("/api/v1/linkedin/queue/copy", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ID) == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		copied, err := linkedinService.MarkConnectionNoteCopied(r.Context(), req.ID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to mark copied: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(copied)
	}))

	// POST /api/v1/linkedin/queue/confirm-sent
	mux.HandleFunc("/api/v1/linkedin/queue/confirm-sent", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ID) == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		sent, err := linkedinService.ConfirmConnectionSent(r.Context(), req.ID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to confirm sent: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(sent)
	}))

	// POST /api/v1/linkedin/queue/reject
	mux.HandleFunc("/api/v1/linkedin/queue/reject", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			ID     string `json:"id"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.ID) == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}

		rejected, err := linkedinService.RejectConnectionQueueItem(r.Context(), req.ID, tenantID, req.Reason)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to reject item: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rejected)
	}))

	// GET & PUT /api/v1/linkedin/queue/budget
	mux.HandleFunc("/api/v1/linkedin/queue/budget", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		if r.Method == http.MethodGet {
			budget, err := linkedinService.GetConnectionBudget(r.Context(), workspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get budget: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(budget)
			return
		}

		if r.Method == http.MethodPut {
			var req struct {
				WorkspaceID string `json:"workspace_id"`
				DailyLimit  int    `json:"daily_limit"`
				WeeklyLimit int    `json:"weekly_limit"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			if req.WorkspaceID == "" {
				req.WorkspaceID = workspaceID
			}

			budget, err := linkedinService.GetConnectionBudget(r.Context(), req.WorkspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get budget: %v", err), http.StatusInternalServerError)
				return
			}

			if req.DailyLimit > 0 {
				if req.DailyLimit > linkedin.MaxSafeDailyInviteLimit {
					req.DailyLimit = linkedin.MaxSafeDailyInviteLimit
				}
				budget.DailyLimit = req.DailyLimit
			}
			if req.WeeklyLimit > 0 {
				if req.WeeklyLimit > linkedin.MaxSafeWeeklyInviteLimit {
					req.WeeklyLimit = linkedin.MaxSafeWeeklyInviteLimit
				}
				budget.WeeklyLimit = req.WeeklyLimit
			}

			if err := linkedinService.SaveConnectionBudget(r.Context(), budget); err != nil {
				http.Error(w, fmt.Sprintf("Failed to save budget: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(budget)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// =========================================================================
	// IMP-LI-08: Company-Follow Planning Endpoints (LI-08, AT-010, FND-011, SRC-L1)
	// =========================================================================

	// 1. Target Company Watchlist
	mux.HandleFunc("/api/v1/linkedin/company-follow/watchlist", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		if r.Method == http.MethodGet {
			statusFilter := linkedin.WatchlistStatus(r.URL.Query().Get("status"))
			items, err := linkedinService.GetCompanyWatchlist(r.Context(), workspaceID, tenantID, statusFilter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list watchlist: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace_id": workspaceID,
				"total":        len(items),
				"items":        items,
			})
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				WorkspaceID     string                         `json:"workspace_id"`
				CompanyRecordID string                         `json:"company_record_id"`
				CompanyName     string                         `json:"company_name"`
				UniversalName   string                         `json:"universal_name"`
				Domain          string                         `json:"domain"`
				Priority        linkedin.CompanyFollowPriority `json:"priority"`
				TargetReason    string                         `json:"target_reason"`
				Tags            []string                       `json:"tags"`
				Status          linkedin.WatchlistStatus       `json:"status"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			if req.WorkspaceID == "" {
				req.WorkspaceID = workspaceID
			}

			item := &linkedin.CompanyWatchlistItem{
				WorkspaceID:     req.WorkspaceID,
				TenantID:        tenantID,
				CompanyRecordID: req.CompanyRecordID,
				CompanyName:     req.CompanyName,
				UniversalName:   req.UniversalName,
				Domain:          req.Domain,
				Priority:        req.Priority,
				TargetReason:    req.TargetReason,
				Tags:            req.Tags,
				Status:          req.Status,
			}

			saved, err := linkedinService.AddCompanyToWatchlist(r.Context(), item)
			if err != nil {
				if errors.Is(err, linkedin.ErrWatchlistCompanyDuplicate) {
					http.Error(w, err.Error(), http.StatusConflict)
					return
				}
				http.Error(w, fmt.Sprintf("Failed to add to watchlist: %v", err), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(saved)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 2. Watchlist Item Deletion
	mux.HandleFunc("/api/v1/linkedin/company-follow/watchlist/item", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method == http.MethodDelete {
			itemID := r.URL.Query().Get("item_id")
			if itemID == "" {
				http.Error(w, "item_id parameter required", http.StatusBadRequest)
				return
			}
			if err := linkedinService.DeleteWatchlistItem(r.Context(), itemID, tenantID); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete watchlist item: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "deleted",
				"item_id": itemID,
			})
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 3. Batch Follow Plans
	mux.HandleFunc("/api/v1/linkedin/company-follow/plans", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		if r.Method == http.MethodGet {
			plans, err := linkedinService.ListBatchFollowPlans(r.Context(), workspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list plans: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace_id": workspaceID,
				"total":        len(plans),
				"plans":        plans,
			})
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				WorkspaceID       string   `json:"workspace_id"`
				PlanName          string   `json:"plan_name"`
				ItemIDs           []string `json:"item_ids"`
				PacingIntervalSec int      `json:"pacing_interval_sec"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			if req.WorkspaceID == "" {
				req.WorkspaceID = workspaceID
			}

			plan, err := linkedinService.CreateBatchFollowPlan(r.Context(), req.WorkspaceID, tenantID, req.PlanName, req.ItemIDs, req.PacingIntervalSec)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create batch plan: %v", err), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(plan)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 4. Batch Plan Detail
	mux.HandleFunc("/api/v1/linkedin/company-follow/plans/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method == http.MethodGet {
			planID := r.URL.Query().Get("plan_id")
			if planID == "" {
				http.Error(w, "plan_id parameter required", http.StatusBadRequest)
				return
			}
			plan, err := linkedinService.GetBatchFollowPlan(r.Context(), planID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Plan not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(plan)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 5. Confirm Follow (Consumes 1 budget unit, updates plan item)
	mux.HandleFunc("/api/v1/linkedin/company-follow/confirm", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PlanID string `json:"plan_id"`
			ItemID string `json:"item_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		plan, err := linkedinService.ConfirmCompanyFollowed(r.Context(), req.PlanID, req.ItemID, tenantID)
		if err != nil {
			if errors.Is(err, linkedin.ErrCompanyFollowBudgetExhausted) || errors.Is(err, linkedin.ErrWeeklyFollowBudgetExhausted) {
				http.Error(w, err.Error(), http.StatusTooManyRequests)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to confirm follow: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(plan)
	}))

	// 6. Skip Follow Item
	mux.HandleFunc("/api/v1/linkedin/company-follow/skip", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PlanID string `json:"plan_id"`
			ItemID string `json:"item_id"`
			Reason string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		plan, err := linkedinService.SkipCompanyFollow(r.Context(), req.PlanID, req.ItemID, tenantID, req.Reason)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to skip item: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(plan)
	}))

	// 7. Follow Budget Endpoint
	mux.HandleFunc("/api/v1/linkedin/company-follow/budget", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		if r.Method == http.MethodGet {
			budget, err := linkedinService.GetCompanyFollowBudget(r.Context(), workspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get follow budget: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(budget)
			return
		}

		if r.Method == http.MethodPut {
			var req struct {
				WorkspaceID string `json:"workspace_id"`
				DailyLimit  int    `json:"daily_limit"`
				WeeklyLimit int    `json:"weekly_limit"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			if req.WorkspaceID == "" {
				req.WorkspaceID = workspaceID
			}

			budget, err := linkedinService.GetCompanyFollowBudget(r.Context(), req.WorkspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get follow budget: %v", err), http.StatusInternalServerError)
				return
			}

			if req.DailyLimit > 0 {
				if req.DailyLimit > 25 {
					req.DailyLimit = 25 // Conservative safety cap
				}
				budget.DailyLimit = req.DailyLimit
			}
			if req.WeeklyLimit > 0 {
				if req.WeeklyLimit > 70 {
					req.WeeklyLimit = 70 // Conservative safety cap
				}
				budget.WeeklyLimit = req.WeeklyLimit
			}

			if err := linkedinService.UpdateCompanyFollowBudget(r.Context(), budget); err != nil {
				http.Error(w, fmt.Sprintf("Failed to update follow budget: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(budget)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// --- Inbox & Conversation Triage Endpoints (IMP-LI-09, LI-09, AT-007, AT-010, AT-011, FND-010, SRC-C3) ---

	// 1. List / Search Conversation Threads
	mux.HandleFunc("/api/v1/linkedin/inbox/threads", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		if r.Method == http.MethodGet {
			filter := linkedin.ConversationThreadFilter{
				WorkspaceID:    workspaceID,
				TenantID:       tenantID,
				ThreadType:     linkedin.InboxThreadType(r.URL.Query().Get("thread_type")),
				Classification: linkedin.InboxClassification(r.URL.Query().Get("classification")),
				Query:          r.URL.Query().Get("query"),
				OnlyUnread:     r.URL.Query().Get("only_unread") == "true",
			}

			threads, err := linkedinService.ListConversationThreads(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list threads: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace_id": workspaceID,
				"total":        len(threads),
				"threads":      threads,
			})
			return
		}

		if r.Method == http.MethodPost {
			var req linkedin.ConversationThread
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			if req.WorkspaceID == "" {
				req.WorkspaceID = workspaceID
			}
			req.TenantID = tenantID

			saved, err := linkedinService.SaveConversationThread(r.Context(), &req)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to save thread: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(saved)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 2. Thread Detail
	mux.HandleFunc("/api/v1/linkedin/inbox/threads/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method == http.MethodGet {
			threadID := r.URL.Query().Get("thread_id")
			if threadID == "" {
				http.Error(w, "thread_id parameter required", http.StatusBadRequest)
				return
			}
			thread, err := linkedinService.GetConversationThread(r.Context(), threadID, tenantID)
			if err != nil {
				if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Thread not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(thread)
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 3. Add Message to Thread (halts active follow-up schedules on recruiter reply)
	mux.HandleFunc("/api/v1/linkedin/inbox/threads/message", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ThreadID   string               `json:"thread_id"`
			Message    linkedin.InboxMessage `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ThreadID == "" {
			http.Error(w, "Invalid message payload or missing thread_id", http.StatusBadRequest)
			return
		}

		updatedThread, halted, err := linkedinService.AddMessageToThread(r.Context(), req.ThreadID, tenantID, req.Message)
		if err != nil {
			if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to add message: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"thread":          updatedThread,
			"followup_halted": halted,
		})
	}))

	// 4. Generate Reply Draft
	mux.HandleFunc("/api/v1/linkedin/inbox/reply/generate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ThreadID              string                  `json:"thread_id"`
			Tone                  linkedin.ReplyDraftTone `json:"tone"`
			CandidateAvailability string                  `json:"candidate_availability"`
			VerifiedFacts         []string                `json:"verified_facts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ThreadID == "" {
			http.Error(w, "Invalid payload or missing thread_id", http.StatusBadRequest)
			return
		}

		draft, err := linkedinService.GenerateReplyDraft(r.Context(), req.ThreadID, tenantID, req.Tone, req.CandidateAvailability, req.VerifiedFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(draft)
	}))

	// 5. Approve Reply Draft (HMAC-SHA256 approval token)
	mux.HandleFunc("/api/v1/linkedin/inbox/reply/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		approver := "user"
		if user != nil && user.Email != "" {
			approver = user.Email
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		approved, err := linkedinService.ApproveReplyDraft(r.Context(), req.DraftID, tenantID, approver)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to approve draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 6. Edit Reply Draft Text (Tamper Invalidation)
	mux.HandleFunc("/api/v1/linkedin/inbox/reply/edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
			NewText string `json:"new_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		updated, err := linkedinService.UpdateReplyDraftText(r.Context(), req.DraftID, tenantID, req.NewText)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to edit draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// 7. Mark Reply Draft Copied (AT-010 1-Click Clipboard Copy)
	mux.HandleFunc("/api/v1/linkedin/inbox/reply/copy", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		copied, err := linkedinService.MarkReplyDraftCopied(r.Context(), req.DraftID, tenantID)
		if err != nil {
			if errors.Is(err, linkedin.ErrInvalidReplyApprovalToken) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to copy draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(copied)
	}))

	// 8. Reject Reply Draft
	mux.HandleFunc("/api/v1/linkedin/inbox/reply/reject", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
			Reason  string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		rejected, err := linkedinService.RejectReplyDraft(r.Context(), req.DraftID, tenantID, req.Reason)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to reject draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rejected)
	}))

	// 9. List Reply Drafts for Thread
	mux.HandleFunc("/api/v1/linkedin/inbox/reply/list", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method == http.MethodGet {
			threadID := r.URL.Query().Get("thread_id")
			if threadID == "" {
				http.Error(w, "thread_id parameter required", http.StatusBadRequest)
				return
			}
			drafts, err := linkedinService.ListReplyDraftsForThread(r.Context(), threadID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list drafts: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"thread_id": threadID,
				"total":     len(drafts),
				"drafts":    drafts,
			})
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// --- Comments, Replies & Thread Sweep Endpoints (IMP-LI-10, LI-10, AT-007, AT-010, AT-011, FND-010, SRC-L1, SRC-L2) ---

	// 1. List / Save Swept Target Posts
	mux.HandleFunc("/api/v1/linkedin/comments/posts", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		if r.Method == http.MethodGet {
			filter := linkedin.SweptPostFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Category:    linkedin.PostSweepCategory(r.URL.Query().Get("category")),
				Query:       r.URL.Query().Get("query"),
			}
			posts, err := linkedinService.ListSweptTargetPosts(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list posts: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace_id": workspaceID,
				"total":        len(posts),
				"posts":        posts,
			})
			return
		}

		if r.Method == http.MethodPost {
			var post linkedin.SweptTargetPost
			if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			if post.WorkspaceID == "" {
				post.WorkspaceID = workspaceID
			}
			post.TenantID = tenantID
			if err := linkedinService.SaveSweptTargetPost(r.Context(), &post); err != nil {
				http.Error(w, fmt.Sprintf("Failed to save post: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(post)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 2. Swept Target Post Detail
	mux.HandleFunc("/api/v1/linkedin/comments/posts/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		postID := r.URL.Query().Get("post_id")
		if postID == "" {
			http.Error(w, "post_id parameter required", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			post, err := linkedinService.GetSweptTargetPost(r.Context(), postID, tenantID)
			if err != nil {
				if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Post not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(post)
			return
		}

		if r.Method == http.MethodDelete {
			if err := linkedinService.DeleteSweptTargetPost(r.Context(), postID, tenantID); err != nil {
				if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Failed to delete post: %v", err), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 3. Generate Comment Drafts
	mux.HandleFunc("/api/v1/linkedin/comments/drafts/generate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			PostID          string   `json:"post_id"`
			TargetCommentID string   `json:"target_comment_id,omitempty"`
			CandidateFacts  []string `json:"candidate_facts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PostID == "" {
			http.Error(w, "Invalid payload or missing post_id", http.StatusBadRequest)
			return
		}

		drafts, err := linkedinService.GenerateCommentDrafts(r.Context(), req.PostID, tenantID, req.TargetCommentID, req.CandidateFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate comment drafts: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(drafts)
	}))

	// 4. Approve Comment Draft
	mux.HandleFunc("/api/v1/linkedin/comments/drafts/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		approver := "user"
		if user != nil && user.Email != "" {
			approver = user.Email
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		approved, err := linkedinService.ApproveCommentDraft(r.Context(), req.DraftID, tenantID, approver)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to approve draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 5. Edit Comment Draft Text
	mux.HandleFunc("/api/v1/linkedin/comments/drafts/edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
			NewText string `json:"new_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		updated, err := linkedinService.UpdateCommentDraftText(r.Context(), req.DraftID, tenantID, req.NewText)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update draft text: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// 6. Mark Comment Draft Copied
	mux.HandleFunc("/api/v1/linkedin/comments/drafts/copy", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		copied, err := linkedinService.MarkCommentDraftCopied(r.Context(), req.DraftID, tenantID)
		if err != nil {
			if errors.Is(err, linkedin.ErrInvalidCommentApprovalToken) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to copy draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(copied)
	}))

	// 7. Reject Comment Draft
	mux.HandleFunc("/api/v1/linkedin/comments/drafts/reject", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
			Reason  string `json:"reason,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		rejected, err := linkedinService.RejectCommentDraft(r.Context(), req.DraftID, tenantID, req.Reason)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to reject draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rejected)
	}))

	// 8. List Comment Drafts for Post
	mux.HandleFunc("/api/v1/linkedin/comments/drafts/list", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method == http.MethodGet {
			postID := r.URL.Query().Get("post_id")
			if postID == "" {
				http.Error(w, "post_id parameter required", http.StatusBadRequest)
				return
			}
			drafts, err := linkedinService.ListCommentDraftsForPost(r.Context(), postID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list drafts: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"post_id": postID,
				"total":   len(drafts),
				"drafts":  drafts,
			})
			return
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// --- Post Writing, Hooks & Audits Endpoints (IMP-LI-11, LI-11, AT-003, AT-019, FND-015, SRC-L2) ---

	// 1. List / Save Post Drafts
	mux.HandleFunc("/api/v1/linkedin/posts/drafts", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		if r.Method == http.MethodGet {
			filter := linkedin.PostDraftFilter{
				WorkspaceID: workspaceID,
				TenantID:    tenantID,
				Angle:       linkedin.PostAngle(r.URL.Query().Get("angle")),
				Status:      linkedin.PostDraftStatus(r.URL.Query().Get("status")),
				Query:       r.URL.Query().Get("query"),
			}
			drafts, err := linkedinService.ListPostDrafts(r.Context(), filter)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list post drafts: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"workspace_id": workspaceID,
				"total":        len(drafts),
				"drafts":       drafts,
			})
			return
		}

		if r.Method == http.MethodPost {
			var draft linkedin.PostDraft
			if err := json.NewDecoder(r.Body).Decode(&draft); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			if draft.WorkspaceID == "" {
				draft.WorkspaceID = workspaceID
			}
			draft.TenantID = tenantID
			if err := linkedinService.SavePostDraft(r.Context(), &draft); err != nil {
				http.Error(w, fmt.Sprintf("Failed to save post draft: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(draft)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 2. Post Draft Detail (Get & Delete)
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		draftID := r.URL.Query().Get("draft_id")
		if draftID == "" {
			http.Error(w, "draft_id parameter required", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			draft, err := linkedinService.GetPostDraft(r.Context(), draftID, tenantID)
			if err != nil {
				if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Post draft not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(draft)
			return
		}

		if r.Method == http.MethodDelete {
			if err := linkedinService.DeletePostDraft(r.Context(), draftID, tenantID); err != nil {
				if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
					http.Error(w, err.Error(), http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Failed to delete post draft: %v", err), http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 3. Generate Post Draft
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/generate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req linkedin.PostWritingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = "ws-alpha"
		}
		req.TenantID = tenantID

		draft, err := linkedinService.GeneratePostDraft(r.Context(), req)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate post draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(draft)
	}))

	// 4. Generate Hook Variants for Post
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/hooks", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		hooks, err := linkedinService.GenerateHookVariantsForDraft(r.Context(), req.DraftID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate hook variants: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(hooks)
	}))

	// 5. Run Editorial Audit on Post Draft
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/audit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		audit, err := linkedinService.AuditPostDraftContent(r.Context(), req.DraftID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to run editorial audit: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(audit)
	}))

	// 6. Approve Post Draft (HMAC-SHA256 token)
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		user, _ := identity.GetUserFromContext(r.Context())
		approver := "user"
		if user != nil && user.Email != "" {
			approver = user.Email
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		approved, err := linkedinService.ApprovePostDraft(r.Context(), req.DraftID, tenantID, approver)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to approve draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 7. Edit Post Draft Text (Tamper Invalidation)
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID          string `json:"draft_id"`
			SelectedHookType string `json:"selected_hook_type,omitempty"`
			SelectedHookText string `json:"selected_hook_text,omitempty"`
			FullPostText     string `json:"full_post_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		updated, err := linkedinService.UpdatePostDraftText(r.Context(), req.DraftID, tenantID, req.SelectedHookType, req.SelectedHookText, req.FullPostText)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to edit post draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// 8. Mark Post Draft Copied (AT-010 1-Click Clipboard Copy)
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/copy", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		copied, err := linkedinService.MarkPostDraftCopied(r.Context(), req.DraftID, tenantID)
		if err != nil {
			if errors.Is(err, linkedin.ErrInvalidPostApprovalToken) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to copy post draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(copied)
	}))

	// 9. Reject Post Draft
	mux.HandleFunc("/api/v1/linkedin/posts/drafts/reject", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID string `json:"draft_id"`
			Reason  string `json:"reason,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		rejected, err := linkedinService.RejectPostDraft(r.Context(), req.DraftID, tenantID, req.Reason)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to reject post draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rejected)
	}))

	// ==========================================
	// IMP-LI-12: Humanizer and Reusable Voice API
	// ==========================================

	// 1. Create or Update Voice Profile
	mux.HandleFunc("/api/v1/linkedin/voice-profiles", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var profile linkedin.VoiceProfile
		if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		profile.TenantID = tenantID
		if profile.WorkspaceID == "" {
			profile.WorkspaceID = "default-ws"
		}

		var saved *linkedin.VoiceProfile
		var err error
		if profile.ProfileID == "" {
			saved, err = linkedinService.CreateVoiceProfile(r.Context(), &profile)
		} else {
			existing, getErr := linkedinService.GetVoiceProfile(r.Context(), profile.ProfileID, tenantID)
			if getErr == nil && existing != nil {
				saved, err = linkedinService.UpdateVoiceProfile(r.Context(), &profile, tenantID)
			} else {
				saved, err = linkedinService.CreateVoiceProfile(r.Context(), &profile)
			}
		}

		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to save voice profile: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(saved)
	}))

	// 2. Get or Delete Voice Profile Detail
	mux.HandleFunc("/api/v1/linkedin/voice-profiles/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		profileID := r.URL.Query().Get("profile_id")
		if profileID == "" {
			http.Error(w, "Missing profile_id parameter", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			profile, err := linkedinService.GetVoiceProfile(r.Context(), profileID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Voice profile not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(profile)

		case http.MethodDelete:
			err := linkedinService.DeleteVoiceProfile(r.Context(), profileID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete voice profile: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]bool{"deleted": true})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// 3. List Voice Profiles
	mux.HandleFunc("/api/v1/linkedin/voice-profiles/list", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		profiles, err := linkedinService.ListVoiceProfiles(r.Context(), workspaceID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list voice profiles: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(profiles)
	}))

	// 4. Approve Voice Profile (HMAC-SHA256)
	mux.HandleFunc("/api/v1/linkedin/voice-profiles/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ProfileID string `json:"profile_id"`
			Approver  string `json:"approver,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProfileID == "" {
			http.Error(w, "Invalid payload or missing profile_id", http.StatusBadRequest)
			return
		}

		approver := req.Approver
		if approver == "" {
			user, _ := identity.GetUserFromContext(r.Context())
			if user != nil {
				approver = user.Email
			} else {
				approver = "current-user"
			}
		}

		approved, err := linkedinService.ApproveVoiceProfile(r.Context(), req.ProfileID, approver, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to approve voice profile: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 5. Humanize Draft Text
	mux.HandleFunc("/api/v1/linkedin/humanizer/process", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			RawText                string   `json:"raw_text"`
			VoiceProfileID         string   `json:"voice_profile_id"`
			WorkspaceID            string   `json:"workspace_id,omitempty"`
			CandidateVerifiedFacts []string `json:"candidate_verified_facts,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.RawText == "" {
			http.Error(w, "Invalid payload or missing raw_text", http.StatusBadRequest)
			return
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = "default-ws"
		}

		result, err := linkedinService.HumanizeDraftContent(r.Context(), req.RawText, req.VoiceProfileID, req.WorkspaceID, tenantID, req.CandidateVerifiedFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to humanize draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))

	// 6. Approve Humanized Draft
	mux.HandleFunc("/api/v1/linkedin/humanizer/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ResultID string `json:"result_id"`
			Approver string `json:"approver,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ResultID == "" {
			http.Error(w, "Invalid payload or missing result_id", http.StatusBadRequest)
			return
		}

		approver := req.Approver
		if approver == "" {
			user, _ := identity.GetUserFromContext(r.Context())
			if user != nil {
				approver = user.Email
			} else {
				approver = "current-user"
			}
		}

		approved, err := linkedinService.ApproveHumanizedDraft(r.Context(), req.ResultID, approver, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to approve humanized draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 7. Edit Humanized Text (Tamper Invalidation)
	mux.HandleFunc("/api/v1/linkedin/humanizer/edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ResultID               string   `json:"result_id"`
			UpdatedText            string   `json:"updated_text"`
			CandidateVerifiedFacts []string `json:"candidate_verified_facts,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ResultID == "" || req.UpdatedText == "" {
			http.Error(w, "Invalid payload or missing result_id / updated_text", http.StatusBadRequest)
			return
		}

		updated, err := linkedinService.UpdateHumanizedText(r.Context(), req.ResultID, req.UpdatedText, tenantID, req.CandidateVerifiedFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update humanized text: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// 8. Mark Humanized Draft Copied (AT-010 1-Click Clipboard Copy)
	mux.HandleFunc("/api/v1/linkedin/humanizer/copy", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ResultID string `json:"result_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ResultID == "" {
			http.Error(w, "Invalid payload or missing result_id", http.StatusBadRequest)
			return
		}

		copied, err := linkedinService.MarkHumanizedDraftCopied(r.Context(), req.ResultID, tenantID)
		if err != nil {
			if errors.Is(err, linkedin.ErrInvalidApprovalToken) {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to copy humanized draft: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(copied)
	}))

	// =========================================================================
	// Story Bank & Guided Interviewer APIs (IMP-LI-13, LI-13, AT-003, SRC-L2)
	// =========================================================================

	// 1. Get Curated Interview Prompts
	mux.HandleFunc("/api/v1/linkedin/interviewer/prompts", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		prompts, err := linkedinService.GetCuratedInterviewPrompts(r.Context())
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch prompts: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(prompts)
	}))

	// 2. Start Interview Session
	mux.HandleFunc("/api/v1/linkedin/interviewer/start", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			WorkspaceID string                 `json:"workspace_id"`
			Category    linkedin.StoryCategory `json:"category"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.WorkspaceID == "" {
			http.Error(w, "Invalid payload or missing workspace_id", http.StatusBadRequest)
			return
		}
		if req.Category == "" {
			req.Category = linkedin.StoryCategoryTurningPoint
		}

		session, err := linkedinService.StartInterviewSession(r.Context(), req.WorkspaceID, tenantID, req.Category)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to start interview: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(session)
	}))

	// 3. Record Interview Answer
	mux.HandleFunc("/api/v1/linkedin/interviewer/answer", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SessionID string `json:"session_id"`
			PromptID  string `json:"prompt_id"`
			Answer    string `json:"answer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" || req.PromptID == "" {
			http.Error(w, "Invalid payload or missing session_id/prompt_id", http.StatusBadRequest)
			return
		}

		session, err := linkedinService.RecordInterviewAnswer(r.Context(), req.SessionID, tenantID, req.PromptID, req.Answer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to record answer: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(session)
	}))

	// 4. Record Interview Follow-Up Answer
	mux.HandleFunc("/api/v1/linkedin/interviewer/follow-up", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SessionID      string `json:"session_id"`
			PromptID       string `json:"prompt_id"`
			FollowUpAnswer string `json:"follow_up_answer"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" || req.PromptID == "" {
			http.Error(w, "Invalid payload or missing session_id/prompt_id", http.StatusBadRequest)
			return
		}

		session, err := linkedinService.RecordInterviewFollowUpAnswer(r.Context(), req.SessionID, tenantID, req.PromptID, req.FollowUpAnswer)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to record follow-up answer: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(session)
	}))

	// 5. Synthesize Interview into StoryEntry
	mux.HandleFunc("/api/v1/linkedin/interviewer/synthesize", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			SessionID     string   `json:"session_id"`
			AuthorName    string   `json:"author_name"`
			VerifiedFacts []string `json:"verified_facts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
			http.Error(w, "Invalid payload or missing session_id", http.StatusBadRequest)
			return
		}

		entry, err := linkedinService.SynthesizeInterviewStory(r.Context(), req.SessionID, tenantID, req.AuthorName, req.VerifiedFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to synthesize story: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(entry)
	}))

	// 6. Create Story Entry (Manual or Direct)
	mux.HandleFunc("/api/v1/linkedin/stories", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Entry         linkedin.StoryEntry `json:"entry"`
			VerifiedFacts []string            `json:"verified_facts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		req.Entry.TenantID = tenantID
		created, err := linkedinService.CreateStoryEntry(r.Context(), &req.Entry, req.VerifiedFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create story: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	}))

	// 7. Get Detail or Delete Story Entry
	mux.HandleFunc("/api/v1/linkedin/stories/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		storyID := r.URL.Query().Get("id")
		if storyID == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			entry, err := linkedinService.GetStoryEntry(r.Context(), storyID, tenantID)
			if err != nil {
				if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Failed to get story: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(entry)

		case http.MethodDelete:
			err := linkedinService.DeleteStoryEntry(r.Context(), storyID, tenantID)
			if err != nil {
				if errors.Is(err, linkedin.ErrCrossTenantAccessDenied) {
					http.Error(w, "Forbidden", http.StatusForbidden)
					return
				}
				http.Error(w, fmt.Sprintf("Failed to delete story: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]bool{"deleted": true})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// 8. List Story Entries
	mux.HandleFunc("/api/v1/linkedin/stories/list", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		filter := linkedin.StoryFilter{
			WorkspaceID:        workspaceID,
			RequestingTenantID: tenantID,
			Category:           linkedin.StoryCategory(r.URL.Query().Get("category")),
			Status:             linkedin.StoryStatus(r.URL.Query().Get("status")),
		}

		entries, err := linkedinService.ListStoryEntries(r.Context(), filter)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list stories: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(entries)
	}))

	// 9. Approve Story Entry (HMAC-SHA256 & AT-003 Zero-Hallucination Guard)
	mux.HandleFunc("/api/v1/linkedin/stories/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			StoryID       string   `json:"story_id"`
			SecretKey     string   `json:"secret_key"`
			VerifiedFacts []string `json:"verified_facts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StoryID == "" {
			http.Error(w, "Invalid payload or missing story_id", http.StatusBadRequest)
			return
		}

		approved, err := linkedinService.ApproveStoryEntry(r.Context(), req.StoryID, tenantID, req.SecretKey, req.VerifiedFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Story approval failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 10. Edit Story Entry (Tamper Invalidation AT-007)
	mux.HandleFunc("/api/v1/linkedin/stories/edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			StoryID       string                  `json:"story_id"`
			Narrative     linkedin.StoryNarrative `json:"narrative"`
			VerifiedFacts []string                `json:"verified_facts"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StoryID == "" {
			http.Error(w, "Invalid payload or missing story_id", http.StatusBadRequest)
			return
		}

		updated, err := linkedinService.UpdateStoryEntry(r.Context(), req.StoryID, tenantID, req.Narrative, req.VerifiedFacts)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update story: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// 11. Copy Story Entry (AT-010 1-Click Clipboard Copy with Markdown)
	mux.HandleFunc("/api/v1/linkedin/stories/copy", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			StoryID   string `json:"story_id"`
			SecretKey string `json:"secret_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StoryID == "" {
			http.Error(w, "Invalid payload or missing story_id", http.StatusBadRequest)
			return
		}

		entry, md, err := linkedinService.MarkStoryEntryCopied(r.Context(), req.StoryID, tenantID, req.SecretKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to copy story: %v", err), http.StatusBadRequest)
			return
		}

		resp := map[string]interface{}{
			"story":             entry,
			"markdown":          md,
			"copied":            true,
			"linkedin_deep_link": "https://www.linkedin.com/feed/?shareActive=true",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))

	// 12. Archive Story Entry
	mux.HandleFunc("/api/v1/linkedin/stories/archive", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			StoryID string `json:"story_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.StoryID == "" {
			http.Error(w, "Invalid payload or missing story_id", http.StatusBadRequest)
			return
		}

		archived, err := linkedinService.ArchiveStoryEntry(r.Context(), req.StoryID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to archive story: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(archived)
	}))

	// =========================================================================
	// LinkedIn Content Planning & Repurposing (IMP-LI-14, LI-14, AT-018, AT-010, AT-007, FND-008, SRC-L2)
	// =========================================================================

	// 1. Repurpose Source Artifact (Single format or All 5 canonical formats)
	mux.HandleFunc("/api/v1/linkedin/content/repurpose", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			WorkspaceID string                  `json:"workspace_id"`
			Source      linkedin.SourceArtifact `json:"source"`
			Format      string                  `json:"format,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.WorkspaceID == "" {
			http.Error(w, "Invalid payload or missing workspace_id", http.StatusBadRequest)
			return
		}

		req.Source.WorkspaceID = req.WorkspaceID
		req.Source.TenantID = tenantID
		if req.Source.CreatedAt.IsZero() {
			req.Source.CreatedAt = time.Now().UTC()
		}

		if req.Format != "" {
			draft, err := linkedinService.RepurposeArtifact(r.Context(), req.Source, linkedin.RepurposedFormat(req.Format))
			if err != nil {
				http.Error(w, fmt.Sprintf("Repurposing failed: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"draft":  draft,
				"drafts": []*linkedin.RepurposedDraft{draft},
			})
			return
		}

		drafts, err := linkedinService.RepurposeArtifactAllFormats(r.Context(), req.Source)
		if err != nil {
			http.Error(w, fmt.Sprintf("Repurposing all formats failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":  len(drafts),
			"drafts": drafts,
		})
	}))

	// 2. List Repurposed Drafts
	mux.HandleFunc("/api/v1/linkedin/content/drafts", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			http.Error(w, "workspace_id query parameter is required", http.StatusBadRequest)
			return
		}

		drafts, err := linkedinService.ListRepurposedDrafts(r.Context(), workspaceID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list drafts: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":  len(drafts),
			"drafts": drafts,
		})
	}))

	// 3. Draft Details, Mutation (AT-007) and Deletion
	mux.HandleFunc("/api/v1/linkedin/content/drafts/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		switch r.Method {
		case http.MethodGet:
			draftID := r.URL.Query().Get("draft_id")
			if draftID == "" {
				http.Error(w, "draft_id query parameter is required", http.StatusBadRequest)
				return
			}
			draft, err := linkedinService.GetRepurposedDraft(r.Context(), draftID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Draft not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(draft)

		case http.MethodPut:
			var req struct {
				DraftID     string `json:"draft_id"`
				Title       string `json:"title"`
				ContentBody string `json:"content_body"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
				http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
				return
			}
			updated, err := linkedinService.UpdateRepurposedDraft(r.Context(), req.DraftID, tenantID, req.Title, req.ContentBody)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to update draft: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updated)

		case http.MethodDelete:
			draftID := r.URL.Query().Get("draft_id")
			if draftID == "" {
				http.Error(w, "draft_id query parameter is required", http.StatusBadRequest)
				return
			}
			if err := linkedinService.DeleteRepurposedDraft(r.Context(), draftID, tenantID); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete draft: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]bool{"deleted": true})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// 4. Approve Draft (AT-007 HMAC Token Generation)
	mux.HandleFunc("/api/v1/linkedin/content/drafts/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			DraftID   string `json:"draft_id"`
			SecretKey string `json:"secret_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.DraftID == "" {
			http.Error(w, "Invalid payload or missing draft_id", http.StatusBadRequest)
			return
		}

		approved, err := linkedinService.ApproveRepurposedDraft(r.Context(), req.DraftID, tenantID, req.SecretKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Approval failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 5. Schedule Content Plan Item (AT-018 DST-safe calendar slotting)
	mux.HandleFunc("/api/v1/linkedin/content/calendar/schedule", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			WorkspaceID        string `json:"workspace_id"`
			DraftID            string `json:"draft_id"`
			RequestedLocalTime string `json:"requested_local_time"`
			UserTimezone       string `json:"user_timezone"`
			MaxDailyBudget     int    `json:"max_daily_budget"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.WorkspaceID == "" || req.DraftID == "" {
			http.Error(w, "Invalid payload, workspace_id, or draft_id", http.StatusBadRequest)
			return
		}

		parsedTime, err := time.Parse(time.RFC3339, req.RequestedLocalTime)
		if err != nil {
			// Fallback try simple format
			parsedTime, err = time.Parse("2006-01-02T15:04:05", req.RequestedLocalTime)
			if err != nil {
				parsedTime = time.Now().Add(24 * time.Hour)
			}
		}

		if req.UserTimezone == "" {
			req.UserTimezone = "UTC"
		}
		if req.MaxDailyBudget <= 0 {
			req.MaxDailyBudget = 3
		}

		item, err := linkedinService.ScheduleContentPlan(
			r.Context(),
			req.WorkspaceID,
			tenantID,
			req.DraftID,
			parsedTime,
			req.UserTimezone,
			req.MaxDailyBudget,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scheduling failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(item)
	}))

	// 6. List Scheduled Content Calendar Items
	mux.HandleFunc("/api/v1/linkedin/content/calendar", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			http.Error(w, "workspace_id query parameter is required", http.StatusBadRequest)
			return
		}

		filter := linkedin.ContentCalendarFilter{
			WorkspaceID:        workspaceID,
			RequestingTenantID: tenantID,
			Status:             linkedin.ContentPlanStatus(r.URL.Query().Get("status")),
			Format:             linkedin.RepurposedFormat(r.URL.Query().Get("format")),
		}

		items, err := linkedinService.ListContentPlanItems(r.Context(), filter)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list calendar items: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count": len(items),
			"items": items,
		})
	}))

	// 7. Calendar Item Detail, Reschedule (AT-007) and Deletion
	mux.HandleFunc("/api/v1/linkedin/content/calendar/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		switch r.Method {
		case http.MethodGet:
			itemID := r.URL.Query().Get("item_id")
			if itemID == "" {
				http.Error(w, "item_id query parameter is required", http.StatusBadRequest)
				return
			}
			item, err := linkedinService.GetContentPlanItem(r.Context(), itemID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Item not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(item)

		case http.MethodPut:
			var req struct {
				ItemID         string  `json:"item_id"`
				Title          string  `json:"title"`
				NewLocalTime   *string `json:"new_local_time,omitempty"`
				UserTimezone   string  `json:"user_timezone"`
				MaxDailyBudget int     `json:"max_daily_budget"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" {
				http.Error(w, "Invalid payload or missing item_id", http.StatusBadRequest)
				return
			}

			var parsedTime *time.Time
			if req.NewLocalTime != nil && *req.NewLocalTime != "" {
				t, err := time.Parse(time.RFC3339, *req.NewLocalTime)
				if err != nil {
					t, _ = time.Parse("2006-01-02T15:04:05", *req.NewLocalTime)
				}
				parsedTime = &t
			}

			updated, err := linkedinService.UpdateContentPlanItem(
				r.Context(),
				req.ItemID,
				tenantID,
				req.Title,
				parsedTime,
				req.UserTimezone,
				req.MaxDailyBudget,
			)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to update item: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(updated)

		case http.MethodDelete:
			itemID := r.URL.Query().Get("item_id")
			if itemID == "" {
				http.Error(w, "item_id query parameter is required", http.StatusBadRequest)
				return
			}
			if err := linkedinService.DeleteContentPlanItem(r.Context(), itemID, tenantID); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete item: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]bool{"deleted": true})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// 8. Approve Calendar Item (AT-007 HMAC)
	mux.HandleFunc("/api/v1/linkedin/content/calendar/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ItemID    string `json:"item_id"`
			SecretKey string `json:"secret_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" {
			http.Error(w, "Invalid payload or missing item_id", http.StatusBadRequest)
			return
		}

		approved, err := linkedinService.ApproveContentPlanItem(r.Context(), req.ItemID, tenantID, req.SecretKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Approval failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 9. Fail-Closed Autopublish Guard (AT-010 & REQ-015)
	mux.HandleFunc("/api/v1/linkedin/content/calendar/publish", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ItemID string `json:"item_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" {
			http.Error(w, "Invalid payload or missing item_id", http.StatusBadRequest)
			return
		}

		err := linkedinService.AttemptDirectPublish(r.Context(), req.ItemID, tenantID)
		// Guaranteed to fail under AT-010
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error_code": "DIRECT_AUTOPUBLISH_DISALLOWED",
			"error":      err.Error(),
			"allowed_alternatives": []string{
				"clipboard_copy_formatted",
				"compose_deep_link",
			},
		})
	}))

	// 10. 1-Click Clipboard Export & LinkedIn Compose Link (AT-010)
	mux.HandleFunc("/api/v1/linkedin/content/calendar/export", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ItemID    string `json:"item_id"`
			SecretKey string `json:"secret_key"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ItemID == "" {
			http.Error(w, "Invalid payload or missing item_id", http.StatusBadRequest)
			return
		}

		item, md, composeURL, err := linkedinService.ExportPlanForClipboard(r.Context(), req.ItemID, tenantID, req.SecretKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Export failed: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"item":        item,
			"markdown":    md,
			"compose_url": composeURL,
			"exported_at": time.Now().UTC(),
		})
	}))

	// =========================================================================
	// IMP-LI-15: Engagement Monitoring & Analytics REST Endpoints
	// (AT-028, AT-010, AT-012, SRC-L2)
	// =========================================================================

	// 1. Ingest / Record Post Analytics Snapshot (AT-028, AT-010)
	mux.HandleFunc("/api/v1/linkedin/analytics/snapshots", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.Header.Get("X-Workspace-ID")
		if workspaceID == "" {
			workspaceID = "default-workspace"
		}

		if r.Method == http.MethodPost {
			var snapshot linkedin.PostAnalyticsSnapshot
			if err := json.NewDecoder(r.Body).Decode(&snapshot); err != nil {
				http.Error(w, "Invalid request payload", http.StatusBadRequest)
				return
			}
			if snapshot.TenantID == "" {
				snapshot.TenantID = tenantID
			}
			if snapshot.WorkspaceID == "" {
				snapshot.WorkspaceID = workspaceID
			}

			if err := linkedinService.RecordPostAnalytics(r.Context(), &snapshot); err != nil {
				http.Error(w, fmt.Sprintf("Failed to record post analytics: %v", err), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(snapshot)
			return
		}

		if r.Method == http.MethodGet {
			list, err := linkedinService.ListPostAnalytics(r.Context(), workspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list post analytics: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"snapshots": list,
				"total":     len(list),
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 2. Get / Delete Specific Post Analytics Snapshot by ID or Post URN
	mux.HandleFunc("/api/v1/linkedin/analytics/snapshots/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		snapshotID := r.URL.Query().Get("id")
		postURN := r.URL.Query().Get("post_urn")

		if r.Method == http.MethodGet {
			if snapshotID != "" {
				snap, err := linkedinService.GetPostAnalytics(r.Context(), snapshotID, tenantID)
				if err != nil {
					http.Error(w, fmt.Sprintf("Snapshot not found: %v", err), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(snap)
				return
			}
			if postURN != "" {
				snap, err := linkedinService.GetPostAnalyticsByURN(r.Context(), postURN, tenantID)
				if err != nil {
					http.Error(w, fmt.Sprintf("Snapshot not found: %v", err), http.StatusNotFound)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(snap)
				return
			}
			http.Error(w, "Missing id or post_urn query parameter", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodDelete {
			if snapshotID == "" {
				http.Error(w, "Missing id query parameter", http.StatusBadRequest)
				return
			}
			if err := linkedinService.DeletePostAnalytics(r.Context(), snapshotID, tenantID); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete snapshot: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "deleted_id": snapshotID})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 3. Segment Engagers List (SRC-L2 ICP Classification)
	mux.HandleFunc("/api/v1/linkedin/analytics/engagers/segment", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Engagers []linkedin.EngagerProfile `json:"engagers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		segmented, breakdown := linkedinService.SegmentEngagers(r.Context(), req.Engagers)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"engagers":      segmented,
			"icp_breakdown": breakdown,
		})
	}))

	// 4. Curated Creator Benchmarks (AT-028)
	mux.HandleFunc("/api/v1/linkedin/analytics/benchmarks", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		benchmarks := linkedin.CuratedCreatorBenchmarks()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"benchmarks": benchmarks,
			"disclosure": "Benchmarks derived from verified historical user post archives in accordance with AT-028. Not artificial or simulated figures.",
		})
	}))

	// =========================================================================
	// IMP-LI-16: Employee Advocacy & Brand Governance REST Endpoints
	// (LI-16, AT-007, AT-011, SRC-L2)
	// =========================================================================

	// 1. Create & List Advocacy Campaigns
	mux.HandleFunc("/api/v1/linkedin/advocacy/campaigns", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.Header.Get("X-Workspace-ID")
		if workspaceID == "" {
			workspaceID = "default-workspace"
		}

		if r.Method == http.MethodPost {
			var campaign linkedin.AdvocacyCampaign
			if err := json.NewDecoder(r.Body).Decode(&campaign); err != nil {
				http.Error(w, "Invalid request payload", http.StatusBadRequest)
				return
			}
			if campaign.TenantID == "" {
				campaign.TenantID = tenantID
			}
			if campaign.WorkspaceID == "" {
				campaign.WorkspaceID = workspaceID
			}

			created, err := linkedinService.CreateAdvocacyCampaign(r.Context(), &campaign)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create campaign: %v", err), http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(created)
			return
		}

		if r.Method == http.MethodGet {
			list, err := linkedinService.ListAdvocacyCampaigns(r.Context(), workspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list campaigns: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"campaigns": list,
				"total":     len(list),
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 2. Get & Delete Advocacy Campaign Detail
	mux.HandleFunc("/api/v1/linkedin/advocacy/campaigns/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		campaignID := r.URL.Query().Get("id")
		if campaignID == "" {
			http.Error(w, "Missing id query parameter", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			campaign, err := linkedinService.GetAdvocacyCampaign(r.Context(), campaignID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Campaign not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(campaign)
			return
		}

		if r.Method == http.MethodDelete {
			if err := linkedinService.DeleteAdvocacyCampaign(r.Context(), campaignID, tenantID); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete campaign: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "deleted_id": campaignID})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 3. Edit Campaign (Triggers AT-007 Tamper Invalidation)
	mux.HandleFunc("/api/v1/linkedin/advocacy/campaigns/edit", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var campaign linkedin.AdvocacyCampaign
		if err := json.NewDecoder(r.Body).Decode(&campaign); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if campaign.TenantID == "" {
			campaign.TenantID = tenantID
		}

		updated, err := linkedinService.UpdateAdvocacyCampaign(r.Context(), &campaign, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update campaign: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(updated)
	}))

	// 4. Approve Campaign (HMAC Signing AT-007)
	mux.HandleFunc("/api/v1/linkedin/advocacy/campaigns/approve", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			CampaignID string `json:"campaign_id"`
			ApproverID string `json:"approver_id"`
			SecretKey  string `json:"secret_key,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		if req.ApproverID == "" {
			req.ApproverID = "lead-approver"
		}

		approved, err := linkedinService.ApproveAdvocacyCampaign(r.Context(), req.CampaignID, req.ApproverID, tenantID, req.SecretKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to approve campaign: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(approved)
	}))

	// 5. Voluntary Employee Share (AT-010, REQ-015 1-Click Clipboard + Deep Link)
	mux.HandleFunc("/api/v1/linkedin/advocacy/campaigns/share", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.Header.Get("X-Workspace-ID")
		if workspaceID == "" {
			workspaceID = "default-workspace"
		}

		var req struct {
			CampaignID     string `json:"campaign_id"`
			VariantID      string `json:"variant_id"`
			EmployeeID     string `json:"employee_id"`
			CustomizedText string `json:"customized_text,omitempty"`
			SecretKey      string `json:"secret_key,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		if req.EmployeeID == "" {
			req.EmployeeID = "employee-user"
		}

		result, err := linkedinService.ShareAdvocacyContent(
			r.Context(),
			req.CampaignID,
			req.VariantID,
			req.EmployeeID,
			req.CustomizedText,
			workspaceID,
			tenantID,
			req.SecretKey,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to prepare share: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	}))

	// 6. List Employee Share Events for Campaign
	mux.HandleFunc("/api/v1/linkedin/advocacy/campaigns/shares", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		workspaceID := r.Header.Get("X-Workspace-ID")
		campaignID := r.URL.Query().Get("campaign_id")

		shares, err := linkedinService.ListEmployeeShares(r.Context(), campaignID, workspaceID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list shares: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"shares": shares,
			"total":  len(shares),
		})
	}))

	// 7. Anti-Pod / Coordinated Fake Engagement Verification Endpoint (LI-16)
	mux.HandleFunc("/api/v1/linkedin/advocacy/anti-pod/check", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req linkedin.CoordinatedEngagementRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		err := linkedinService.TriggerCoordinatedPod(r.Context(), req)
		if err != nil {
			// Expected to fail closed with 403 Forbidden
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"blocked": true,
				"error":   err.Error(),
				"policy":  "Coordinated engagement pods, mutual auto-liking rings, and automated comments are permanently prohibited by LI-16 policy.",
			})
			return
		}

		http.Error(w, "Unexpected pass of prohibited pod request", http.StatusInternalServerError)
	}))

	// =========================================================================
	// Provider Fallback & Diagnostics Endpoints (IMP-LI-17, LI-17, AT-010, AT-016, SRC-L5)
	// =========================================================================

	// 1. Provider Adapter Registration & Listing
	mux.HandleFunc("/api/v1/linkedin/providers", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		if r.Method == http.MethodPost {
			var provider linkedin.ProviderAdapter
			if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
				http.Error(w, "Invalid payload", http.StatusBadRequest)
				return
			}
			provider.TenantID = tenantID
			if provider.WorkspaceID == "" {
				provider.WorkspaceID = "ws-alpha"
			}
			if err := linkedinService.RegisterProviderAdapter(r.Context(), &provider); err != nil {
				http.Error(w, fmt.Sprintf("Failed to register provider: %v", err), http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(provider)
			return
		}

		if r.Method == http.MethodGet {
			workspaceID := r.URL.Query().Get("workspace_id")
			if workspaceID == "" {
				workspaceID = "ws-alpha"
			}
			providers, err := linkedinService.ListProviderAdapters(r.Context(), workspaceID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to list providers: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"providers": providers,
				"total":     len(providers),
			})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 2. Provider Adapter Detail & Deletion
	mux.HandleFunc("/api/v1/linkedin/providers/detail", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		providerID := r.URL.Query().Get("id")
		if providerID == "" {
			http.Error(w, "Missing id query parameter", http.StatusBadRequest)
			return
		}

		if r.Method == http.MethodGet {
			provider, err := linkedinService.GetProviderAdapter(r.Context(), providerID, tenantID)
			if err != nil {
				http.Error(w, fmt.Sprintf("Provider not found: %v", err), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(provider)
			return
		}

		if r.Method == http.MethodDelete {
			if err := linkedinService.DeleteProviderAdapter(r.Context(), providerID, tenantID); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete provider: %v", err), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "deleted_id": providerID})
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}))

	// 3. Provider Health Update
	mux.HandleFunc("/api/v1/linkedin/providers/health", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			ProviderID   string                        `json:"provider_id"`
			HealthStatus linkedin.ProviderHealthStatus `json:"health_status"`
			LatencyMs    int64                         `json:"latency_ms"`
			ErrorMessage string                        `json:"error_message,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		if err := linkedinService.UpdateProviderHealth(r.Context(), req.ProviderID, tenantID, req.HealthStatus, req.LatencyMs, req.ErrorMessage); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update health: %v", err), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "updated_id": req.ProviderID})
	}))

	// 4. Diagnostic Doctor Probe Runner (SRC-L5 doctor.py)
	mux.HandleFunc("/api/v1/linkedin/providers/doctor", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			WorkspaceID string `json:"workspace_id"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.WorkspaceID == "" {
			req.WorkspaceID = "ws-alpha"
		}

		report, err := linkedinService.RunDiagnosticDoctor(r.Context(), req.WorkspaceID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Diagnostic doctor probe failed: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// 5. Provider Dispatch With Safe Fallback (AT-010, SRC-L5)
	mux.HandleFunc("/api/v1/linkedin/providers/dispatch", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req linkedin.DispatchActionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		req.TenantID = tenantID
		if req.WorkspaceID == "" {
			req.WorkspaceID = "ws-alpha"
		}

		res, err := linkedinService.DispatchWithFallback(r.Context(), req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if errors.Is(err, linkedin.ErrActionUnsupportedOrProhibited) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusServiceUnavailable)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
				"policy":  "Fallback operates strictly among independently permitted paths without simulation.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(res)
	}))

	// 6. Credential Scrubber Test Endpoint (AT-016)
	mux.HandleFunc("/api/v1/linkedin/providers/scrub-test", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			RawText string `json:"raw_text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}

		scrubbed := linkedinService.ScrubDiagnosticText(req.RawText)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"raw_length":      len(req.RawText),
			"scrubbed_text":   scrubbed,
			"scrubbed_length": len(scrubbed),
		})
	}))

	// =========================================================================
	// Platform Limits, CAPTCHA / Reauth & Safety Gatekeeper Endpoints
	// (IMP-LI-18, LI-18, REQ-009, REQ-010, AT-008, AT-009, AT-010, FND-011, FND-013, FND-014, SRC-L1, SRC-S3)
	// =========================================================================

	// 1. Get Live Safety State & Circuit Breaker
	mux.HandleFunc("/api/v1/linkedin/limits/safety-state", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			accountID = "li:member_default"
		}

		state, err := linkedinService.GetAccountSafetyState(r.Context(), accountID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get safety state: %v", err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(state)
	}))

	// 2. Detect & Record Restriction / Upstream Error
	mux.HandleFunc("/api/v1/linkedin/limits/detect-restriction", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			AccountID       string `json:"account_id"`
			WorkspaceID     string `json:"workspace_id"`
			ActionAttempted string `json:"action_attempted"`
			StatusCode      int    `json:"status_code"`
			RawError        string `json:"raw_error"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if req.AccountID == "" {
			req.AccountID = "li:member_default"
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = "ws-alpha"
		}

		incident, state, err := linkedinService.DetectAndRecordRestriction(
			r.Context(),
			req.AccountID,
			tenantID,
			req.WorkspaceID,
			req.ActionAttempted,
			req.StatusCode,
			req.RawError,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Restriction analysis error: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"incident": incident,
			"state":    state,
		})
	}))

	// 3. Resolve Challenge via Legitimate User Re-authentication (AT-010)
	mux.HandleFunc("/api/v1/linkedin/limits/resolve-challenge", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			AccountID        string `json:"account_id"`
			ResolutionMethod string `json:"resolution_method"`
			VerifiedByUser   bool   `json:"verified_by_user"`
			Notes            string `json:"notes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if req.AccountID == "" {
			req.AccountID = "li:member_default"
		}

		resolvedState, err := linkedinService.ResolveChallenge(
			r.Context(),
			req.AccountID,
			tenantID,
			linkedin.ResolveChallengeRequest{
				ResolutionMethod: req.ResolutionMethod,
				VerifiedByUser:   req.VerifiedByUser,
				Notes:            req.Notes,
			},
		)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  err.Error(),
				"policy": "Anti-bot evasion or unverified challenge solve is strictly forbidden (AT-010).",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resolvedState)
	}))

	// 4. Activity Log & Incident Feed (FND-013)
	mux.HandleFunc("/api/v1/linkedin/limits/activity-log", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			accountID = "li:member_default"
		}

		limit := 50
		activities, err := linkedinService.ListAccountActivity(r.Context(), accountID, tenantID, limit)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list activity: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"account_id": accountID,
			"total":      len(activities),
			"activities": activities,
		})
	}))

	// 5. Cross-Workspace Combined Account Budget Report (AT-008, AT-009, REQ-010)
	mux.HandleFunc("/api/v1/linkedin/limits/combined-budget", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			accountID = "li:member_default"
		}

		report, err := linkedinService.GetCombinedBudgetReport(r.Context(), accountID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to compute combined budget: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(report)
	}))

	// 6. Pre-flight Safety Gate Check
	mux.HandleFunc("/api/v1/linkedin/limits/check-gate", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			AccountID string `json:"account_id"`
			Action    string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		if req.AccountID == "" {
			req.AccountID = "li:member_default"
		}
		if req.Action == "" {
			req.Action = "connection_request"
		}

		decision, err := linkedinService.CheckSafetyGate(r.Context(), req.AccountID, tenantID, req.Action)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check safety gate: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(decision)
	}))

	// =========================================================================
	// Session Lifecycle & Optional Local Tools Endpoints (IMP-LI-19, LI-19, AT-016, FND-009, SRC-L3)
	// =========================================================================

	// 1. Create Isolated Session (Per-owner isolated storage, zero plaintext cookie paste)
	mux.HandleFunc("/api/v1/linkedin/sessions/create", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req linkedin.CreateSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid payload", http.StatusBadRequest)
			return
		}
		req.TenantID = tenantID
		if req.OwnerID == "" {
			req.OwnerID = "user-default"
		}

		session, err := linkedinService.CreateSession(r.Context(), req)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			if errors.Is(err, linkedin.ErrPlaintextCredentialProhibited) {
				w.WriteHeader(http.StatusForbidden)
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error":  err.Error(),
				"policy": "Plaintext credential and raw cookie paste is permanently prohibited (AT-016, REQ-021).",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(session)
	}))

	// 2. List Sessions for Owner
	mux.HandleFunc("/api/v1/linkedin/sessions", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}
		ownerID := r.URL.Query().Get("owner_id")

		sessions, err := linkedinService.ListSessions(r.Context(), ownerID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list sessions: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"tenant_id": tenantID,
			"total":     len(sessions),
			"sessions":  sessions,
		})
	}))

	// 3. Revoke Session (FND-009)
	mux.HandleFunc("/api/v1/linkedin/sessions/revoke", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			SessionID string `json:"session_id"`
			Reason    string `json:"reason"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
			http.Error(w, "Invalid payload or missing session_id", http.StatusBadRequest)
			return
		}
		if req.Reason == "" {
			req.Reason = "Explicit user revocation"
		}

		session, err := linkedinService.RevokeSession(r.Context(), req.SessionID, tenantID, req.Reason)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to revoke session: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(session)
	}))

	// 4. Acquire Short-Lived Viewer Lease (SRC-L3 profile_lease.py, daemon_lock.py)
	mux.HandleFunc("/api/v1/linkedin/sessions/lease/acquire", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			SessionID               string `json:"session_id"`
			Purpose                 string `json:"purpose"`
			DurationSeconds         int    `json:"duration_seconds"`
			RequiresUserSupervision bool   `json:"requires_user_supervision"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SessionID == "" {
			http.Error(w, "Invalid payload or missing session_id", http.StatusBadRequest)
			return
		}

		lease, err := linkedinService.AcquireViewerLease(r.Context(), req.SessionID, tenantID, linkedin.AcquireViewerLeaseRequest{
			SessionID:               req.SessionID,
			Purpose:                 req.Purpose,
			DurationSeconds:         req.DurationSeconds,
			RequiresUserSupervision: req.RequiresUserSupervision,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(lease)
	}))

	// 5. Release Viewer Lease
	mux.HandleFunc("/api/v1/linkedin/sessions/lease/release", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			LeaseID string `json:"lease_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LeaseID == "" {
			http.Error(w, "Invalid payload or missing lease_id", http.StatusBadRequest)
			return
		}

		if err := linkedinService.ReleaseViewerLease(r.Context(), req.LeaseID, tenantID); err != nil {
			http.Error(w, fmt.Sprintf("Failed to release lease: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"success":    true,
			"released_id": req.LeaseID,
		})
	}))

	// =========================================================================
	// LinkedIn Relationship CSV & Google Sheets Exports (IMP-LI-20, LI-20, REQ-006, REQ-019, AT-013, AT-014)
	// =========================================================================

	// 1. Export Scoped Recruiter Relationships to RFC4180 CSV
	mux.HandleFunc("/api/v1/linkedin/export/relationships/csv", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			WorkspaceID          string   `json:"workspace_id"`
			StatusFilter         string   `json:"status_filter,omitempty"`
			CompanyFilter        string   `json:"company_filter,omitempty"`
			IncludeNotes         bool     `json:"include_notes"`
			IncludePersonRecords bool     `json:"include_person_records"`
			SelectedColumns      []string `json:"selected_columns,omitempty"`
			WithBOM              bool     `json:"with_bom"`
			Timezone             string   `json:"timezone,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = "ws-alpha"
		}

		filter := linkedin.RelationshipExportFilter{
			WorkspaceID:          req.WorkspaceID,
			TenantID:             tenantID,
			RequestingTenantID:   tenantID,
			RequestingOwnerID:    user.ID,
			Destination:          linkedin.DestinationCSVDownload,
			StatusFilter:         linkedin.LeadStatus(req.StatusFilter),
			CompanyFilter:        req.CompanyFilter,
			IncludeNotes:         req.IncludeNotes,
			IncludePersonRecords: req.IncludePersonRecords,
			SelectedColumns:      req.SelectedColumns,
			WithBOM:              req.WithBOM,
			Timezone:             req.Timezone,
		}

		manifest, err := linkedinService.ExportRelationshipsCSV(r.Context(), filter)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to export relationship CSV: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(manifest)
	}))

	// 2. Reconcile / Idempotently Sync Recruiter Relationships into Google Sheets
	mux.HandleFunc("/api/v1/linkedin/export/relationships/sheets-sync", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, _ := identity.GetUserFromContext(r.Context())
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		var req struct {
			WorkspaceID   string     `json:"workspace_id"`
			SpreadsheetID string     `json:"spreadsheet_id"`
			SheetName     string     `json:"sheet_name"`
			Timezone      string     `json:"timezone,omitempty"`
			IncludeNotes  bool       `json:"include_notes"`
			ExistingRows  [][]string `json:"existing_rows,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.WorkspaceID == "" {
			req.WorkspaceID = "ws-alpha"
		}
		if req.SpreadsheetID == "" || req.SheetName == "" {
			http.Error(w, "spreadsheet_id and sheet_name are required", http.StatusBadRequest)
			return
		}

		config := linkedin.RelationshipSheetsSyncConfig{
			ConfigID:      fmt.Sprintf("cfg_sync_%d", time.Now().UnixNano()),
			WorkspaceID:   req.WorkspaceID,
			TenantID:      tenantID,
			SpreadsheetID: req.SpreadsheetID,
			SheetName:     req.SheetName,
			Timezone:      req.Timezone,
			IncludeNotes:  req.IncludeNotes,
			SyncMode:      "one_way_upsert",
			CreatedAt:     time.Now().UTC(),
		}

		reconciledRows, result, err := linkedinService.SyncRelationshipsToSheets(r.Context(), req.ExistingRows, config, tenantID, user.ID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to sync relationships to Google Sheets: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"result":          result,
			"reconciled_rows": reconciledRows,
		})
	}))

	// 3. List Relationship Export and Sync Audit Records
	mux.HandleFunc("/api/v1/linkedin/export/relationships/history", identity.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tenantID := r.Header.Get("X-Tenant-ID")
		if tenantID == "" {
			tenantID = "default-tenant"
		}

		workspaceID := r.URL.Query().Get("workspace_id")
		if workspaceID == "" {
			workspaceID = "ws-alpha"
		}

		audits, err := linkedinService.ListRelationshipExportAudits(r.Context(), workspaceID, tenantID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to list relationship export history: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"audits": audits,
			"total":  len(audits),
		})
	}))

	handler := identity.AuthMiddleware(authService)(mux)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("[core-api] Starting Go API server on port %s...", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[core-api] Server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("[core-api] Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[core-api] Shutdown error: %v", err)
	}
	log.Println("[core-api] Server stopped cleanly.")
}