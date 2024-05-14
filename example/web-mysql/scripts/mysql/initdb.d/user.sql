create table user
(
    id         int primary key auto_increment,
    name       varchar(64),
    email      varchar(255),
    created_at datetime default current_timestamp,
    updated_at datetime default current_timestamp on update current_timestamp,
    deleted_at datetime
);