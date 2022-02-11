-- Add trivia question table

create table trivia_questions (
    id serial primary key,
    channel_id int not null,
    username varchar not null,
    question varchar not null,
    correct_answer varchar not null,
    wrong_answer_1 varchar,
    wrong_answer_2 varchar,
    wrong_answer_3 varchar,
    constraint fk_channel foreign key(channel_id) references channels(id),
    constraint unique_question unique(question)
);

---- create above / drop below ----

drop table trivia_questions;