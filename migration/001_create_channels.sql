-- Create channel database

create table channels (
  id serial primary key,
  token varchar not null,
  channel_name varchar not null,
  constraint unique_name unique(channel_name)
);

---- create above / drop below ----

drop table channels;
