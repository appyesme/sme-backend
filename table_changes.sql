begin;

alter table appointments add column home_address text default null,

commit;
