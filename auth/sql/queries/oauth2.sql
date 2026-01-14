-- name: CreateOAuth2Client :one
INSERT INTO public.oauth2_client (id, name, secret, redirect_uris, scopes, grant_types, access_token_ttl, refresh_token_ttl, enabled, created, updated)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, name, secret, redirect_uris, scopes, grant_types, access_token_ttl, refresh_token_ttl, enabled, created, updated;


-- name: GetOAuth2Client :one
SELECT id, name, secret, redirect_uris, scopes, grant_types, access_token_ttl, refresh_token_ttl, enabled, created, updated
FROM public.oauth2_client
WHERE id = $1;

-- name: DisableOAuth2Client :exec
UPDATE public.oauth2_client set enable = false
WHERE id = $1;

-- name: ListOAuth2Client :many
SELECT id, name, secret, redirect_uris, scopes, grant_types, access_token_ttl, refresh_token_ttl, enabled, created, updated
FROM public.oauth2_client;

-- name: CreateOAuth2AuthorizationCode :one
INSERT INTO public.oauth2_authorization_code (code, client_id, user_id, redirect_uri, scope, state, nonce, expires, created)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
RETURNING code, client_id, user_id, redirect_uri, scope, state, nonce, expires, created;

-- name: GetOAuth2AuthorizationCode :one
SELECT code, client_id, user_id, redirect_uri, scope, state, nonce, expires, created
FROM public.oauth2_authorization_code
WHERE code = $1 AND expires > NOW();

-- name: DeleteOAuth2AuthorizationCode :exec
DELETE FROM public.oauth2_authorization_code WHERE code = $1;

-- name: CleanupExpiredAuthorizationCodes :exec
DELETE FROM public.oauth2_authorization_code WHERE expires <= NOW();

-- name: CreateOAuth2Token :one
INSERT INTO public.oauth2_token (access_token, token_type, client_id, user_id, scope, expires, created)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
RETURNING access_token, token_type, client_id, user_id, scope, expires, created;

-- name: GetOAuth2Token :one
SELECT access_token, token_type, client_id, user_id, scope, expires, created
FROM public.oauth2_token
WHERE access_token = $1 AND expires > NOW();

-- name: GetOAuth2TokensByUser :many
SELECT access_token, token_type, client_id, user_id, scope, expires, created
FROM public.oauth2_token
WHERE user_id = $1 AND expires > NOW()
ORDER BY created DESC;

-- name: RevokeOAuth2Token :exec
DELETE FROM public.oauth2_token WHERE access_token = $1;

-- name: RevokeOAuth2TokensByUser :exec
DELETE FROM public.oauth2_token WHERE user_id = $1;

-- name: CleanupExpiredTokens :exec
DELETE FROM public.oauth2_token WHERE expires <= NOW();

-- name: CreateOAuth2RefreshToken :one
INSERT INTO public.oauth2_refresh_token (refresh_token, client_id, user_id, scope, expires, created)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING refresh_token, client_id, user_id, scope, expires, created;

-- name: GetOAuth2RefreshToken :one
SELECT refresh_token, client_id, user_id, scope, expires, created
FROM public.oauth2_refresh_token
WHERE refresh_token = $1 AND expires > NOW();

-- name: GetOAuth2RefreshTokensByUser :many
SELECT refresh_token, client_id, user_id, scope, expires, created
FROM public.oauth2_refresh_token
WHERE user_id = $1 AND expires > NOW()
ORDER BY created DESC;

-- name: RevokeOAuth2RefreshToken :exec
DELETE FROM public.oauth2_refresh_token WHERE refresh_token = $1;

-- name: RevokeOAuth2RefreshTokensByUser :exec
DELETE FROM public.oauth2_refresh_token WHERE user_id = $1;

-- name: CleanupExpiredRefreshTokens :exec
DELETE FROM public.oauth2_refresh_token WHERE expires <= NOW();

-- name: GetOAuth2ClientStats :one
SELECT 
    c.id,
    c.name,
    COUNT(DISTINCT ac.code) as auth_codes_count,
    COUNT(DISTINCT t.access_token) as active_tokens_count,
    COUNT(DISTINCT rt.refresh_token) as active_refresh_tokens_count
FROM public.oauth2_client c
LEFT JOIN public.oauth2_authorization_code ac ON c.id = ac.client_id
LEFT JOIN public.oauth2_token t ON c.id = t.client_id AND t.expires > NOW()
LEFT JOIN public.oauth2_refresh_token rt ON c.id = rt.client_id AND rt.expires > NOW()
WHERE c.id = $1
GROUP BY c.id, c.name;

-- name: GetOIDCJwk :one
SELECT id, kid, kty, use, alg, n, e, created, updated
FROM public.oidc_jwk
WHERE kid = $1;

-- name: ListOIDCJwks :many
SELECT id, kid, kty, use, alg, n, e, created, updated
FROM public.oidc_jwk
ORDER BY created DESC;

-- name: GetActiveOIDCJwks :many
SELECT id, kid, kty, use, alg, n, e, created, updated
FROM public.oidc_jwk
ORDER BY created DESC;
