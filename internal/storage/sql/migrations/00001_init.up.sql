begin transaction;
-- Типы хранимых данных
create type data_type as enum ('auth', 'text', 'binary', 'card');
-- Пользователи
create table users(
    id uuid default gen_random_uuid(),
    login varchar(200) unique not null,
    pass varchar(64) not null,
    primary key (id)
);
-- Записи
create table records(
    id uuid default gen_random_uuid(),
    userid uuid not null,
    description varchar(300) not null,
    dtype data_type not null,
    created timestamp with time zone default current_timestamp,
    modified timestamp with time zone default current_timestamp,
    hashsum varchar(64) not null,
    primary key (id),
    foreign key (userid) references users (id)
);
-- Бинарные данные записей
create table datarecords(
    seq int generated always as identity,
    recordid uuid not null,
    data bytea not null,
    primary key (seq)
);
-- Метаинформация записей
create table metainfo(
    seq int generated always as identity,
    recordid uuid not null,
    key varchar not null,
    value varchar not null,
    primary key (seq),
    foreign key (recordid) references records (id)
);
commit;