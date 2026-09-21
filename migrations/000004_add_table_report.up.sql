CREATE EXTENSION IF NOT EXISTS pg_trgm;


create table if not exists report
(
    id           text primary key,
    name        text                     not null,
    path text                    not null,
    extension         text,
    folder_id         text,
    folder_name text,
    parent_folder_id         text,
    size         text,
    size_bytes         BIGINT,
    modified_at   timestamp with time zone not null,
    FOREIGN KEY (folder_id) REFERENCES folder (id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folder (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);
