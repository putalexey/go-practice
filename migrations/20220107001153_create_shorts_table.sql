-- +goose Up
-- +goose StatementBegin
create table if not exists shorts (
  short varchar(255) primary key,
  original varchar(2048),
  user_id varchar(255)
);
create index if not exists shorts_user_id_idx ON shorts (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists shorts
-- +goose StatementEnd
