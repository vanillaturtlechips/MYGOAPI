CREATE TABLE students (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100) UNIQUE,
    borrowed_books TEXT[] -- Go의 []string과 매핑되는 Postgres 배열
);

INSERT INTO students (name, email, borrowed_books) 
VALUES ('이명일', 'myongil@test.com', '{"book-1", "book-2"}');