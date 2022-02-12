-- Add active flag for trivia quiestions

alter table trivia_questions
    add column active int default 0

---- create above / drop below ----

alter table trivia_questions 
    drop column active