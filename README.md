# Gogo

Gogo is a Go backend framework planned around Django-style applications, models, migrations, admin, auth, APIs, and Celery-style queues.

## Current Status

Implementation has started from the incremental plans in `.plans/`.

## Planned CLI

- `gogo help`
- `gogo version`
- `gogo check`
- `gogo runserver`
- `gogo startproject`
- `gogo startapp`
- `gogo makemigrations`
- `gogo migrate`
- `gogo showmigrations`
- `gogo sqlmigrate`
- `gogo squashmigrations`
- `gogo createsuperuser`
- `gogo changepassword`
- `gogo collectstatic`
- `gogo shell`
- `gogo dbshell`
- `gogo test`
- `gogo worker`
- `gogo beat`
- `gogo inspect`
- `gogo queues`
- `gogo dumpdata`
- `gogo loaddata`

## Environment

Copy `.env.example` to `.env` for local development. Keep `.env` out of Git.

Required variables:

- `GOGO_SECRET_KEY`
- `DATABASE_URL`

Defaults:

- `GOGO_ENV=development`
- `GOGO_HTTP_ADDR=:8000`
- `GOGO_ALLOWED_HOSTS=localhost,127.0.0.1`

Queue variables are required only when queue workers are enabled:

- `GOGO_BROKER_URL`
- `GOGO_RESULT_BACKEND`

## Security

Gogo is intended for production-grade public products. Secrets, local databases, generated uploads, and machine-local files must not be committed.
