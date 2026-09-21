CREATE EXTENSION IF NOT EXISTS pg_trgm;


create table if not exists file
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

CREATE INDEX ix_folder_parent_name
    ON folder (parent_id, name, id);

CREATE INDEX ix_folder_root_parent
    ON folder (root_folder_id, parent_id);

CREATE INDEX ix_video_root_parent
    ON video (folder_id, parent_folder_id);

CREATE INDEX ix_file_root_parent
    ON file (folder_id, parent_folder_id);

CREATE INDEX ix_video_parent_folder
    ON video (parent_folder_id);

CREATE INDEX ix_file_parent_folder
    ON file (parent_folder_id);


CREATE INDEX ix_video_title_trgm
    ON video USING gin (lower(title) gin_trgm_ops);

CREATE INDEX ix_file_name_trgm
    ON file USING gin (lower(name) gin_trgm_ops);

CREATE INDEX ix_folder_name_trgm
    ON folder USING gin (lower(name) gin_trgm_ops);