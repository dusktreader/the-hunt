-- +goose Up
-- +goose StatementBegin
create table tokens (
    hash       bytea                    primary key,
    user_id    bigint                   references users(id) on delete cascade,
    expires_at timestamp with time zone not null,
    scope      text                     not null,
    is_admin   boolean                  not null default false
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table tokens;
-- +goose StatementEnd
