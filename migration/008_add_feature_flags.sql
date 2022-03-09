-- Create table for feature flags

create table feature_flags (
    id serial primary key,
    channel_id int not null,
    command varchar not null,
    flag_value boolean not null default TRUE,
    constraint fk_channel foreign key(channel_id) references channels(id),
    constraint unique_command_in_channel unique(channel_id, command)
);

---- create above / drop below ----

drop table feature_flags;