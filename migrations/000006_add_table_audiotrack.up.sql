create table if not exists audio_track
(
    id           text primary key,
    title        text                     not null,
    stream_index         integer,
    channels         integer,
    video_id         text,
    language text,
    codec text,
    is_default         bool,
    status         text,
    FOREIGN KEY (video_id) REFERENCES video (id)
        ON DELETE CASCADE ON UPDATE CASCADE

);
