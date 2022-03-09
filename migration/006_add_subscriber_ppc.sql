-- Add points per chat config

alter table channels
    add column subscriber_points_per_chat int default 0,
    add column vip_points_per_chat int default 0

---- create above / drop below ----

alter table channels 
    drop column subscriber_points_per_chat
    drop column vip_points_per_chat