-- name: AppByKey :one
SELECT *
FROM public.app
WHERE key = $1;
