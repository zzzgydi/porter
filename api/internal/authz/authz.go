package authz

import (
	"net/http"

	"github.com/zzzgydi/porter/api/internal/httpx"
	"github.com/zzzgydi/porter/api/internal/members"
	"github.com/zzzgydi/porter/api/internal/session"
)

func RequireMember(membersRepo *members.Repo, sessionMgr *session.Manager, r *http.Request, projectID string) (*session.Claims, error) {
	claims, err := session.FromRequest(sessionMgr, r)
	if err != nil {
		return nil, err
	}
	if claims.Role == "platform_admin" {
		return claims, nil
	}
	_, err = membersRepo.GetRole(r.Context(), projectID, claims.UserID)
	if err != nil {
		return nil, httpx.Forbidden("not a project member")
	}
	return claims, nil
}

func RequireRole(membersRepo *members.Repo, sessionMgr *session.Manager, r *http.Request, projectID string, allowed ...string) (*session.Claims, error) {
	claims, err := session.FromRequest(sessionMgr, r)
	if err != nil {
		return nil, err
	}
	if claims.Role == "platform_admin" {
		return claims, nil
	}
	role, err := membersRepo.GetRole(r.Context(), projectID, claims.UserID)
	if err != nil {
		return nil, httpx.Forbidden("not a project member")
	}
	for _, a := range allowed {
		if role == a {
			return claims, nil
		}
	}
	return nil, httpx.Forbidden("insufficient permissions")
}
