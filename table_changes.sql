begin;

alter table payments add column refund_id varchar(255);
alter table payments add column refund_status varchar(50);
UPDATE TABLE payments SET status = 'SETTLED' WHERE status = 'CLEARED'; -- Check this app side also
-- Add webhook refund response

commit;
