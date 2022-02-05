-- Add points per chat config

alter table channels
    add column points_per_chat int default 0

---- create above / drop below ----

alter table channels 
    drop column points_per_chat