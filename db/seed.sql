CREATE TABLE accounts(
   id serial PRIMARY KEY,
   name VARCHAR (50) UNIQUE NOT NULL
);

INSERT INTO accounts VALUES (1, 'Demo account 1');
INSERT INTO accounts VALUES (2, 'Demo account 2');
INSERT INTO accounts VALUES (3, 'Demo account 3');