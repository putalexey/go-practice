-- +goose Up
-- +goose StatementBegin
create unique index shorts_original_idx ON shorts (original);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop index shorts_original_idx;
-- +goose StatementEnd
