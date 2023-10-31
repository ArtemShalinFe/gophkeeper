CREATE USER gopher
    PASSWORD 'gopher';

CREATE DATABASE gophkeeper
    OWNER 'gopher'
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';