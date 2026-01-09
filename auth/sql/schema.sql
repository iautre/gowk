create table public."user"
(
    id       bigserial
        constraint user_pk
            primary key,
    phone    varchar,
    email    varchar,
    nickname varchar,
    "group"  varchar,
    status   integer,
    created  timestamp with time zone,
    updated  timestamp with time zone,
    aid      varchar,
    secret   varchar
);

alter table public."user"
    owner to postgres;

create table public.app
(
    id          bigint default nextval('app_id_seq'::regclass) not null
        constraint app_pk
            primary key,
    name        varchar,
    url         varchar,
    type        varchar,
    auth_ignore boolean,
    auth_key    varchar,
    auth_secret varchar,
    key         varchar
);

alter table public.app
    owner to postgres;

create table public.app_data
(
    id     bigint,
    module varchar,
    data   jsonb,
    app_id integer
);

alter table public.app_data
    owner to postgres;

-- OAuth2 Clients Table
create table public.oauth2_client
(
    id varchar primary key,
    name varchar not null,
    secret varchar not null,
    redirect_uris text not null, -- JSON array of redirect URIs
    scopes text not null, -- JSON array of allowed scopes
    grant_types text not null, -- JSON array of allowed grant types
    access_token_ttl bigint not null default 3600, -- Access token TTL in seconds (default: 1 hour)
    refresh_token_ttl bigint not null default 2592000, -- Refresh token TTL in seconds (default: 30 days)
    created timestamp with time zone not null default now(),
    updated timestamp with time zone not null default now()
);

alter table public.oauth2_client
    owner to postgres;

-- OAuth2 Authorization Codes Table
create table public.oauth2_authorization_code
(
    code varchar primary key,
    client_id varchar not null,
    user_id bigint not null,
    redirect_uri text,
    scope text,
    state text,
    nonce text,
    expires timestamp with time zone not null,
    created timestamp with time zone not null default now(),
    constraint oauth2_authorization_code_client_id_fkey
        foreign key (client_id) references public.oauth2_client (id) on delete cascade
);

alter table public.oauth2_authorization_code
    owner to postgres;

-- OAuth2 Access Tokens Table
create table public.oauth2_token
(
    access_token varchar primary key,
    token_type text not null default 'Bearer',
    client_id varchar not null,
    user_id bigint not null,
    scope text,
    expires timestamp with time zone not null,
    created timestamp with time zone not null default now(),
    constraint oauth2_token_client_id_fkey
        foreign key (client_id) references public.oauth2_client (id) on delete cascade
);

alter table public.oauth2_token
    owner to postgres;

-- OAuth2 Refresh Tokens Table
create table public.oauth2_refresh_token
(
    refresh_token varchar primary key,
    client_id varchar not null,
    user_id bigint not null,
    scope text,
    expires timestamp with time zone not null,
    created timestamp with time zone not null default now(),
    constraint oauth2_refresh_token_client_id_fkey
        foreign key (client_id) references public.oauth2_client (id) on delete cascade
);

alter table public.oauth2_refresh_token
    owner to postgres;

-- Indexes for better performance
create index if not exists idx_oauth2_authorization_code_client_id on public.oauth2_authorization_code (client_id);
create index if not exists idx_oauth2_authorization_code_expires on public.oauth2_authorization_code (expires);
create index if not exists idx_oauth2_token_client_id on public.oauth2_token (client_id);
create index if not exists idx_oauth2_token_expires on public.oauth2_token (expires);
create index if not exists idx_oauth2_refresh_token_client_id on public.oauth2_refresh_token (client_id);
create index if not exists idx_oauth2_refresh_token_expires on public.oauth2_refresh_token (expires);

-- OIDC JWK Keys Table
create table public.oidc_jwk
(
    id varchar primary key,
    kid varchar not null unique, -- Key ID
    kty varchar not null, -- Key Type (e.g., "RSA")
    use varchar not null, -- Public Key Use (e.g., "sig")
    alg varchar not null, -- Algorithm (e.g., "RS256")
    n text not null, -- Modulus for RSA keys
    e text not null, -- Exponent for RSA keys
    created timestamp with time zone not null default now(),
    updated timestamp with time zone not null default now()
);

alter table public.oidc_jwk
    owner to postgres;

-- Index for JWK performance
create index if not exists idx_oidc_jwk_kid on public.oidc_jwk (kid);

