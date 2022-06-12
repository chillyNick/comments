-- +goose Up
CREATE TYPE comment_status AS ENUM ('new', 'under_moderation', 'moderation_failed', 'moderation_passed');

CREATE TABLE comment (
     id bigserial,
     user_id int not null,
     item_id int not null,
     comment varchar(512) not null,
     status_id comment_status
);

-- +goose Down
DROP TABLE comment;