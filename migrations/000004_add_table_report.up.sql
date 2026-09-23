create table if not exists report_video
(
    id              text primary key,
    reason          text not null,
    title           text not null,
    video_id        text,
    position_second BIGINT,
    comment         text,
    status          text,
    created_at timestamp with time zone,
    FOREIGN KEY (video_id) REFERENCES video (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);
