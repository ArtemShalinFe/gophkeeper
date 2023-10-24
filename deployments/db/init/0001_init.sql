CREATE USER gkeeper
    PASSWORD 'gkeeper';

CREATE DATABASE gophkeeper
    OWNER 'gkeeper'
    ENCODING 'UTF8'
    LC_COLLATE = 'en_US.utf8'
    LC_CTYPE = 'en_US.utf8';