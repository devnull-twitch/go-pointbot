-- Add table to store current track

create table current_track (
    id serial primary key,
    channel_id int not null,
    track_name varchar not null,
    artist_name varchar not null,
    constraint fk_channel foreign key(channel_id) references channels(id),
    constraint unique_per_channel unique(channel_id)
);

---- create above / drop below ----

drop table current_track;