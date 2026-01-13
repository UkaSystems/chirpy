-- +goose Up
Create Table users (
    id UUID default gen_random_uuid() Primary Key,
    created_at Timestamp Not Null,
    updated_at Timestamp Not Null,
    email Varchar(255) Not Null Unique
);

-- +goose Down
Drop Table users;