-- Create table for reminder messages

create table reminder_messages (
    id serial primary key,
    channel_id int not null,
    reminder_message varchar not null,
    interval interval not null,
    trigger_once boolean not null,
    last_send timestamp,
    constraint fk_channel foreign key(channel_id) references channels(id)
);

---- create above / drop below ----

drop table reminder_messages;