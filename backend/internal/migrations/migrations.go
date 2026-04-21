package migrations

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Up(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		create extension if not exists pgcrypto;

		create table if not exists companies (
			id uuid primary key default gen_random_uuid(),
			normalized_linkedin_url text not null,
			normalized_domain text null,
			company_name text not null,
			attio_record_id text not null,
			attio_web_url text not null,
			raw_extracted_json jsonb not null,
			first_synced_at timestamptz not null,
			last_synced_at timestamptz not null,
			last_sync_status text not null,
			last_sync_error text null
		);

		create unique index if not exists companies_normalized_linkedin_url_idx on companies (normalized_linkedin_url);
		create unique index if not exists companies_normalized_domain_idx on companies (normalized_domain) where normalized_domain is not null;
	`)
	return err
}
