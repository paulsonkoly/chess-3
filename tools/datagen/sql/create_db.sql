pragma foreign_keys = on;

create table games (
  id integer primary key,
  wdl integer
);

create table positions (
  id integer primary key,
  game_id integer not null,
  fen text not null,
  best_move  integer,
  eval integer,
  foreign key (game_id) references games(id)
);
