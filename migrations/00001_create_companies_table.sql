-- +goose Up
-- +goose StatementBegin
create table companies (
  id         bigserial    primary key,
  created_at timestamp(0) not null default now(),
  updated_at timestamp(0) not null default now(),
  name       text         unique not null,
  url        text         not null default '',
  tech_stack text[]       not null,
  version    bigint       not null default 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table companies;
-- +goose StatementEnd
