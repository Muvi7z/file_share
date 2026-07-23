create table if not exists "user"
(
    id           text primary key,
    login        text                     not null,
    passwordHash integer                  not null,
    role         text,
    created_at   timestamp with time zone not null,
    updated_at   timestamp with time zone not null,

    unique (login)
);

create table if not exists user_session
(
    token      text primary key,
    login      text,
    role       text,
    expires_at timestamp not null
);

create table if not exists folder
(
    id                 text primary key,
    name               text,
    path               text,
    parent_id          text,
    root_folder_id     text,
    is_root            boolean,
    enabled            boolean,
    files_count        integer,
    video_count        integer,
    child_folder_count integer,
    last_scan_at       timestamp,

    FOREIGN KEY (parent_id) REFERENCES folder (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

create table if not exists video
(
    id               text primary key,
    title            text,
    folder_id        text,
    folder_name      text,
    parent_folder_id text,
    size             text,
    size_bytes       BIGINT,
    duration         text,
    modified_at      timestamp,
    codec            text,
    resolution       text,
    poster_url       text,
    stream_url       text,
    path             text,
    FOREIGN KEY (folder_id) REFERENCES folder (id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (parent_folder_id) REFERENCES folder (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

create table if not exists scan_jobs
(
    id                text primary key,
    folder_id         text,
    status            text,
    processed_videos  integer,
    processed_folders integer,
    started_at        timestamp,
    finished_at       timestamp,
    error             text
);

