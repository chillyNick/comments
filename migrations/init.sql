CREATE TABLE status (
    id int PRIMARY KEY,
    name varchar(32)
);

INSERT INTO status(id, name)
VALUES
    (1, 'new'),
    (2, 'under_moderation'),
    (3, 'moderation_failed'),
    (4, 'moderation_passed')
;


CREATE TABLE comment (
     id bigserial PRIMARY KEY,
     user_id int not null,
     item_id int not null,
     comment varchar(512) not null,
     status_id int not null REFERENCES status(id)
);