package authz

import (
	"net/http"

	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/session"
)

// RequireSession parses the console session or returns a 401 error.
func RequireSession(sessionMgr *session.Manager, r *http.Request) (*session.Claims, error) {
	claims, err := session.FromRequest(sessionMgr, r)
	if err != nil {
		return nil, httpx.Unauthorized("session required")
	}
	return claims, nil
}

func RequireMember(membersRepo *members.Repo, sessionMgr *session.Manager, r *http.Request, projectID string) (*session.Claims, error) {
	claims, err := RequireSession(sessionMgr, r)
	if err != nil {
		return nil, err
	}
	if claims.Role == "platform_admin" {
		return claims, nil
	}
	_, err = membersRepo.GetRole(r.Context(), projectID, claims.UserID)
	if err != nil {
		if members.IsNotMember(err) {
			return nil, httpx.Forbidden("not a project member")
		}
		return nil, httpx.Internal("db error")
	}
	return claims, nil
}

func RequireRole(membersRepo *members.Repo, sessionMgr *session.Manager, r *http.Request, projectID string, allowed ...string) (*session.Claims, error) {
	claims, err := RequireSession(sessionMgr, r)
	if err != nil {
		return nil, err
	}
	if claims.Role == "platform_admin" {
		return claims, nil
	}
	role, err := membersRepo.GetRole(r.Context(), projectID, claims.UserID)
	if err != nil {
		if members.IsNotMember(err) {
			return nil, httpx.Forbidden("not a project member")
		}
		return nil, httpx.Internal("db error")
	}
	for _, a := range allowed {
		if role == a {
			return claims, nil
		}
	}
	return nil, httpx.Forbidden("insufficient permissions")
}
