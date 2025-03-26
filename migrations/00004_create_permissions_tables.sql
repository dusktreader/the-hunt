-- +goose Up
-- +goose StatementBegin
create table permissions (
    id   bigserial primary key,
    code text      not null
);

insert into permissions (code) values
    ('companies:read'),
    ('companies:write'),
    ('users:read'),
    ('users:write')
;

create table user_permissions (
    user_id       bigint not null references users(id) on delete cascade,
    permission_id bigint not null references permissions(id) on delete cascade,

    primary key (user_id, permission_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table user_permissions;
drop table permissions;
-- +goose StatementEnd
