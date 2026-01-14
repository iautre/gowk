-- Database Schema for SQLC
-- This file contains the complete database structure for SQLC code generation
-- Includes tables, indexes, and constraints

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- User Table
CREATE TABLE IF NOT EXISTS public."user"
(
    id       bigserial
        CONSTRAINT user_pk
            PRIMARY KEY,
    phone    varchar,
    email    varchar,
    nickname varchar,
    "group"  varchar,
    enabled  boolean NOT NULL DEFAULT true, -- true = active, false = disabled
    created  timestamp with time zone NOT NULL DEFAULT now(),
    updated  timestamp with time zone NOT NULL DEFAULT now(),
    aid      varchar,
    secret   varchar,
    last_login_at timestamp with time zone, -- Last successful login timestamp
    login_count integer DEFAULT 0 CONSTRAINT chk_user_login_count CHECK (login_count >= 0), -- Number of login attempts
    avatar    varchar, -- User avatar URL
    is_verified boolean DEFAULT false CONSTRAINT chk_user_is_verified CHECK (is_verified IN (true, false)) -- Phone/email verification status
);

-- OAuth2 Clients Table
CREATE TABLE IF NOT EXISTS public.oauth2_client
(
    id varchar PRIMARY KEY,
    name varchar NOT NULL,
    secret varchar NOT NULL,
    redirect_uris text NOT NULL, -- JSON array of redirect URIs
    scopes text NOT NULL, -- JSON array of allowed scopes
    grant_types text NOT NULL, -- JSON array of allowed grant types
    access_token_ttl bigint NOT NULL DEFAULT 3600, -- Access token TTL in seconds (default: 1 hour)
    refresh_token_ttl bigint NOT NULL DEFAULT 2592000, -- Refresh token TTL in seconds (default: 30 days)
    enabled boolean NOT NULL DEFAULT true, -- true = active, false = disabled
    created timestamp with time zone NOT NULL DEFAULT now(),
    updated timestamp with time zone NOT NULL DEFAULT now()
);

-- OAuth2 Authorization Codes Table
CREATE TABLE IF NOT EXISTS public.oauth2_authorization_code
(
    code varchar PRIMARY KEY,
    client_id varchar NOT NULL,
    user_id bigint NOT NULL,
    redirect_uri text,
    scope text,
    state text,
    nonce text,
    expires timestamp with time zone NOT NULL,
    created timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT oauth2_authorization_code_client_id_fkey
        FOREIGN KEY (client_id) REFERENCES public.oauth2_client (id) ON DELETE CASCADE
);

-- OAuth2 Access Tokens Table
CREATE TABLE IF NOT EXISTS public.oauth2_token
(
    access_token varchar PRIMARY KEY,
    token_type text NOT NULL DEFAULT 'Bearer',
    client_id varchar NOT NULL,
    user_id bigint NOT NULL,
    scope text,
    expires timestamp with time zone NOT NULL,
    created timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT oauth2_token_client_id_fkey
        FOREIGN KEY (client_id) REFERENCES public.oauth2_client (id) ON DELETE CASCADE
);

-- OAuth2 Refresh Tokens Table
CREATE TABLE IF NOT EXISTS public.oauth2_refresh_token
(
    refresh_token varchar PRIMARY KEY,
    client_id varchar NOT NULL,
    user_id bigint NOT NULL,
    scope text,
    expires timestamp with time zone NOT NULL,
    created timestamp with time zone NOT NULL DEFAULT now(),
    CONSTRAINT oauth2_refresh_token_client_id_fkey
        FOREIGN KEY (client_id) REFERENCES public.oauth2_client (id) ON DELETE CASCADE
);

-- OIDC JWK Keys Table
CREATE TABLE IF NOT EXISTS public.oidc_jwk
(
    id varchar PRIMARY KEY,
    kid varchar NOT NULL UNIQUE, -- Key ID
    kty varchar NOT NULL, -- Key Type (e.g., "RSA")
    use varchar NOT NULL, -- Public Key Use (e.g., "sig")
    alg varchar NOT NULL, -- Algorithm (e.g., "RS256")
    n text NOT NULL, -- Modulus for RSA keys
    e text NOT NULL, -- Exponent for RSA keys
    created timestamp with time zone NOT NULL DEFAULT now(),
    updated timestamp with time zone NOT NULL DEFAULT now()
);

-- Indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_status ON public."user" (status);
CREATE INDEX IF NOT EXISTS idx_user_phone ON public."user" (phone);
CREATE INDEX IF NOT EXISTS idx_user_email ON public."user" (email) WHERE email IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_oauth2_authorization_code_client_id ON public.oauth2_authorization_code (client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_authorization_code_expires ON public.oauth2_authorization_code (expires);
CREATE INDEX IF NOT EXISTS idx_oauth2_token_client_id ON public.oauth2_token (client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_token_expires ON public.oauth2_token (expires);
CREATE INDEX IF NOT EXISTS idx_oauth2_refresh_token_client_id ON public.oauth2_refresh_token (client_id);
CREATE INDEX IF NOT EXISTS idx_oauth2_refresh_token_expires ON public.oauth2_refresh_token (expires);

CREATE INDEX IF NOT EXISTS idx_oidc_jwk_kid ON public.oidc_jwk (kid);

-- Set table permissions
ALTER TABLE public."user" OWNER TO postgres;
ALTER TABLE public.oauth2_client OWNER TO postgres;
ALTER TABLE public.oauth2_authorization_code OWNER TO postgres;
ALTER TABLE public.oauth2_token OWNER TO postgres;
ALTER TABLE public.oauth2_refresh_token OWNER TO postgres;
ALTER TABLE public.oidc_jwk OWNER TO postgres;
