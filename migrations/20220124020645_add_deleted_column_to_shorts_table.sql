-- +goose Up
-- +goose StatementBegin
alter table shorts add column deleted bool NULL DEFAULT false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
alter table shorts drop column deleted;
-- +goose StatementEnd
