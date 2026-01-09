-- name: UserById :one
SELECT *
FROM public.user
WHERE id = $1;

-- name: UserByPhone :one
SELECT *
FROM public.user
WHERE phone = $1;

