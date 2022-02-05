-- Create user database

create table users (
  id serial primary key,
  channel_id int not null,
  username varchar not null,
  points int default 0,
  last_activity timestamp default current_timestamp,
  constraint fk_channel foreign key(channel_id) references channels(id),
  constraint unique_user_in_channel unique(channel_id, username)
);

---- create above / drop below ----

drop table users;
